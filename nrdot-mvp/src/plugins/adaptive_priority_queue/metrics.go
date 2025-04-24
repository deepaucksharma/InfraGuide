package adaptivepriorityqueue

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

// metricsProcessor is the processor for applying priority queuing to metrics.
type metricsProcessor struct {
	logger       *zap.Logger
	config       *Config
	nextConsumer consumer.Metrics
	queue        *AdaptivePriorityQueue
	dlqExporter  OverflowHandler
}

// newMetricsProcessor creates a new metrics processor for priority queuing.
func newMetricsProcessor(
	ctx context.Context,
	logger *zap.Logger,
	config *Config,
	nextConsumer consumer.Metrics,
) (*metricsProcessor, error) {
	// Create the DLQ overflow handler
	dlqHandler := &metricsDLQHandler{
		logger: logger,
		// The actual DLQ exporter would be injected here
	}
	
	p := &metricsProcessor{
		logger:       logger,
		config:       config,
		nextConsumer: nextConsumer,
		dlqExporter:  dlqHandler,
	}
	
	// Create the priority queue
	p.queue = NewAdaptivePriorityQueue(logger, config, p.dlqExporter)
	
	// Start the worker to process queued items
	go p.worker(ctx)
	
	return p, nil
}

// ConsumeMetrics enqueues metrics to be processed based on priority.
func (p *metricsProcessor) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	// Determine the priority based on the metrics content
	priority := p.determinePriority(md)
	
	// Check if the circuit breaker is open
	if p.queue.IsCircuitOpen() {
		// Circuit is open, send directly to DLQ
		item := &QueueItem{
			Value:    md,
			Priority: priority,
			Added:    time.Now(),
		}
		return p.dlqExporter.HandleOverflow(ctx, item)
	}
	
	// Try to enqueue the metrics
	if !p.queue.Enqueue(ctx, md, priority) {
		// Failed to enqueue, already handled by overflow handler
		return nil
	}
	
	// Successfully enqueued
	return nil
}

// determinePriority determines the priority of the metrics.
func (p *metricsProcessor) determinePriority(md pmetric.Metrics) PriorityLevel {
	// Implementation placeholder
	// This would analyze the metrics to determine their priority
	// For example, based on resource attributes, metric names, or other criteria
	
	// Default implementation: assign normal priority
	return PriorityNormal
}

// worker processes items from the queue and forwards them to the next consumer.
func (p *metricsProcessor) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Dequeue the next item
			item := p.queue.Dequeue()
			if item == nil {
				// Queue is empty, wait a bit before trying again
				time.Sleep(10 * time.Millisecond)
				continue
			}
			
			// Process the item
			md := item.Value.(pmetric.Metrics)
			
			// Forward to the next consumer
			err := p.nextConsumer.ConsumeMetrics(ctx, md)
			if err != nil {
				p.logger.Error("Failed to process metrics", zap.Error(err))
				p.queue.RecordError()
			} else {
				p.queue.RecordSuccess()
			}
		}
	}
}

// Capabilities returns the capabilities of the processor.
func (p *metricsProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

// Shutdown stops the processor.
func (p *metricsProcessor) Shutdown(context.Context) error {
	// No cleanup needed
	return nil
}

// metricsDLQHandler handles metrics overflow by sending them to a DLQ.
type metricsDLQHandler struct {
	logger *zap.Logger
	// The actual DLQ exporter would be added here
}

// HandleOverflow implements the OverflowHandler interface.
func (h *metricsDLQHandler) HandleOverflow(ctx context.Context, item *QueueItem) error {
	// This would send the metrics to the DLQ
	// Implementation placeholder
	h.logger.Info("Sending metrics to DLQ",
		zap.String("priority", string(item.Priority)),
		zap.Time("added", item.Added),
	)
	
	return nil
}
