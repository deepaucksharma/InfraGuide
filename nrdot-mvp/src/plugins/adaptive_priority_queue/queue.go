package adaptivepriorityqueue

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PriorityLevel represents a priority level in the queue.
type PriorityLevel string

const (
	PriorityCritical PriorityLevel = "critical"
	PriorityHigh     PriorityLevel = "high"
	PriorityNormal   PriorityLevel = "normal"
)

// QueueItem represents an item in the priority queue.
type QueueItem struct {
	Value    interface{}
	Priority PriorityLevel
	Index    int
	Added    time.Time
}

// AdaptivePriorityQueue implements a weighted round-robin priority queue.
type AdaptivePriorityQueue struct {
	logger            *zap.Logger
	config            *Config
	items             []*QueueItem
	lock              sync.RWMutex
	priorityWeights   map[PriorityLevel]int
	currentRound      int
	roundSelections   map[PriorityLevel]int
	circuitOpen       bool
	lastCircuitTrip   time.Time
	successCount      int64
	errorCount        int64
	circuitLock       sync.RWMutex
	overflowHandler   OverflowHandler
	overflowCount     int64
	processedCount    map[PriorityLevel]int64
	processedCountMux sync.Mutex
}

// OverflowHandler defines the interface for handling queue overflow.
type OverflowHandler interface {
	HandleOverflow(ctx context.Context, item *QueueItem) error
}

// NewAdaptivePriorityQueue creates a new adaptive priority queue.
func NewAdaptivePriorityQueue(logger *zap.Logger, config *Config, overflowHandler OverflowHandler) *AdaptivePriorityQueue {
	// Convert string map keys to PriorityLevel
	priorityWeights := make(map[PriorityLevel]int, len(config.Priorities))
	for k, v := range config.Priorities {
		priorityWeights[PriorityLevel(k)] = v
	}

	q := &AdaptivePriorityQueue{
		logger:          logger,
		config:          config,
		items:           make([]*QueueItem, 0, config.MaxQueueSize),
		priorityWeights: priorityWeights,
		roundSelections: make(map[PriorityLevel]int),
		overflowHandler: overflowHandler,
		processedCount:  make(map[PriorityLevel]int64),
	}

	// Initialize selection counters
	for priority := range priorityWeights {
		q.roundSelections[priority] = 0
	}

	return q
}

// Enqueue adds an item to the queue with the specified priority.
// Returns true if the item was added, false if it was rejected due to overflow.
func (q *AdaptivePriorityQueue) Enqueue(ctx context.Context, value interface{}, priority PriorityLevel) bool {
	q.lock.Lock()
	defer q.lock.Unlock()

	// Check if queue is full
	if len(q.items) >= int(float64(q.config.MaxQueueSize)*float64(q.config.QueueFullThreshold)/100.0) {
		// Queue is nearly full, apply overflow strategy
		item := &QueueItem{
			Value:    value,
			Priority: priority,
			Added:    time.Now(),
		}

		q.lock.Unlock() // Unlock before handling overflow
		err := q.overflowHandler.HandleOverflow(ctx, item)
		q.lock.Lock() // Lock again before returning

		if err != nil {
			q.logger.Error("Failed to handle queue overflow", zap.Error(err))
		}

		q.overflowCount++
		return false
	}

	// Add item to the queue
	item := &QueueItem{
		Value:    value,
		Priority: priority,
		Index:    len(q.items),
		Added:    time.Now(),
	}
	q.items = append(q.items, item)
	heap.Push(q, item)
	return true
}

// Dequeue removes and returns the next item from the queue based on WRR scheduling.
// Returns nil if the queue is empty.
func (q *AdaptivePriorityQueue) Dequeue() *QueueItem {
	q.lock.Lock()
	defer q.lock.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	// Determine which priority to dequeue based on WRR scheduling
	priority := q.selectNextPriority()

	// Find and remove the first item with the selected priority
	for i, item := range q.items {
		if item.Priority == priority {
			q.incrementProcessedCount(priority)
			return heap.Remove(q, i).(*QueueItem)
		}
	}

	// If no item with the selected priority is found, dequeue the highest priority item
	item := heap.Pop(q).(*QueueItem)
	q.incrementProcessedCount(item.Priority)
	return item
}

