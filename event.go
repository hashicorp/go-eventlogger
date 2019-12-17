package eventlogger

import (
	"time"

	iradix "github.com/hashicorp/go-immutable-radix"
)

// EventType is a string that uniquely identifies the type of an Event within a
// given Broker.
type EventType string

// An Event is analogous to a log entry.
type Event struct {
	Type         EventType
	CreatedAt    time.Time
	Metadata     *iradix.Tree
	UserMetadata *iradix.Tree
	Data         *iradix.Tree

	Writable interface{} // A writable representation of the Event
}

func (e *Event) Clone() *Event {
	return &Event{
		Type:         e.Type,
		CreatedAt:    e.CreatedAt,
		Metadata:     e.Metadata,
		UserMetadata: e.UserMetadata,
		Data:         e.Data,
		Writable:     e.Writable,
	}
}
