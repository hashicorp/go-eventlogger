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
	Type      EventType
	CreatedAt time.Time
	l         sync.RWMutex
	Formatted map[string][]byte
	Payload   interface{}
}
