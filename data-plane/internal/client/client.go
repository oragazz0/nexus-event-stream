package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/oragazz0/nexus-event-stream/data-plane/internal/domain"
)

// ErrNotFound is returned when the requested signal does not exist.
var ErrNotFound = errors.New("signal not found")

// DataPlane is an HTTP client for the data-plane read API.
type DataPlane struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a DataPlane client targeting the given base URL.
func New(baseURL string) DataPlane {
	return DataPlane{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// ListSignals returns all signals, optionally filtered by priority.
func (d DataPlane) ListSignals(priority string) ([]domain.Signal, error) {
	path := "/signals"
	if priority != "" {
		path = path + "?priority=" + priority
	}
	var signals []domain.Signal
	err := d.fetchJSON(path, &signals)
	return signals, err
}

// GetSignal returns a single signal by its ID.
func (d DataPlane) GetSignal(id string) (domain.Signal, error) {
	var signal domain.Signal
	err := d.fetchJSON("/signals/"+id, &signal)
	return signal, err
}

// Health checks the data-plane's health endpoint.
func (d DataPlane) Health() error {
	var result map[string]string
	return d.fetchJSON("/health", &result)
}

func (d DataPlane) fetchJSON(path string, target interface{}) error {
	response, err := d.get(path)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer response.Body.Close()
	return decodeResponse(response, target)
}

func (d DataPlane) get(path string) (*http.Response, error) {
	url := d.baseURL + path
	return d.httpClient.Get(url)
}

func decodeResponse(response *http.Response, target interface{}) error {
	if response.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", response.StatusCode)
	}
	decoder := json.NewDecoder(response.Body)
	return decoder.Decode(target)
}
