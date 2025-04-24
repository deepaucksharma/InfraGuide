package enhanceddlq

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DLQStorage manages the file-based DLQ storage operations.
type DLQStorage struct {
	config           *Config
	logger           *zap.Logger
	currentFile      *os.File
	currentFileSize  int64
	currentFilePath  string
	currentFileMutex sync.Mutex
	
	// Metrics
	totalWrittenBytes int64
	totalWrittenItems int64
	totalFiles        int64
	
	// Replay state
	replayActive     bool
	replayMutex      sync.Mutex
	rateLimiter      *RateLimiter
	replayInterleave *InterleaveController
}

// RateLimiter controls the replay rate to avoid overwhelming the system.
type RateLimiter struct {
	bytesPerSecond int64
	lastTime       time.Time
	bytesConsumed  int64
	mutex          sync.Mutex
}

// InterleaveController manages the interleaving of replay and live traffic.
type InterleaveController struct {
	ratio          int
	replayCounter  int
	liveCounter    int
	mutex          sync.Mutex
	replayAllowed  bool
	liveAllowed    bool
}

// NewDLQStorage creates a new DLQ storage manager.
func NewDLQStorage(config *Config, logger *zap.Logger) (*DLQStorage, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(config.Directory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create DLQ directory: %w", err)
	}
	
	// Create rate limiter
	rateLimiter := &RateLimiter{
		bytesPerSecond: int64(config.ReplayRateMiBSec * 1024 * 1024),
		lastTime:       time.Now(),
	}
	
	// Create interleave controller
	interleave := &InterleaveController{
		ratio:         config.InterleaveRatio,
		replayAllowed: true,
		liveAllowed:   true,
	}
	
	storage := &DLQStorage{
		config:           config,
		logger:           logger,
		rateLimiter:      rateLimiter,
		replayInterleave: interleave,
	}
	
	// Initialize the current file
	if err := storage.rotateFileIfNeeded(); err != nil {
		return nil, fmt.Errorf("failed to initialize DLQ file: %w", err)
	}
	
	// Start a background cleanup goroutine
	go storage.cleanupLoop(context.Background())
	
	return storage, nil
}

// rotateFileIfNeeded checks if a new file is needed and creates one if necessary.
func (s *DLQStorage) rotateFileIfNeeded() error {
	s.currentFileMutex.Lock()
	defer s.currentFileMutex.Unlock()
	
	// Check if we have a file and it's below the size limit
	if s.currentFile != nil && s.currentFileSize < int64(s.config.FileSizeLimitMiB)*1024*1024 {
		return nil
	}
	
	// Close the current file if it exists
	if s.currentFile != nil {
		if err := s.currentFile.Close(); err != nil {
			s.logger.Error("Failed to close current DLQ file", zap.Error(err))
		}
		s.currentFile = nil
	}
	
	// Create a new file
	timestamp := time.Now().UTC().Format("20060102-150405.000")
	filename := fmt.Sprintf("%s-%s.dlq", s.config.FilePrefix, timestamp)
	filepath := filepath.Join(s.config.Directory, filename)
	
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create new DLQ file: %w", err)
	}
	
	s.currentFile = file
	s.currentFilePath = filepath
	s.currentFileSize = 0
	s.totalFiles++
	
	s.logger.Info("Created new DLQ file", 
		zap.String("path", filepath),
		zap.Int64("totalFiles", s.totalFiles),
	)
	
	return nil
}

// Write writes data to the DLQ with SHA-256 verification.
func (s *DLQStorage) Write(ctx context.Context, data []byte) error {
	// Ensure we have a valid file to write to
	if err := s.rotateFileIfNeeded(); err != nil {
		return err
	}
	
	s.currentFileMutex.Lock()
	defer s.currentFileMutex.Unlock()
	
	// Calculate SHA-256 hash if enabled
	var hash string
	if s.config.VerifySHA256 {
		h := sha256.New()
		h.Write(data)
		hash = hex.EncodeToString(h.Sum(nil))
	}
	
	// Prepare the record header
	timestamp := time.Now().UTC().UnixNano()
	header := fmt.Sprintf("--- DLQ RECORD START %d ---\n", timestamp)
	footer := fmt.Sprintf("--- DLQ RECORD END %d", timestamp)
	
	if s.config.VerifySHA256 {
		footer += fmt.Sprintf(" SHA256:%s", hash)
	}
	footer += " ---\n"
	
	// Write the record
	if _, err := s.currentFile.WriteString(header); err != nil {
		return fmt.Errorf("failed to write DLQ record header: %w", err)
	}
	
	n, err := s.currentFile.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write DLQ data: %w", err)
	}
	
	if _, err := s.currentFile.WriteString("\n" + footer); err != nil {
		return fmt.Errorf("failed to write DLQ record footer: %w", err)
	}
	
	// Ensure data is synced to disk
	if err := s.currentFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync DLQ file to disk: %w", err)
	}
	
	// Update stats
	s.currentFileSize += int64(n + len(header) + len(footer) + 1) // +1 for newline
	s.totalWrittenBytes += int64(n)
	s.totalWrittenItems++
	
	return nil
}

