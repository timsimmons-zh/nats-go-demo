package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	defaultTopic            = "events"
	defaultNatsClusterURL   = "nats://n1.local.dev:4222"
	defaultNumConsumers     = 0
	defaultNumPublishers    = 0
	defaultMaxPublishedMsgs = 0
	defaultPublishDelay     = 10 * time.Millisecond
)

type Config struct {
	Topic               string
	NatsClusterURL      string
	NumConsumers        int
	NumPublishers       int
	MaxPublishedMsgs    int
	PublishDelay        time.Duration
	DurableConsumerName string
	StreamName          string
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	valueStr := getEnv(key, strconv.Itoa(defaultValue))
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Fatalf("Invalid value for %s: %v", key, err)
	}
	return value
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, defaultValue.String())
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		log.Fatalf("Invalid value for %s: %v", key, err)
	}
	return value
}

func LoadConfig() (*Config, error) {
	config := &Config{
		Topic:               getEnv("NATS_SUBJECT", defaultTopic),
		NatsClusterURL:      getEnv("NATS_CLUSTER_URL", defaultNatsClusterURL),
		NumConsumers:        getEnvInt("NUM_CONSUMERS", defaultNumConsumers),
		NumPublishers:       getEnvInt("NUM_PUBLISHERS", defaultNumPublishers),
		MaxPublishedMsgs:    getEnvInt("MAX_PUBLISHED_MSGS", defaultMaxPublishedMsgs),
		PublishDelay:        getEnvDuration("MAX_PUBLISH_DELAY", defaultPublishDelay),
		DurableConsumerName: getEnv("NATS_DURABLE_CONSUMER", ""),
		StreamName:          getEnv("NATS_STREAM_NAME", ""),
	}
	if err := validateConfig(config); err != nil {
		return nil, err
	}
	return config, nil
}

func validateConfig(c *Config) error {
	if c.StreamName == "" {
		return fmt.Errorf("must specify a Jetstream stream name")
	}
	if c.NumConsumers > 0 && c.DurableConsumerName == "" {
		return fmt.Errorf("value for NATS_DURABLE_CONSUMER must be defined if consumers are enabled")
	}
	return nil
}
