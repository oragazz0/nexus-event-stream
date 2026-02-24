package projection_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/oragazz0/nexus-event-stream/data-plane/internal/domain"
	"github.com/oragazz0/nexus-event-stream/data-plane/internal/projection"
	"github.com/redis/go-redis/v9"
)

func setupProjection(t *testing.T) (projection.SignalProjection, *miniredis.Miniredis) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { client.Close() })
	return projection.New(client), server
}

func sampleEvent(action domain.Action, id string) domain.SignalEvent {
	return domain.SignalEvent{
		Action:    action,
		ID:        id,
		Title:     "Server Alert",
		Content:   "CPU at 95%",
		Priority:  "High",
		Author:    "otavio",
		CreatedAt: "2026-02-23T15:00:00-03:00",
		UpdatedAt: "2026-02-23T15:05:00-03:00",
	}
}

func TestApply_Created(t *testing.T) {
	proj, _ := setupProjection(t)
	ctx := context.Background()
	event := sampleEvent(domain.ActionCreated, "signal-1")

	err := proj.Apply(ctx, event)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	signal, err := proj.FindByID(ctx, "signal-1")
	if err != nil {
		t.Fatalf("unexpected error finding signal: %v", err)
	}
	if signal.Title != "Server Alert" {
		t.Errorf("expected title %q, got %q", "Server Alert", signal.Title)
	}
	if signal.Priority != "High" {
		t.Errorf("expected priority %q, got %q", "High", signal.Priority)
	}
}

func TestApply_Updated(t *testing.T) {
	proj, _ := setupProjection(t)
	ctx := context.Background()

	original := sampleEvent(domain.ActionCreated, "signal-1")
	proj.Apply(ctx, original)

	updated := sampleEvent(domain.ActionUpdated, "signal-1")
	updated.Title = "Updated Alert"
	updated.Priority = "Low"

	err := proj.Apply(ctx, updated)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	signal, err := proj.FindByID(ctx, "signal-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if signal.Title != "Updated Alert" {
		t.Errorf("expected title %q, got %q", "Updated Alert", signal.Title)
	}
	if signal.Priority != "Low" {
		t.Errorf("expected priority %q, got %q", "Low", signal.Priority)
	}
}

func TestApply_Deleted(t *testing.T) {
	proj, _ := setupProjection(t)
	ctx := context.Background()

	proj.Apply(ctx, sampleEvent(domain.ActionCreated, "signal-1"))

	deleteEvent := domain.SignalEvent{
		Action: domain.ActionDeleted,
		ID:     "signal-1",
	}
	err := proj.Apply(ctx, deleteEvent)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = proj.FindByID(ctx, "signal-1")
	if err != projection.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestApply_DeleteNonExistent(t *testing.T) {
	proj, _ := setupProjection(t)
	ctx := context.Background()

	deleteEvent := domain.SignalEvent{
		Action: domain.ActionDeleted,
		ID:     "does-not-exist",
	}

	err := proj.Apply(ctx, deleteEvent)

	if err != nil {
		t.Fatalf("deleting non-existent signal should not error, got: %v", err)
	}
}

func TestApply_Idempotent(t *testing.T) {
	proj, _ := setupProjection(t)
	ctx := context.Background()
	event := sampleEvent(domain.ActionCreated, "signal-1")

	proj.Apply(ctx, event)
	proj.Apply(ctx, event)

	signals, err := proj.ListByCreatedAt(ctx, 0, 49)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(signals) != 1 {
		t.Errorf("expected 1 signal after duplicate apply, got %d", len(signals))
	}
}

func TestFindByID_NotFound(t *testing.T) {
	proj, _ := setupProjection(t)
	ctx := context.Background()

	_, err := proj.FindByID(ctx, "nonexistent")

	if err != projection.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListByCreatedAt_Empty(t *testing.T) {
	proj, _ := setupProjection(t)
	ctx := context.Background()

	signals, err := proj.ListByCreatedAt(ctx, 0, 49)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(signals) != 0 {
		t.Errorf("expected empty list, got %d signals", len(signals))
	}
}

func TestListByCreatedAt_Order(t *testing.T) {
	proj, _ := setupProjection(t)
	ctx := context.Background()

	older := sampleEvent(domain.ActionCreated, "older")
	older.CreatedAt = "2026-02-22T10:00:00-03:00"

	newer := sampleEvent(domain.ActionCreated, "newer")
	newer.CreatedAt = "2026-02-23T10:00:00-03:00"

	proj.Apply(ctx, older)
	proj.Apply(ctx, newer)

	signals, err := proj.ListByCreatedAt(ctx, 0, 49)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(signals) != 2 {
		t.Fatalf("expected 2 signals, got %d", len(signals))
	}
	if signals[0].ID != "newer" {
		t.Errorf("expected newest first, got %q", signals[0].ID)
	}
	if signals[1].ID != "older" {
		t.Errorf("expected oldest second, got %q", signals[1].ID)
	}
}

func TestListByPriority_Filter(t *testing.T) {
	proj, _ := setupProjection(t)
	ctx := context.Background()

	high := sampleEvent(domain.ActionCreated, "high-1")
	high.Priority = "High"

	low := sampleEvent(domain.ActionCreated, "low-1")
	low.Priority = "Low"
	low.CreatedAt = "2026-02-22T10:00:00-03:00"

	proj.Apply(ctx, high)
	proj.Apply(ctx, low)

	signals, err := proj.ListByPriority(ctx, "High")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(signals) != 1 {
		t.Fatalf("expected 1 high-priority signal, got %d", len(signals))
	}
	if signals[0].ID != "high-1" {
		t.Errorf("expected signal %q, got %q", "high-1", signals[0].ID)
	}
}

func TestListByPriority_NoMatch(t *testing.T) {
	proj, _ := setupProjection(t)
	ctx := context.Background()

	low := sampleEvent(domain.ActionCreated, "low-1")
	low.Priority = "Low"
	proj.Apply(ctx, low)

	signals, err := proj.ListByPriority(ctx, "High")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(signals) != 0 {
		t.Errorf("expected no signals, got %d", len(signals))
	}
}

func TestListByCreatedAt_RemovedAfterDelete(t *testing.T) {
	proj, _ := setupProjection(t)
	ctx := context.Background()

	proj.Apply(ctx, sampleEvent(domain.ActionCreated, "signal-1"))
	proj.Apply(ctx, domain.SignalEvent{Action: domain.ActionDeleted, ID: "signal-1"})

	signals, err := proj.ListByCreatedAt(ctx, 0, 49)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(signals) != 0 {
		t.Errorf("expected empty list after delete, got %d", len(signals))
	}
}

func TestHealth_Healthy(t *testing.T) {
	proj, _ := setupProjection(t)
	ctx := context.Background()

	err := proj.Health(ctx)

	if err != nil {
		t.Fatalf("expected healthy, got error: %v", err)
	}
}

func TestHealth_Unhealthy(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	proj := projection.New(client)
	client.Close()

	err := proj.Health(context.Background())

	if err == nil {
		t.Fatal("expected error for closed connection, got nil")
	}
}
