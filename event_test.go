package eventlogger

import (
	"testing"
	"time"
)

func TestEvent(t *testing.T) {
	e := &Event{}
	e.CreatedAt = time.Now()
}
