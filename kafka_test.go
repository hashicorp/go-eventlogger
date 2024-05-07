package eventlogger

import (
	"bytes"
	"context"
	"net/url"
	"testing"
	"time"

	"context"
	"os"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
)

func TestKafkaSink_Process(t *testing.T) {
	ctx := context.Background()

	kc := NewKafkaCluster(t, ctx)
	kc.start(t, ctx)
}

const (
	CLUSTER_NETWORK_NAME = "kafka-cluster"
	ZOOKEEPER_PORT       = "2181"
	KAFKA_BROKER_PORT    = "9092"
	KAFKA_CLIENT_PORT    = "9093"
	ZOOKEEPER_IMAGE      = "confluentinc/cp-zookeeper:5.2.1"
	KAFKA_IMAGE          = "confluentinc/cp-kafka:5.2.1"
)

type KafkaCluster struct {
	kafkaContainer     testcontainers.Container
	zookeeperContainer testcontainers.Container
}

func NewKafkaCluster(t *testing.T, ctx context.Context) *KafkaCluster {
	clusterNet, err := network.New(ctx, network.WithDriver("bridge"))
	// creates a network, so kafka and zookeeper can communicate directly
	require.NoError(t, err, "failed to create network")

	zookeeperContainer := createZookeeperContainer(t, ctx, clusterNet)
	kafkaContainer := createKafkaContainer(t, ctx, clusterNet)

	return &KafkaCluster{
		zookeeperContainer: zookeeperContainer,
		kafkaContainer:     kafkaContainer,
	}
}

func (kc *KafkaCluster) start(t *testing.T, ctx context.Context) {

	kc.zookeeperContainer.Start(ctx)
	kc.kafkaContainer.Start(ctx)
	kc.startKafka(t, ctx)
}

func (kc *KafkaCluster) getKafkaHost(t *testing.T, ctx context.Context) string {
	host, err := kc.kafkaContainer.Host(ctx)
	require.NoError(t, err, "failed to get kafka host")

	port, err := kc.kafkaContainer.MappedPort(ctx, KAFKA_CLIENT_PORT)
	require.NoError(t, err, "failed to get kafka port")

	addr, err := url.JoinPath(host, port.Port())
	require.NoError(t, err, "failed to join host and port")
	return addr
}

func (kc *KafkaCluster) startKafka(t *testing.T, ctx context.Context) {
	kafkaStartFile, err := os.CreateTemp("", "testcontainers_start.sh")
	require.NoError(t, err, "failed to create start file in temp directory")
	defer os.Remove(kafkaStartFile.Name())

	// needs to set KAFKA_ADVERTISED_LISTENERS with the exposed kafka port
	exposedHost := kc.getKafkaHost(t, ctx)
	kafkaStartFile.WriteString("#!/bin/bash \n")
	kafkaStartFile.WriteString("export KAFKA_ADVERTISED_LISTENERS='PLAINTEXT://" + exposedHost + ",BROKER://kafka:" + KAFKA_BROKER_PORT + "'\n")
	kafkaStartFile.WriteString(". /etc/confluent/docker/bash-config \n")
	kafkaStartFile.WriteString("/etc/confluent/docker/configure \n")
	kafkaStartFile.WriteString("/etc/confluent/docker/launch \n")

	err = kc.kafkaContainer.CopyFileToContainer(ctx, kafkaStartFile.Name(), "testcontainers_start.sh", 0700)
	require.NoError(t, err, "failed to copy start file to container")
}

func createZookeeperContainer(t *testing.T, ctx context.Context, network *testcontainers.DockerNetwork) testcontainers.Container {
	// creates the zookeeper container, but do not start it yet
	zookeeperContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:          ZOOKEEPER_IMAGE,
			ExposedPorts:   []string{ZOOKEEPER_PORT},
			Env:            map[string]string{"ZOOKEEPER_CLIENT_PORT": ZOOKEEPER_PORT, "ZOOKEEPER_TICK_TIME": "2000"},
			Networks:       []string{network.Name},
			NetworkAliases: map[string][]string{network.Name: {"zookeeper"}},
		},
	})
	require.NoError(t, err, "failed to create zookeeper container")

	return zookeeperContainer
}

func createKafkaContainer(t *testing.T, ctx context.Context, network *testcontainers.DockerNetwork) testcontainers.Container {
	// creates the kafka container, but do not start it yet
	kafkaContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        KAFKA_IMAGE,
			ExposedPorts: []string{KAFKA_CLIENT_PORT},
			Env: map[string]string{
				"KAFKA_BROKER_ID":                        "1",
				"KAFKA_ZOOKEEPER_CONNECT":                "zookeeper:" + ZOOKEEPER_PORT,
				"KAFKA_LISTENERS":                        "PLAINTEXT://0.0.0.0:" + KAFKA_CLIENT_PORT + ",BROKER://0.0.0.0:" + KAFKA_BROKER_PORT,
				"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":   "BROKER:PLAINTEXT,PLAINTEXT:PLAINTEXT",
				"KAFKA_INTER_BROKER_LISTENER_NAME":       "BROKER",
				"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR": "1",
			},
			Networks:       []string{network.Name},
			NetworkAliases: map[string][]string{network.Name: {"kafka"}},
			// the container only starts when it finds and run /testcontainers_start.sh
			Cmd: []string{"sh", "-c", "while [ ! -f /testcontainers_start.sh ]; do sleep 0.1; done; /testcontainers_start.sh"},
		},
	})
	require.NoError(t, err, "failed to create kafka container")

	return kafkaContainer
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
