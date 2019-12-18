package eventlogger

import (
	"time"
)

// EventType is a string that uniquely identifies the type of an Event within a
// given Broker.
type EventType string

type MetadataType map[string]interface{}
type PayloadType map[string]interface{}

// An Event is analogous to a log entry.
type Event struct {
	Type      EventType
	CreatedAt time.Time
	Metadata  MetadataType
	Payload   PayloadType

	// Marshalled is a writable representation of the Event, e.g. a []byte.
	// Events that come in to the Broker should never have this field be
	// populated.  Instead, it should be populated by Nodes like ByteWriter as
	// the Event is propogated through its Graph.
	Marshalled interface{}
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
