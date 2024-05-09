package eventlogger

import (
	"errors"
	"testing"
	"time"

	"github.com/IBM/sarama"
	saramamocks "github.com/IBM/sarama/mocks"
	"github.com/stretchr/testify/require"
)

func TestKafkaSink_emit(t *testing.T) {
	sink := KafkaSink{
		Format: JSONFormat,
		Topic:  "test-topic",
		// Brokers: []string{brokers[0]},
	}

	evt := &Event{
		Type:      "test",
		CreatedAt: time.Now(),
		Formatted: map[string][]byte{
			JSONFormat: []byte(`"hello world"`),
		},
		Payload: []byte("hello world!"),
	}

	config := sarama.NewConfig()
	config.Version = sarama.DefaultVersion
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	p := saramamocks.NewSyncProducer(t, config)

	p.ExpectSendMessageWithCheckerFunctionAndSucceed(func(val []byte) error {
		if string(val) != `"hello world"` {
			t.Fatalf("unexpected value: %s", val)
		}
		return nil
	})

	err := sink.emit(p, evt)
	if err != nil {
		t.Fatalf("failed to process event: %v", err)
	}

	internalErr := errors.New("internal error")
	p.ExpectSendMessageWithCheckerFunctionAndFail(func(val []byte) error {
		if string(val) != `"hello world"` {
			t.Fatalf("unexpected value: %s", val)
		}
		return nil
	}, internalErr)

	err = sink.emit(p, evt)
	require.ErrorIs(t, err, internalErr)
}
