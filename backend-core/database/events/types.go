package events

import (
	"context"
	"time"
)

// EventStoreRepository handles event sourcing
type EventStoreRepository interface {
	AppendEvent(ctx context.Context, streamID string, event Event) error
	GetEvents(ctx context.Context, streamID string, fromVersion int) ([]Event, error)
	GetSnapshot(ctx context.Context, streamID string) (*Snapshot, error)
	SaveSnapshot(ctx context.Context, streamID string, snapshot *Snapshot) error
}

// Event represents a domain event
type Event struct {
	ID        string                 `json:"id"`
	StreamID  string                 `json:"stream_id"`
	Version   int                    `json:"version"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// Snapshot represents a snapshot of an aggregate
type Snapshot struct {
	StreamID  string                 `json:"stream_id"`
	Version   int                    `json:"version"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}
