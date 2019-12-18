package eventlogger

import (
	"testing"
)

func TestBroker(t *testing.T) {
	b := &Broker{}
	b.Process(nil, "", nil)
}