// ListDLQFiles returns a list of all DLQ files in the storage directory.
func (s *DLQStorage) ListDLQFiles() ([]string, error) {
	// Get all files in the directory
	pattern := filepath.Join(s.config.Directory, fmt.Sprintf("%s-*.dlq", s.config.FilePrefix))
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list DLQ files: %w", err)
	}
	
	return files, nil
}

// StartReplay begins replaying data from the DLQ at the configured rate.
func (s *DLQStorage) StartReplay(ctx context.Context, consumer DLQConsumer) error {
	s.replayMutex.Lock()
	defer s.replayMutex.Unlock()
	
	if s.replayActive {
		return fmt.Errorf("replay is already active")
	}
	
	// List all DLQ files
	files, err := s.ListDLQFiles()
	if err != nil {
		return err
	}
	
	if len(files) == 0 {
		return nil // Nothing to replay
	}
	
	s.replayActive = true
	s.replayInterleave.Reset()
	s.rateLimiter.Reset()
	
	// Start replay in background
	go func() {
		s.logger.Info("Starting DLQ replay", 
			zap.Int("fileCount", len(files)),
			zap.Float64("rateMiBSec", s.config.ReplayRateMiBSec),
			zap.Int("interleaveRatio", s.config.InterleaveRatio),
		)
		
		// Create worker pool for replay
		var wg sync.WaitGroup
		recordCh := make(chan *DLQRecord, 1000)
		
		// Start worker goroutines
		for i := 0; i < s.config.ReplayConcurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for record := range recordCh {
					// Wait for rate limiter
					s.rateLimiter.Wait(len(record.Data))
					
					// Wait for interleave controller
					for !s.replayInterleave.AllowReplay() {
						time.Sleep(1 * time.Millisecond)
						
						// Check if context is cancelled
						select {
						case <-ctx.Done():
							return
						default:
						}
					}
					
					// Process the record
					if err := consumer.ConsumeDLQRecord(ctx, record); err != nil {
						s.logger.Error("Failed to consume DLQ record", 
							zap.Error(err),
							zap.Time("timestamp", record.Timestamp),
						)
					}
				}
			}()
		}
		
		// Read files and send records to workers
		for _, file := range files {
			if err := s.replayFile(ctx, file, recordCh); err != nil {
				s.logger.Error("Failed to replay DLQ file", 
					zap.Error(err),
					zap.String("file", file),
				)
			}
			
			// Check if context is cancelled
			select {
			case <-ctx.Done():
				close(recordCh)
				wg.Wait()
				s.markReplayCompleted()
				return
			default:
			}
		}
		
		close(recordCh)
		wg.Wait()
		s.markReplayCompleted()
		s.logger.Info("DLQ replay completed")
	}()
	
	return nil
}

// markReplayCompleted marks the replay as completed.
func (s *DLQStorage) markReplayCompleted() {
	s.replayMutex.Lock()
	defer s.replayMutex.Unlock()
	s.replayActive = false
}

// replayFile replays a single DLQ file, parsing records and sending them to the channel.
func (s *DLQStorage) replayFile(ctx context.Context, filePath string, recordCh chan<- *DLQRecord) error {
	// Implementation omitted for brevity
	// This would parse the file, extract records, verify SHA-256 if enabled,
	// and send each record to the recordCh channel
	
	return nil
}

// IsReplayActive returns whether a replay is currently active.
func (s *DLQStorage) IsReplayActive() bool {
	s.replayMutex.Lock()
	defer s.replayMutex.Unlock()
	return s.replayActive
}

