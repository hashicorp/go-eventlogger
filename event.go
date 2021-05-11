package eventlogger

import (
	"sync"
	"time"
)

// EventType is a string that uniquely identifies the type of an Event within a
// given Broker.
type EventType string

// An Event is analogous to a log entry.
type Event struct {
	// Type of Event
	Type EventType

	// CreatedAt defines the time the event was Sent
	CreatedAt time.Time

	l sync.RWMutex

	// Formatted used by Formatters to store formatted Event data which Sinks
	// can use when writing.  The keys correspond to different formats (json,
	// text, etc).
	Formatted map[string][]byte

	// Payload is the Event's payload data
	Payload interface{}
}
