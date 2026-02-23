package consumer

import (
	"context"
	"log"
	"time"

	"github.com/oragazz0/nexus-event-stream/data-plane/internal/domain"
	"github.com/oragazz0/nexus-event-stream/data-plane/internal/projection"
	"github.com/segmentio/kafka-go"
)

// Consumer reads events from Kafka and applies them to the projection.
type Consumer struct {
	reader     *kafka.Reader
	projection projection.SignalProjection
}

// New creates a Consumer.
func New(reader *kafka.Reader, proj projection.SignalProjection) Consumer {
	return Consumer{reader: reader, projection: proj}
}

// Start begins the consume loop. Blocks until the context is cancelled.
func (c Consumer) Start(ctx context.Context) error {
	for ctx.Err() == nil {
		c.processNext(ctx)
	}
	return ctx.Err()
}

func (c Consumer) processNext(ctx context.Context) {
	message, err := c.reader.FetchMessage(ctx)
	if err != nil {
		log.Printf("error fetching message: %v", err)
		return
	}

	event, err := domain.ParseSignalEvent(message.Value)
	if err != nil {
		log.Printf("skipping malformed message at offset %d: %v", message.Offset, err)
		c.commit(ctx, message)
		return
	}

	if !c.applyWithRetry(ctx, event) {
		return
	}

	c.commit(ctx, message)
	log.Printf("projected signal %s [%s]", event.ID, event.Action)
}

// applyWithRetry retries the projection until success or context cancellation.
// Returns true on success, false on context cancellation.
func (c Consumer) applyWithRetry(ctx context.Context, event domain.SignalEvent) bool {
	for {
		err := c.projection.Apply(ctx, event)
		if err == nil {
			return true
		}
		log.Printf("projection failed, retrying in 1s: %v", err)
		if !wait(ctx, time.Second) {
			return false
		}
	}
}

func (c Consumer) commit(ctx context.Context, message kafka.Message) {
	if err := c.reader.CommitMessages(ctx, message); err != nil {
		log.Printf("offset commit failed: %v", err)
	}
}

func wait(ctx context.Context, duration time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(duration):
		return true
	}
}
