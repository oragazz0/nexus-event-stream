package domain_test

import (
	"testing"

	"github.com/oragazz0/nexus-event-stream/data-plane/internal/domain"
)

func TestParseSignalEvent_ValidCreated(t *testing.T) {
	payload := []byte(`{
		"action": "created",
		"id": "abc-123",
		"title": "Server Alert",
		"content": "CPU at 95%",
		"priority": "High",
		"author": "otavio",
		"created_at": "2026-02-23T15:00:00-03:00",
		"updated_at": "2026-02-23T15:00:00-03:00"
	}`)

	event, err := domain.ParseSignalEvent(payload)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Action != domain.ActionCreated {
		t.Errorf("expected action %q, got %q", domain.ActionCreated, event.Action)
	}
	if event.ID != "abc-123" {
		t.Errorf("expected id %q, got %q", "abc-123", event.ID)
	}
	if event.Priority != "High" {
		t.Errorf("expected priority %q, got %q", "High", event.Priority)
	}
}

func TestParseSignalEvent_DeletedMinimalPayload(t *testing.T) {
	payload := []byte(`{"action": "deleted", "id": "abc-123"}`)

	event, err := domain.ParseSignalEvent(payload)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Action != domain.ActionDeleted {
		t.Errorf("expected action %q, got %q", domain.ActionDeleted, event.Action)
	}
	if event.Title != "" {
		t.Errorf("expected empty title for delete event, got %q", event.Title)
	}
}

func TestParseSignalEvent_InvalidJSON(t *testing.T) {
	payload := []byte(`{not valid json}`)

	_, err := domain.ParseSignalEvent(payload)

	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParseSignalEvent_EmptyPayload(t *testing.T) {
	_, err := domain.ParseSignalEvent([]byte{})

	if err == nil {
		t.Fatal("expected error for empty payload, got nil")
	}
}

func TestParseSignalEvent_UnknownFieldsIgnored(t *testing.T) {
	payload := []byte(`{"action": "created", "id": "abc-123", "unknown_field": "value"}`)

	event, err := domain.ParseSignalEvent(payload)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.ID != "abc-123" {
		t.Errorf("expected id %q, got %q", "abc-123", event.ID)
	}
}

func TestSignalEventFields(t *testing.T) {
	event := domain.SignalEvent{
		ID:        "abc-123",
		Title:     "Test",
		Content:   "Body",
		Priority:  "Low",
		Author:    "otavio",
		CreatedAt: "2026-02-23T15:00:00-03:00",
		UpdatedAt: "2026-02-23T15:05:00-03:00",
	}

	fields := event.Fields()

	expectedKeys := []string{"id", "title", "content", "priority", "author", "created_at", "updated_at"}
	for _, key := range expectedKeys {
		if _, ok := fields[key]; !ok {
			t.Errorf("missing key %q in fields map", key)
		}
	}
	if fields["id"] != "abc-123" {
		t.Errorf("expected id %q, got %q", "abc-123", fields["id"])
	}
	if fields["priority"] != "Low" {
		t.Errorf("expected priority %q, got %q", "Low", fields["priority"])
	}
}

func TestSignalFromMap(t *testing.T) {
	data := map[string]string{
		"id":         "abc-123",
		"title":      "Alert",
		"content":    "Disk full",
		"priority":   "Medium",
		"author":     "otavio",
		"created_at": "2026-02-23T15:00:00-03:00",
		"updated_at": "2026-02-23T15:05:00-03:00",
	}

	signal := domain.SignalFromMap(data)

	if signal.ID != "abc-123" {
		t.Errorf("expected id %q, got %q", "abc-123", signal.ID)
	}
	if signal.Title != "Alert" {
		t.Errorf("expected title %q, got %q", "Alert", signal.Title)
	}
	if signal.Priority != "Medium" {
		t.Errorf("expected priority %q, got %q", "Medium", signal.Priority)
	}
}

func TestSignalFromMap_EmptyMap(t *testing.T) {
	signal := domain.SignalFromMap(map[string]string{})

	if signal.ID != "" {
		t.Errorf("expected empty id, got %q", signal.ID)
	}
	if signal.Title != "" {
		t.Errorf("expected empty title, got %q", signal.Title)
	}
}

func TestSignalEventFieldsRoundTrip(t *testing.T) {
	original := domain.SignalEvent{
		ID:        "uuid-456",
		Title:     "Round Trip",
		Content:   "Testing",
		Priority:  "High",
		Author:    "tester",
		CreatedAt: "2026-01-01T00:00:00Z",
		UpdatedAt: "2026-01-01T01:00:00Z",
	}

	fields := original.Fields()
	signal := domain.SignalFromMap(fields)

	if signal.ID != original.ID {
		t.Errorf("round trip failed for ID: %q != %q", signal.ID, original.ID)
	}
	if signal.Title != original.Title {
		t.Errorf("round trip failed for Title: %q != %q", signal.Title, original.Title)
	}
	if signal.Priority != original.Priority {
		t.Errorf("round trip failed for Priority: %q != %q", signal.Priority, original.Priority)
	}
	if signal.Author != original.Author {
		t.Errorf("round trip failed for Author: %q != %q", signal.Author, original.Author)
	}
}
