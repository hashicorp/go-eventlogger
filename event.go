package eventlogger

import (
	"time"
)

// EventType is a string that uniquely identifies the type of an Event within a
// given Broker.
type EventType string

// TODO needs an immutable type
type MetadataType map[string]interface{}

// TODO needs an immutable type
type PayloadType map[string]interface{}

// An Event is analogous to a log entry.
type Event struct {
	Type      EventType
	CreatedAt time.Time
	Metadata  MetadataType
	Payload   PayloadType
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