// selectNextPriority selects the next priority level based on WRR scheduling.
func (q *AdaptivePriorityQueue) selectNextPriority() PriorityLevel {
	// Reset round if all selections have been made
	allSelectionsUsed := true
	for priority, weight := range q.priorityWeights {
		if q.roundSelections[priority] < weight {
			allSelectionsUsed = false
			break
		}
	}

	if allSelectionsUsed {
		q.currentRound++
		for priority := range q.priorityWeights {
			q.roundSelections[priority] = 0
		}
	}

	// Select the highest priority level that hasn't used up its allocation
	var selectedPriority PriorityLevel
	priorityOrder := []PriorityLevel{PriorityCritical, PriorityHigh, PriorityNormal}

	for _, priority := range priorityOrder {
		weight := q.priorityWeights[priority]
		if weight > 0 && q.roundSelections[priority] < weight {
			selectedPriority = priority
			q.roundSelections[priority]++
			break
		}
	}

	// If no priority was selected (which shouldn't happen), use the highest priority
	if selectedPriority == "" {
		selectedPriority = PriorityCritical
		q.roundSelections[selectedPriority]++
	}

	return selectedPriority
}

// IsCircuitOpen returns whether the circuit breaker is open.
func (q *AdaptivePriorityQueue) IsCircuitOpen() bool {
	q.circuitLock.RLock()
	defer q.circuitLock.RUnlock()
	
	// Check if the circuit is open and if the reset timeout has passed
	if q.circuitOpen && time.Since(q.lastCircuitTrip) > time.Duration(q.config.CircuitBreakerResetTimeout)*time.Second {
		// Reset the circuit (will be done properly by RecordSuccess/RecordError)
		q.circuitLock.RUnlock()
		q.circuitLock.Lock()
		q.circuitOpen = false
		q.successCount = 0
		q.errorCount = 0
		q.circuitLock.Unlock()
		q.circuitLock.RLock()
	}
	
	return q.circuitOpen
}

// RecordSuccess records a successful operation for the circuit breaker.
func (q *AdaptivePriorityQueue) RecordSuccess() {
	if !q.config.CircuitBreakerEnabled {
		return
	}
	
	q.circuitLock.Lock()
	defer q.circuitLock.Unlock()
	
	q.successCount++
	
	// Reset the circuit if it was previously open
	if q.circuitOpen && time.Since(q.lastCircuitTrip) > time.Duration(q.config.CircuitBreakerResetTimeout)*time.Second {
		q.circuitOpen = false
		q.successCount = 1
		q.errorCount = 0
	}
}

// RecordError records an error for the circuit breaker.
func (q *AdaptivePriorityQueue) RecordError() {
	if !q.config.CircuitBreakerEnabled {
		return
	}
	
	q.circuitLock.Lock()
	defer q.circuitLock.Unlock()
	
	q.errorCount++
	
	// Check if we need to trip the circuit
	total := q.successCount + q.errorCount
	if total >= 10 { // Need a minimum number of requests before tripping
		errorPercentage := float64(q.errorCount) / float64(total) * 100.0
		if errorPercentage >= float64(q.config.CircuitBreakerErrorThreshold) {
			q.circuitOpen = true
			q.lastCircuitTrip = time.Now()
		}
	}
}

// Size returns the current number of items in the queue.
func (q *AdaptivePriorityQueue) Size() int {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return len(q.items)
}

// GetProcessedCount returns the number of items processed by priority.
func (q *AdaptivePriorityQueue) GetProcessedCount() map[PriorityLevel]int64 {
	q.processedCountMux.Lock()
	defer q.processedCountMux.Unlock()
	
	// Create a copy to avoid data races
	result := make(map[PriorityLevel]int64, len(q.processedCount))
	for k, v := range q.processedCount {
		result[k] = v
	}
	
	return result
}

// GetOverflowCount returns the number of items that couldn't be queued.
func (q *AdaptivePriorityQueue) GetOverflowCount() int64 {
	return q.overflowCount
}

// incrementProcessedCount increments the processed count for a priority.
func (q *AdaptivePriorityQueue) incrementProcessedCount(priority PriorityLevel) {
	q.processedCountMux.Lock()
	defer q.processedCountMux.Unlock()
	q.processedCount[priority]++
}

// heap.Interface implementation
func (q *AdaptivePriorityQueue) Len() int { return len(q.items) }

func (q *AdaptivePriorityQueue) Less(i, j int) bool {
	// Compare based on priority
	pi := q.items[i].Priority
	pj := q.items[j].Priority
	
	// Higher weight = higher priority
	return q.priorityWeights[pi] > q.priorityWeights[pj]
}

func (q *AdaptivePriorityQueue) Swap(i, j int) {
	q.items[i], q.items[j] = q.items[j], q.items[i]
	q.items[i].Index = i
	q.items[j].Index = j
}

func (q *AdaptivePriorityQueue) Push(x interface{}) {
	item := x.(*QueueItem)
	item.Index = len(q.items)
	q.items = append(q.items, item)
}

func (q *AdaptivePriorityQueue) Pop() interface{} {
	n := len(q.items)
	item := q.items[n-1]
	q.items[n-1] = nil // avoid memory leak
	q.items = q.items[0 : n-1]
	return item
}
