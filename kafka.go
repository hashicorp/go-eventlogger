package eventlogger

import (
	"context"
	"errors"
	"fmt"

	"github.com/IBM/sarama"
)

type KafkaSink struct {
	// Format specifies the format the []byte representation is formatted in
	// Defaults to JSONFormat
	Format string

	// Topic specifies the topic the event should be written to.
	Topic string

	// The Producer is used to publish events onto a Kafka stream.
	Producer sarama.SyncProducer
}

// Type describes the type of the node as a Sink.
func (_ *KafkaSink) Type() NodeType {
	return NodeTypeSink
}

func (ks *KafkaSink) Process(_ context.Context, e *Event) (*Event, error) {
	format := ks.Format
	if format == "" {
		format = JSONFormat
	}

	val, ok := e.Format(format)
	if !ok {
		return nil, errors.New("event was not marshaled")
	}

	msg := &sarama.ProducerMessage{
		Topic: ks.Topic,
		Value: sarama.ByteEncoder(val),
	}

	if _, _, err := ks.Producer.SendMessage(msg); err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return nil, nil
}
