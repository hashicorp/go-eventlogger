package eventlogger

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
)

func TestKafkaSink_Process(t *testing.T) {
	ctx := context.Background()

	kafkaContainer, err := kafka.RunContainer(ctx,
		kafka.WithClusterID("test-cluster"),
		testcontainers.WithImage("confluentinc/confluent-local:7.5.0"),
	)
	if err != nil {
		t.Fatalf("failed to start kafka container: %v", err)
	}
	t.Cleanup(func() {
		if err := kafkaContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %v", err)
		}
	})

	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		t.Fatalf("failed to query brokers: %v", err)
	}

	sink := KafkaSink{
		Format:  JSONFormat,
		Topic:   "test-topic",
		Brokers: []string{brokers[0]},
	}

	evt := &Event{
		Type:      "test",
		CreatedAt: time.Now(),
		Formatted: map[string][]byte{
			JSONFormat: []byte(`"hello world"`),
		},
		Payload: []byte("hello world"),
	}

	config := sarama.NewConfig()
	config.Version = sarama.DefaultVersion
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		t.Fatalf("failed to create kafka client: %v", err)
	}
	admin, err := sarama.NewClusterAdminFromClient(client)
	if err != nil {
		t.Fatalf("failed to create kafka admin client: %v", err)
	}
	t.Cleanup(func() {
		if err := admin.Close(); err != nil {
			t.Fatalf("failed to close admin client: %v", err)
		}
	})

	if err = admin.CreateTopic("test-topic", &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false); err != nil {
		t.Fatalf("failed to create topic: %v", err)
	}
	t.Cleanup(func() {
		if err = admin.DeleteTopic("test-topic"); err != nil {
			t.Fatalf("failed to delete topic: %v", err)
		}
	})

	_, err = sink.Process(ctx, evt)
	if err != nil {
		t.Fatalf("failed to process event: %v", err)
	}

	consumer, err := sarama.NewConsumerGroupFromClient("test-consumer", client)
	if err != nil {
		t.Fatalf("failed to create consumer: %v", err)
	}
	t.Cleanup(func() {
		if err = consumer.Close(); err != nil {
			t.Fatalf("failed to close consumer: %v", err)
		}
	})

	err = consumer.Consume(ctx, []string{"test-topic"}, &testConsumer{
		t: t,
		assert: func(msg *sarama.ConsumerMessage) {
			if msg.Topic != "test-topic" {
				t.Fatalf("expected topic test-topic, got %s", msg.Topic)
			}

			if !bytes.Equal(evt.Formatted[JSONFormat], msg.Value) {
				t.Fatalf("expected %s, got %s", evt.Formatted[JSONFormat], msg.Value)
			}
		},
	})
}

type testConsumer struct {
	t      *testing.T
	assert func(msg *sarama.ConsumerMessage)
}

func (c *testConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *testConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *testConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				c.t.Fatalf("failed to recieve message")
			}

			c.assert(message)
			session.MarkMessage(message, "")
			return nil
		case <-session.Context().Done():
			return nil
		}
	}
}
