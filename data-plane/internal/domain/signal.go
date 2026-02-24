package domain

import "encoding/json"

// Action represents the CRUD operation that triggered the event.
type Action string

const (
	ActionCreated Action = "created"
	ActionUpdated Action = "updated"
	ActionDeleted Action = "deleted"
)

// SignalEvent represents an event received from the nexus.signals topic.
type SignalEvent struct {
	Action    Action `json:"action"`
	ID        string `json:"id"`
	Title     string `json:"title,omitempty"`
	Content   string `json:"content,omitempty"`
	Priority  string `json:"priority,omitempty"`
	Author    string `json:"author,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// ParseSignalEvent deserializes a JSON payload into a SignalEvent.
func ParseSignalEvent(data []byte) (SignalEvent, error) {
	var event SignalEvent
	err := json.Unmarshal(data, &event)
	return event, err
}

// Fields returns the event data as a flat map for Redis hash storage.
func (e SignalEvent) Fields() map[string]string {
	return map[string]string{
		"id":         e.ID,
		"title":      e.Title,
		"content":    e.Content,
		"priority":   e.Priority,
		"author":     e.Author,
		"created_at": e.CreatedAt,
		"updated_at": e.UpdatedAt,
	}
}

// Signal is the read model served by the API.
type Signal struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Priority  string `json:"priority"`
	Author    string `json:"author"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// SignalFromMap builds a Signal from a Redis hash result.
func SignalFromMap(data map[string]string) Signal {
	return Signal{
		ID:        data["id"],
		Title:     data["title"],
		Content:   data["content"],
		Priority:  data["priority"],
		Author:    data["author"],
		CreatedAt: data["created_at"],
		UpdatedAt: data["updated_at"],
	}
}
