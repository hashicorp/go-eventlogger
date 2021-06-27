package cloudevents

import (
	"testing"

	"github.com/hashicorp/eventlogger"
	"github.com/stretchr/testify/assert"
)

func TestFormatter_Type(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	f := Formatter{}
	assert.Equal(eventlogger.NodeTypeFormatter, f.Type())
}

func TestFormatter_Name(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	f := Formatter{}
	assert.Equal(NodeName, f.Name())
}

func TestFormatter_Reopen(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	f := Formatter{}
	assert.NoError(f.Reopen())
}
