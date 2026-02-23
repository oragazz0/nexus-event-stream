package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/oragazz0/nexus-event-stream/data-plane/internal/domain"
	"github.com/oragazz0/nexus-event-stream/data-plane/internal/projection"
)

// SignalHandler serves the read API for the signals materialized view.
type SignalHandler struct {
	projection projection.SignalProjection
}

// New creates a SignalHandler.
func New(proj projection.SignalProjection) SignalHandler {
	return SignalHandler{projection: proj}
}

// Register mounts the handler routes on the given ServeMux.
func (h SignalHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /signals", h.listSignals)
	mux.HandleFunc("GET /signals/{id}", h.getSignal)
	mux.HandleFunc("GET /health", h.health)
}

func (h SignalHandler) listSignals(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	priority := query.Get("priority")
	signals, err := h.fetchSignals(request.Context(), priority)
	if err != nil {
		writeError(writer, http.StatusInternalServerError, "failed to list signals")
		return
	}
	writeJSON(writer, http.StatusOK, signals)
}

func (h SignalHandler) fetchSignals(ctx context.Context, priority string) ([]domain.Signal, error) {
	if priority != "" {
		return h.projection.ListByPriority(ctx, priority)
	}
	return h.projection.ListByCreatedAt(ctx, 0, 49)
}

func (h SignalHandler) getSignal(writer http.ResponseWriter, request *http.Request) {
	id := request.PathValue("id")
	signal, err := h.projection.FindByID(request.Context(), id)
	if errors.Is(err, projection.ErrNotFound) {
		writeError(writer, http.StatusNotFound, "signal not found")
		return
	}
	if err != nil {
		writeError(writer, http.StatusInternalServerError, "failed to get signal")
		return
	}
	writeJSON(writer, http.StatusOK, signal)
}

func (h SignalHandler) health(writer http.ResponseWriter, request *http.Request) {
	err := h.projection.Health(request.Context())
	if err != nil {
		writeError(writer, http.StatusServiceUnavailable, "redis unhealthy")
		return
	}
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(writer http.ResponseWriter, status int, data interface{}) {
	headers := writer.Header()
	headers.Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	encoder := json.NewEncoder(writer)
	encoder.Encode(data)
}

func writeError(writer http.ResponseWriter, status int, message string) {
	writeJSON(writer, status, map[string]string{"error": message})
}
