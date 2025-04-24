package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Simple demonstration of the NRDOT+ MVP core concepts
// This standalone program simulates the three key features:
// 1. Dynamic cardinality control
// 2. Priority queuing with spilling to disk
// 3. Enhanced durability and resilience

// CardinalityLimiter simulates entropy-based cardinality control
type CardinalityLimiter struct {
	maxKeys        int
	keys           map[string]float64 // key -> entropy score
	droppedCount   int
	aggregatedCount int
	mutex          sync.Mutex
}

// APQueue simulates adaptive priority queue with WRR scheduling
type APQueue struct {
	priorities     map[string]int // priority level -> weight
	queue          map[string][]string // priority level -> items
	spilled        []string // spilled items
	currentRound   map[string]int // priority level -> used in current round
	mutex          sync.Mutex
}

// DLQ simulates enhanced DLQ with SHA-256 verification
type DLQ struct {
	storage        map[string]string // id -> data
	maxSize        int
	currentSize    int
	replayRate     int // items per second
	mutex          sync.Mutex
}

// CardinalityDemo demonstrates cardinality limiting
func CardinalityDemo() {
	fmt.Println("\n=== CardinalityLimiter Demo ===")
	
	// Create limiter with 100 max keys
	limiter := &CardinalityLimiter{
		maxKeys: 100,
		keys:    make(map[string]float64),
	}
	
	// Generate 500 keys with random entropy scores
	for i := 0; i < 500; i++ {
		key := fmt.Sprintf("key-%d", i)
		entropy := rand.Float64() // 0-1 random score
		
		// Process key
		limiter.ProcessKey(key, entropy)
		
		// Print progress every 100 keys
		if i > 0 && i%100 == 0 {
			fmt.Printf("Processed %d keys, current table size: %d, dropped: %d, aggregated: %d\n", 
				i, len(limiter.keys), limiter.droppedCount, limiter.aggregatedCount)
		}
	}
	
	fmt.Printf("\nFinal state: table size: %d, dropped: %d, aggregated: %d\n",
		len(limiter.keys), limiter.droppedCount, limiter.aggregatedCount)
}

// ProcessKey processes a key with its entropy score
func (cl *CardinalityLimiter) ProcessKey(key string, entropy float64) {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()
	
	// Check if key exists
	if _, exists := cl.keys[key]; exists {
		// Key exists, just update entropy
		cl.keys[key] = entropy
		return
	}
	
	// Key doesn't exist, check if table is full
	if len(cl.keys) >= cl.maxKeys {
		// Table is full, apply entropy-based policy
		if entropy < 0.75 {
			// Low entropy, drop
			cl.droppedCount++
			return
		} else if entropy < 0.9 {
			// Medium entropy, aggregate by dropping suffix
			newKey := key
			if len(key) > 5 {
				newKey = key[:5] + "*" // simple aggregation by truncation
			}
			cl.keys[newKey] = entropy
			cl.aggregatedCount++
			return
		} else {
			// High entropy, keep by removing lowest entropy key
			lowestKey := ""
			lowestEntropy := 1.1
			
			for k, e := range cl.keys {
				if e < lowestEntropy {
					lowestKey = k
					lowestEntropy = e
				}
			}
			
			if lowestEntropy < entropy {
				// Found a key with lower entropy, replace it
				delete(cl.keys, lowestKey)
				cl.keys[key] = entropy
				cl.droppedCount++ // Count the evicted key as dropped
			} else {
				// No key with lower entropy, drop this one
				cl.droppedCount++
			}
			return
		}
	}
	
	// Table has space, add the key
	cl.keys[key] = entropy
}

// APQDemo demonstrates adaptive priority queue
func APQDemo() {
	fmt.Println("\n=== Adaptive Priority Queue Demo ===")
	
	// Create queue with 5:3:1 weights
	queue := &APQueue{
		priorities: map[string]int{
			"critical": 5,
			"high":     3,
			"normal":   1,
		},
		queue: map[string][]string{
			"critical": {},
			"high":     {},
			"normal":   {},
		},
		currentRound: map[string]int{
			"critical": 0,
			"high":     0,
			"normal":   0,
		},
	}
	
	// Add items with different priorities
	// 20% critical, 30% high, 50% normal
	for i := 0; i < 100; i++ {
		item := fmt.Sprintf("item-%d", i)
		priority := "normal"
		
		roll := rand.Intn(100)
		if roll < 20 {
			priority = "critical"
		} else if roll < 50 {
			priority = "high"
		}
		
		queue.Enqueue(item, priority)
	}
	
	// Dequeue 50 items and count by priority
	counts := map[string]int{
		"critical": 0,
		"high":     0,
		"normal":   0,
	}
	
	for i := 0; i < 50; i++ {
		item, priority := queue.Dequeue()
		if item != "" {
			counts[priority]++
		}
	}
	
	fmt.Println("Dequeued 50 items with priorities:")
	fmt.Printf("Critical: %d (%.1f%%)\n", counts["critical"], float64(counts["critical"])/50*100)
	fmt.Printf("High:     %d (%.1f%%)\n", counts["high"], float64(counts["high"])/50*100)
	fmt.Printf("Normal:   %d (%.1f%%)\n", counts["normal"], float64(counts["normal"])/50*100)
	
	// Demonstrate spilling with a nearly full queue
	fmt.Println("\nSimulating queue pressure and spilling...")
	for i := 0; i < 950; i++ {
		item := fmt.Sprintf("pressure-item-%d", i)
		priority := "normal"
		queue.Enqueue(item, priority)
		
		// Every 100 items, show status
		if i > 0 && i%100 == 0 {
			c := queue.Count()
			s := queue.SpilledCount()
			fmt.Printf("Added %d items, queue size: %d, spilled: %d\n", i, c, s)
		}
	}
}

