package eventlogger

import (
	"context"
	"errors"
	"fmt"

	"github.com/IBM/sarama"
)

type KafkaSink struct {
	// Brokers is a list of Kafka brokers to connect to
	Broker []string

	// Config contains configuration values for the Kafka producer.
	Config *sarama.Config

	// Format specifies the format the []byte representation is formatted in
	// Defaults to JSONFormat
	Format string

	// Topic specifies the topic the message should be written to.
	Topic string

	producer sarama.SyncProducer
}

// Type describes the type of the node as a Sink.
func (_ *KafkaSink) Type() NodeType {
	return NodeTypeSink
}

func (ks *KafkaSink) Process(ctx context.Context, e *Event) (*Event, error) {
	if ks.producer == nil {
		var err error
		ks.producer, err = sarama.NewSyncProducer(ks.Broker, ks.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to create kafka producer: %w", err)
		}
	}

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

	if _, _, err := ks.producer.SendMessage(msg); err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return nil, nil
}