// StopReplay stops an active replay operation.
func (s *DLQStorage) StopReplay() {
	s.replayMutex.Lock()
	defer s.replayMutex.Unlock()
	s.replayActive = false
}

// Shutdown closes the DLQ storage.
func (s *DLQStorage) Shutdown() error {
	s.currentFileMutex.Lock()
	defer s.currentFileMutex.Unlock()
	
	if s.currentFile != nil {
		if err := s.currentFile.Close(); err != nil {
			return fmt.Errorf("failed to close DLQ file: %w", err)
		}
		s.currentFile = nil
	}
	
	return nil
}

// cleanupLoop periodically cleans up old DLQ files based on retention policy.
func (s *DLQStorage) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.cleanupOldFiles(); err != nil {
				s.logger.Error("Failed to clean up old DLQ files", zap.Error(err))
			}
		}
	}
}

// cleanupOldFiles removes DLQ files that exceed the retention period.
func (s *DLQStorage) cleanupOldFiles() error {
	// Get all DLQ files
	files, err := s.ListDLQFiles()
	if err != nil {
		return err
	}
	
	// Calculate cutoff time
	cutoff := time.Now().Add(-time.Duration(s.config.RetentionHours) * time.Hour)
	
	for _, file := range files {
		// Get file info
		info, err := os.Stat(file)
		if err != nil {
			s.logger.Warn("Failed to get file info during cleanup", 
				zap.Error(err),
				zap.String("file", file),
			)
			continue
		}
		
		// Check if file is older than retention period
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(file); err != nil {
				s.logger.Warn("Failed to delete old DLQ file", 
					zap.Error(err),
					zap.String("file", file),
				)
				continue
			}
			
			s.logger.Info("Deleted old DLQ file", 
				zap.String("file", file),
				zap.Time("modTime", info.ModTime()),
				zap.Time("cutoff", cutoff),
			)
		}
	}
	
	return nil
}

// DLQRecord represents a record stored in the DLQ.
type DLQRecord struct {
	Timestamp time.Time
	Data      []byte
	Hash      string
}

// DLQConsumer interface for consuming DLQ records.
type DLQConsumer interface {
	ConsumeDLQRecord(ctx context.Context, record *DLQRecord) error
}

// Reset resets the rate limiter.
func (r *RateLimiter) Reset() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.lastTime = time.Now()
	r.bytesConsumed = 0
}

// Wait waits until the rate limit allows processing the specified number of bytes.
func (r *RateLimiter) Wait(bytes int) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	// Calculate how long we should wait
	r.bytesConsumed += int64(bytes)
	expectedDuration := time.Duration(float64(r.bytesConsumed) / float64(r.bytesPerSecond) * float64(time.Second))
	elapsedTime := time.Since(r.lastTime)
	
	if expectedDuration > elapsedTime {
		// Need to wait
		time.Sleep(expectedDuration - elapsedTime)
	}
	
	// If too much time has passed, reset the counters
	if elapsedTime > time.Second*2 {
		r.lastTime = time.Now()
		r.bytesConsumed = int64(bytes)
	}
}

// Reset resets the interleave controller.
func (i *InterleaveController) Reset() {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.replayCounter = 0
	i.liveCounter = 0
	i.replayAllowed = true
	i.liveAllowed = true
}

// AllowReplay returns whether replay processing is allowed at this time.
func (i *InterleaveController) AllowReplay() bool {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	
	// Check if replay is allowed
	if !i.replayAllowed {
		// Need to wait for live traffic
		return false
	}
	
	// Increment replay counter
	i.replayCounter++
	
	// Check if we need to switch to live traffic
	if i.replayCounter >= i.ratio {
		i.replayAllowed = false
		i.liveAllowed = true
		i.replayCounter = 0
	}
	
	return true
}

// AllowLive returns whether live traffic processing is allowed at this time.
func (i *InterleaveController) AllowLive() bool {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	
	// Check if live traffic is allowed
	if !i.liveAllowed {
		// Need to wait for replay
		return false
	}
	
	// Increment live counter
	i.liveCounter++
	
	// Check if we need to switch to replay
	if i.liveCounter >= i.ratio {
		i.liveAllowed = false
		i.replayAllowed = true
		i.liveCounter = 0
	}
	
	return true
}