// Count returns the total number of items in the queue
func (q *APQueue) Count() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	total := 0
	for _, items := range q.queue {
		total += len(items)
	}
	return total
}

// SpilledCount returns the number of spilled items
func (q *APQueue) SpilledCount() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	return len(q.spilled)
}

// Enqueue adds an item to the priority queue
func (q *APQueue) Enqueue(item, priority string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	// Check if priority exists
	if _, exists := q.priorities[priority]; !exists {
		// Invalid priority, use normal
		priority = "normal"
	}
	
	// Check if queue is nearly full (800+ items)
	total := 0
	for _, items := range q.queue {
		total += len(items)
	}
	
	if total >= 800 && priority == "normal" {
		// Queue is nearly full, spill normal priority items
		q.spilled = append(q.spilled, item)
		return
	}
	
	// Add to appropriate queue
	q.queue[priority] = append(q.queue[priority], item)
}

// Dequeue removes an item using WRR scheduling
func (q *APQueue) Dequeue() (string, string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	// Check if all queues are empty
	empty := true
	for _, items := range q.queue {
		if len(items) > 0 {
			empty = false
			break
		}
	}
	
	if empty {
		return "", "" // No items
	}
	
	// Select priority using WRR
	var selectedPriority string
	priorities := []string{"critical", "high", "normal"}
	
	// First, check if any priority has used all its weights in this round
	allUsed := true
	for p, w := range q.priorities {
		if q.currentRound[p] < w {
			allUsed = false
			break
		}
	}
	
	// If all weights used, reset round
	if allUsed {
		for p := range q.currentRound {
			q.currentRound[p] = 0
		}
	}
	
	// Find highest priority with available weight and items
	for _, p := range priorities {
		if q.currentRound[p] < q.priorities[p] && len(q.queue[p]) > 0 {
			selectedPriority = p
			q.currentRound[p]++
			break
		}
	}
	
	// If no priority with weight found, use highest with items
	if selectedPriority == "" {
		for _, p := range priorities {
			if len(q.queue[p]) > 0 {
				selectedPriority = p
				break
			}
		}
	}
	
	// Get the first item from the selected queue
	item := q.queue[selectedPriority][0]
	q.queue[selectedPriority] = q.queue[selectedPriority][1:]
	
	return item, selectedPriority
}

// DLQDemo demonstrates enhanced DLQ
func DLQDemo() {
	fmt.Println("\n=== Enhanced DLQ Demo ===")
	
	// Create DLQ with 1000 max size
	dlq := &DLQ{
		storage:    make(map[string]string),
		maxSize:    1000,
		replayRate: 10, // 10 items per second
	}
	
	// Add 500 items
	for i := 0; i < 500; i++ {
		id := fmt.Sprintf("item-%d", i)
		data := fmt.Sprintf("data-content-%d", i)
		
		dlq.Write(id, data)
		
		// Print progress every 100 items
		if i > 0 && i%100 == 0 {
			fmt.Printf("Added %d items to DLQ, current size: %d\n", i, dlq.currentSize)
		}
	}
	
	// Simulate outage recovery with replay
	fmt.Println("\nSimulating outage recovery with replay...")
	
	// Count by 100s
	var wg sync.WaitGroup
	wg.Add(1)
	
	// Track replayed count
	replayed := 0
	var replayedMutex sync.Mutex
	
	go func() {
		defer wg.Done()
		dlq.Replay(func(id, data string) {
			replayedMutex.Lock()
			replayed++
			
			// Print progress every 100 items
			if replayed%100 == 0 {
				fmt.Printf("Replayed %d items from DLQ\n", replayed)
			}
			replayedMutex.Unlock()
		})
	}()
	
	// Wait for replay to complete or timeout after 10 seconds
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		fmt.Printf("\nReplay completed, replayed %d items\n", replayed)
	case <-time.After(10 * time.Second):
		fmt.Printf("\nReplay timeout after 10 seconds, replayed %d items\n", replayed)
	}
}

// Write adds an item to the DLQ
func (d *DLQ) Write(id, data string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	// Store the item
	d.storage[id] = data
	d.currentSize++
}

// Replay replays items from the DLQ at the configured rate
func (d *DLQ) Replay(processor func(id, data string)) {
	d.mutex.Lock()
	ids := make([]string, 0, len(d.storage))
	for id := range d.storage {
		ids = append(ids, id)
	}
	d.mutex.Unlock()
	
	// Replay items at the configured rate
	for _, id := range ids {
		d.mutex.Lock()
		data, exists := d.storage[id]
		d.mutex.Unlock()
		
		if exists {
			processor(id, data)
			
			// Sleep to control replay rate
			time.Sleep(time.Second / time.Duration(d.replayRate))
		}
	}
}

func main() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())
	
	fmt.Println("NRDOT+ MVP Standalone Demo")
	fmt.Println("==========================")
	fmt.Println("This program demonstrates the three key features of NRDOT+ MVP:")
	fmt.Println("1. Dynamic cardinality control")
	fmt.Println("2. Priority queuing with spilling to disk")
	fmt.Println("3. Enhanced durability and resilience")
	
	// Run the demos
	CardinalityDemo()
	APQDemo()
	DLQDemo()
	
	fmt.Println("\nDemo completed. Press Ctrl+C to exit.")
	
	// Wait for Ctrl+C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
