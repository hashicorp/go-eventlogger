package eventlogger

import (
	"time"

	iradix "github.com/hashicorp/go-immutable-radix"
)

// EventType is a string that uniquely identifies the type an Event within a given Broker.
type EventType string

// An Event is a collection of data, analogous to a log entry, that we want to
// process in a DeliveryTree.
type Event struct {
	Type         EventType
	CreatedAt    time.Time
	Metadata     *iradix.Tree
	UserMetadata *iradix.Tree
	Data         *iradix.Tree
}
