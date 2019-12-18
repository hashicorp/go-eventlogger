package eventlogger

import (
	"time"
	//iradix "github.com/hashicorp/go-immutable-radix"
)

// EventType is a string that uniquely identifies the type of an Envelope within a
// given Broker.
type EventType string

// An Envelope is analogous to a log entry.
type Envelope struct {
	Type         EventType
	CreatedAt    time.Time
	Metadata     map[string]interface{} // immutable
	UserMetadata map[string]interface{} // immutable

	//TODO: We haven't decided what this is going to be yet.
	// *iradix.Tree? something else?
	// Maybe we can allow for layering
	Data map[string]interface{}

	// Marshalled is a writable representation of the Envelope, e.g. a []byte.
	// Events that come in to the Broker should never have this field be
	// populated.  Instead, it should be populated by Nodes like ByteWriter as
	// the Envelope is propogated through its Graph.
	Marshalled interface{}
}

//func (e *Envelope) Clone() *Envelope {
//	return &Envelope{
//		Type:         e.Type,
//		CreatedAt:    e.CreatedAt,
//		Metadata:     e.Metadata,
//		UserMetadata: e.UserMetadata,
//		Data:         e.Data,
//		Writable:     e.Writable,
//	}
//}
