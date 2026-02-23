package projection

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/oragazz0/nexus-event-stream/data-plane/internal/domain"
	"github.com/redis/go-redis/v9"
)

const (
	keyByCreatedAt = "signals:by_created_at"
	keyByPriority  = "signals:by_priority"
)

// ErrNotFound is returned when a signal does not exist in the projection.
var ErrNotFound = errors.New("signal not found")

var priorityScores = map[string]float64{
	"Low":    1,
	"Medium": 2,
	"High":   3,
}

// SignalProjection manages the Redis materialized view of signals.
type SignalProjection struct {
	client *redis.Client
}

// New creates a SignalProjection backed by the given Redis client.
func New(client *redis.Client) SignalProjection {
	return SignalProjection{client: client}
}

// Apply processes a signal event and updates the materialized view.
func (p SignalProjection) Apply(ctx context.Context, event domain.SignalEvent) error {
	if event.Action == domain.ActionDeleted {
		return p.evict(ctx, event.ID)
	}
	return p.upsert(ctx, event)
}

func (p SignalProjection) upsert(ctx context.Context, event domain.SignalEvent) error {
	pipe := p.client.TxPipeline()
	pipe.HSet(ctx, signalKey(event.ID), event.Fields())
	pipe.ZAdd(ctx, keyByCreatedAt, redis.Z{
		Score:  parseTimestamp(event.CreatedAt),
		Member: event.ID,
	})
	pipe.ZAdd(ctx, keyByPriority, redis.Z{
		Score:  priorityScores[event.Priority],
		Member: event.ID,
	})
	_, err := pipe.Exec(ctx)
	return err
}

func (p SignalProjection) evict(ctx context.Context, id string) error {
	pipe := p.client.TxPipeline()
	pipe.Del(ctx, signalKey(id))
	pipe.ZRem(ctx, keyByCreatedAt, id)
	pipe.ZRem(ctx, keyByPriority, id)
	_, err := pipe.Exec(ctx)
	return err
}

// ListByCreatedAt returns signals ordered by newest first.
func (p SignalProjection) ListByCreatedAt(ctx context.Context, start, stop int64) ([]domain.Signal, error) {
	ids, err := p.client.ZRevRange(ctx, keyByCreatedAt, start, stop).Result()
	if err != nil {
		return nil, err
	}
	return p.fetchMany(ctx, ids)
}

// ListByPriority returns signals filtered by priority level.
func (p SignalProjection) ListByPriority(ctx context.Context, priority string) ([]domain.Signal, error) {
	score := fmt.Sprintf("%g", priorityScores[priority])
	rangeBy := &redis.ZRangeBy{Min: score, Max: score}
	ids, err := p.client.ZRangeByScore(ctx, keyByPriority, rangeBy).Result()
	if err != nil {
		return nil, err
	}
	return p.fetchMany(ctx, ids)
}

// FindByID returns a single signal from the projection.
func (p SignalProjection) FindByID(ctx context.Context, id string) (domain.Signal, error) {
	data, err := p.client.HGetAll(ctx, signalKey(id)).Result()
	if err != nil {
		return domain.Signal{}, err
	}
	if len(data) == 0 {
		return domain.Signal{}, ErrNotFound
	}
	return domain.SignalFromMap(data), nil
}

// Health checks the Redis connection.
func (p SignalProjection) Health(ctx context.Context) error {
	return p.client.Ping(ctx).Err()
}

func (p SignalProjection) fetchMany(ctx context.Context, ids []string) ([]domain.Signal, error) {
	if len(ids) == 0 {
		return []domain.Signal{}, nil
	}
	pipe := p.client.Pipeline()
	commands := make([]*redis.MapStringStringCmd, len(ids))
	for index, id := range ids {
		commands[index] = pipe.HGetAll(ctx, signalKey(id))
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	return hydrateSignals(commands), nil
}

func hydrateSignals(commands []*redis.MapStringStringCmd) []domain.Signal {
	signals := make([]domain.Signal, 0, len(commands))
	for _, command := range commands {
		data := command.Val()
		if len(data) == 0 {
			continue
		}
		signals = append(signals, domain.SignalFromMap(data))
	}
	return signals
}

func signalKey(id string) string {
	return "signal:" + id
}

func parseTimestamp(value string) float64 {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return 0
	}
	return float64(parsed.Unix())
}
