package eventlogger

import (
	"time"
)

// EventType is a string that uniquely identifies the type of an Event within a
// given Broker.
type EventType string

// TODO needs an immutable type
type Metadata map[string]interface{}

// TODO needs an immutable type
type Payload map[string]interface{}

// An Event is analogous to a log entry.
type Event struct {
	Type      EventType
	CreatedAt time.Time
	Metadata  Metadata
	Payload   Payload
}

//func (e *Event) Clone() *Event {
//	return &Event{
//		Type:         e.Type,
//		CreatedAt:    e.CreatedAt,
//		Metadata:     e.Metadata,
//		UserMetadata: e.UserMetadata,
//		Payload:      e.Payload,
//		Writable:     e.Writable,
//	}
//}
