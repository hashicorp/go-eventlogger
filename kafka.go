package eventlogger

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

type KafkaSink struct {
	// Format specifies the format the []byte representation is formatted in
	// Defaults to JSONFormat
	Format string

	// Topic specifies the topic the event should be written to.
	Topic string

	// Brokers is a list of Kafka brokers addresses.
	Brokers []string

	// The maximum permitted size of a message (defaults to 1000000). Should be
	// set equal to or smaller than the broker's `message.max.bytes`.
	MaxMessageBytes int

	// The following config options control how often messages are batched up and
	// sent to the broker. By default, messages are sent as fast as possible, and
	// all messages received while the current batch is in-flight are placed
	// into the subsequent batch.
	Flush struct {
		// The best-effort number of bytes needed to trigger a flush. Use the
		// global sarama.MaxRequestSize to set a hard upper limit.
		Bytes int
		// The best-effort number of messages needed to trigger a flush. Use
		// `MaxMessages` to set a hard upper limit.
		Messages int
		// The best-effort frequency of flushes. Equivalent to
		// `queue.buffering.max.ms` setting of JVM producer.
		Frequency time.Duration
		// The maximum number of messages the producer will send in a single
		// broker request. Defaults to 0 for unlimited. Similar to
		// `queue.buffering.max.messages` in the JVM producer.
		MaxMessages int
	}

	Retry struct {
		// The total number of times to retry sending a message (default 3).
		// Similar to the `message.send.max.retries` setting of the JVM producer.
		Max int
		// How long to wait for the cluster to settle between retries
		// (default 100ms). Similar to the `retry.backoff.ms` setting of the
		// JVM producer.
		Backoff time.Duration
	}

	producerOnce sync.Once

	producer sarama.SyncProducer

	lock sync.Mutex
}

// Type describes the type of the node as a Sink.
func (_ *KafkaSink) Type() NodeType {
	return NodeTypeSink
}

func (ks *KafkaSink) Process(_ context.Context, e *Event) (*Event, error) {

	var err error
	ks.producerOnce.Do(func() {
		err = ks.initProducer()
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	defer ks.producer.Close()

	if err := ks.emit(ks.producer, e); err != nil {
		return nil, err
	}

	// sink nodes are the end of the pipeline, no need to return the event
	return nil, nil
}

func (ks *KafkaSink) emit(p sarama.SyncProducer, e *Event) error {
	format := ks.Format
	if format == "" {
		format = JSONFormat
	}

	val, ok := e.Format(format)
	if !ok {
		return errors.New("event was not marshaled")
	}

	msg := &sarama.ProducerMessage{
		Topic: ks.Topic,
		Value: sarama.ByteEncoder(val),
	}

	if _, _, err := p.SendMessage(msg); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

func (ks *KafkaSink) Reopen() error {

	ks.lock.Lock()
	defer ks.lock.Unlock()

	ks.producer.Close()

	var err error
	ks.producerOnce = sync.Once{}
	ks.producerOnce.Do(func() {
		err = ks.initProducer()
	})
	if err != nil {
		return fmt.Errorf("failed to recreate producer: %w", err)
	}

	return nil
}

func (ks *KafkaSink) parseConfig() *sarama.Config {
	config := sarama.NewConfig()

	if ks.MaxMessageBytes > 0 {
		config.Producer.MaxMessageBytes = ks.MaxMessageBytes
	}

	if ks.Flush.Bytes > 0 {
		config.Producer.Flush.Bytes = ks.Flush.Bytes
	}

	if ks.Flush.Messages > 0 {
		config.Producer.Flush.Messages = ks.Flush.Messages
	}

	if ks.Flush.Frequency > 0 {
		config.Producer.Flush.Frequency = ks.Flush.Frequency
	}

	if ks.Flush.MaxMessages > 0 {
		config.Producer.Flush.MaxMessages = ks.Flush.MaxMessages
	}

	if ks.Retry.Max > 0 {
		config.Producer.Retry.Max = ks.Retry.Max
	}

	if ks.Retry.Backoff > 0 {
		config.Producer.Retry.Backoff = ks.Retry.Backoff
	}

	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	return config
}

func (ks *KafkaSink) initProducer() error {
	var err error

	c := ks.parseConfig()
	ks.producer, err = sarama.NewSyncProducer(ks.Brokers, c)

	return err
}
