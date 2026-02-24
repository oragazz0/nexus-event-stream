package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/oragazz0/nexus-event-stream/data-plane/internal/domain"
	"github.com/oragazz0/nexus-event-stream/data-plane/internal/handler"
	"github.com/oragazz0/nexus-event-stream/data-plane/internal/projection"
	"github.com/redis/go-redis/v9"
)

func setupHandler(t *testing.T) (*http.ServeMux, projection.SignalProjection) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { client.Close() })

	proj := projection.New(client)
	signalHandler := handler.New(proj)
	mux := http.NewServeMux()
	signalHandler.Register(mux)

	return mux, proj
}

func seedSignal(t *testing.T, proj projection.SignalProjection, id, priority, createdAt string) {
	t.Helper()
	event := domain.SignalEvent{
		Action:    domain.ActionCreated,
		ID:        id,
		Title:     "Signal " + id,
		Content:   "Content for " + id,
		Priority:  priority,
		Author:    "otavio",
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}
	if err := proj.Apply(t.Context(), event); err != nil {
		t.Fatalf("failed to seed signal %s: %v", id, err)
	}
}

func TestListSignals_Empty(t *testing.T) {
	mux, _ := setupHandler(t)

	request := httptest.NewRequest(http.MethodGet, "/signals", nil)
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var signals []domain.Signal
	json.NewDecoder(recorder.Body).Decode(&signals)

	if len(signals) != 0 {
		t.Errorf("expected empty list, got %d signals", len(signals))
	}
}

func TestListSignals_ReturnsSeededSignals(t *testing.T) {
	mux, proj := setupHandler(t)
	seedSignal(t, proj, "s1", "High", "2026-02-23T15:00:00-03:00")
	seedSignal(t, proj, "s2", "Low", "2026-02-22T10:00:00-03:00")

	request := httptest.NewRequest(http.MethodGet, "/signals", nil)
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var signals []domain.Signal
	json.NewDecoder(recorder.Body).Decode(&signals)

	if len(signals) != 2 {
		t.Fatalf("expected 2 signals, got %d", len(signals))
	}
	if signals[0].ID != "s1" {
		t.Errorf("expected newest signal first (s1), got %q", signals[0].ID)
	}
}

func TestListSignals_FilterByPriority(t *testing.T) {
	mux, proj := setupHandler(t)
	seedSignal(t, proj, "high-1", "High", "2026-02-23T15:00:00-03:00")
	seedSignal(t, proj, "low-1", "Low", "2026-02-22T10:00:00-03:00")

	request := httptest.NewRequest(http.MethodGet, "/signals?priority=Low", nil)
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var signals []domain.Signal
	json.NewDecoder(recorder.Body).Decode(&signals)

	if len(signals) != 1 {
		t.Fatalf("expected 1 signal with Low priority, got %d", len(signals))
	}
	if signals[0].ID != "low-1" {
		t.Errorf("expected signal %q, got %q", "low-1", signals[0].ID)
	}
}

func TestGetSignal_Found(t *testing.T) {
	mux, proj := setupHandler(t)
	seedSignal(t, proj, "abc-123", "Medium", "2026-02-23T15:00:00-03:00")

	request := httptest.NewRequest(http.MethodGet, "/signals/abc-123", nil)
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var signal domain.Signal
	json.NewDecoder(recorder.Body).Decode(&signal)

	if signal.ID != "abc-123" {
		t.Errorf("expected id %q, got %q", "abc-123", signal.ID)
	}
	if signal.Priority != "Medium" {
		t.Errorf("expected priority %q, got %q", "Medium", signal.Priority)
	}
}

func TestGetSignal_NotFound(t *testing.T) {
	mux, _ := setupHandler(t)

	request := httptest.NewRequest(http.MethodGet, "/signals/nonexistent", nil)
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestHealth_OK(t *testing.T) {
	mux, _ := setupHandler(t)

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var body map[string]string
	json.NewDecoder(recorder.Body).Decode(&body)

	if body["status"] != "ok" {
		t.Errorf("expected status %q, got %q", "ok", body["status"])
	}
}

func TestListSignals_ContentTypeJSON(t *testing.T) {
	mux, _ := setupHandler(t)

	request := httptest.NewRequest(http.MethodGet, "/signals", nil)
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, request)

	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type %q, got %q", "application/json", contentType)
	}
}
