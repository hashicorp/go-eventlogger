package eventlogger

import (
	"testing"
	"time"
)

func TestEvent(t *testing.T) {
	e := &Envelope{}
	e.CreatedAt = time.Now()
}
