package client_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oragazz0/nexus-event-stream/data-plane/internal/client"
	"github.com/oragazz0/nexus-event-stream/data-plane/internal/domain"
)

func fakeServer(handler http.HandlerFunc) (*httptest.Server, client.DataPlane) {
	server := httptest.NewServer(handler)
	dataPlane := client.New(server.URL)
	return server, dataPlane
}

func respondJSON(writer http.ResponseWriter, status int, body interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	json.NewEncoder(writer).Encode(body)
}

func TestListSignals_ReturnsSignals(t *testing.T) {
	signals := []domain.Signal{
		{ID: "s1", Title: "Alert", Priority: "High", Author: "otavio"},
		{ID: "s2", Title: "Info", Priority: "Low", Author: "otavio"},
	}
	server, dataPlane := fakeServer(func(writer http.ResponseWriter, request *http.Request) {
		respondJSON(writer, http.StatusOK, signals)
	})
	defer server.Close()

	result, err := dataPlane.ListSignals("")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 signals, got %d", len(result))
	}
	if result[0].ID != "s1" {
		t.Errorf("expected first signal ID %q, got %q", "s1", result[0].ID)
	}
}

func TestListSignals_SendsPriorityQuery(t *testing.T) {
	server, dataPlane := fakeServer(func(writer http.ResponseWriter, request *http.Request) {
		query := request.URL.Query()
		priority := query.Get("priority")
		if priority != "High" {
			t.Errorf("expected priority query %q, got %q", "High", priority)
		}
		respondJSON(writer, http.StatusOK, []domain.Signal{})
	})
	defer server.Close()

	dataPlane.ListSignals("High")
}

func TestListSignals_EmptyList(t *testing.T) {
	server, dataPlane := fakeServer(func(writer http.ResponseWriter, request *http.Request) {
		respondJSON(writer, http.StatusOK, []domain.Signal{})
	})
	defer server.Close()

	result, err := dataPlane.ListSignals("")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty list, got %d signals", len(result))
	}
}

func TestGetSignal_Found(t *testing.T) {
	expected := domain.Signal{ID: "abc-123", Title: "Alert", Priority: "High"}
	server, dataPlane := fakeServer(func(writer http.ResponseWriter, request *http.Request) {
		respondJSON(writer, http.StatusOK, expected)
	})
	defer server.Close()

	result, err := dataPlane.GetSignal("abc-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "abc-123" {
		t.Errorf("expected ID %q, got %q", "abc-123", result.ID)
	}
	if result.Priority != "High" {
		t.Errorf("expected priority %q, got %q", "High", result.Priority)
	}
}

func TestGetSignal_NotFound(t *testing.T) {
	server, dataPlane := fakeServer(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	_, err := dataPlane.GetSignal("nonexistent")

	if !errors.Is(err, client.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetSignal_ServerError(t *testing.T) {
	server, dataPlane := fakeServer(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	})
	defer server.Close()

	_, err := dataPlane.GetSignal("abc-123")

	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestHealth_Healthy(t *testing.T) {
	server, dataPlane := fakeServer(func(writer http.ResponseWriter, request *http.Request) {
		respondJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
	})
	defer server.Close()

	err := dataPlane.Health()

	if err != nil {
		t.Fatalf("expected healthy, got error: %v", err)
	}
}

func TestHealth_Unhealthy(t *testing.T) {
	server, dataPlane := fakeServer(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusServiceUnavailable)
	})
	defer server.Close()

	err := dataPlane.Health()

	if err == nil {
		t.Fatal("expected error for unhealthy response, got nil")
	}
}

func TestConnectionRefused(t *testing.T) {
	dataPlane := client.New("http://localhost:1")

	_, err := dataPlane.ListSignals("")

	if err == nil {
		t.Fatal("expected connection error, got nil")
	}
}

func TestGetSignal_RequestsCorrectPath(t *testing.T) {
	server, dataPlane := fakeServer(func(writer http.ResponseWriter, request *http.Request) {
		expectedPath := "/signals/uuid-456"
		if request.URL.Path != expectedPath {
			t.Errorf("expected path %q, got %q", expectedPath, request.URL.Path)
		}
		respondJSON(writer, http.StatusOK, domain.Signal{ID: "uuid-456"})
	})
	defer server.Close()

	dataPlane.GetSignal("uuid-456")
}
