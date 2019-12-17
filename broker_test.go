package eventlogger

import (
	"testing"
)

func TestBroker(t *testing.T) {
	b := &Broker{}
	e := &Envelope{}
	b.Process(e)
}
