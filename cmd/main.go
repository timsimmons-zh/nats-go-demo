package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"golang.org/x/sync/errgroup"
)

func handleMessage(msg jetstream.Msg) {
	log.Println("Received message: ", string(msg.Data()))
	msg.Ack()
}

func consumeEvents(ctx context.Context, cons jetstream.Consumer, id int) func() error {
	return func() error {
		consCtx, err := cons.Consume(
			func(msg jetstream.Msg) {
				log.Printf("Received message(consumer=%d): %v\n", id, string(msg.Data()))
				msg.DoubleAck(ctx) // Double-Ack to ensure the message is removed from the stream
			},
			jetstream.ConsumeErrHandler(func(_ jetstream.ConsumeContext, err error) {
				log.Printf("consumer error (id=%d): %v", id, err)
			}),
		)
		if err != nil {
			return err
		}

		log.Printf("Consumer running(id=%d)...\n", id)

		// Wait for shutdown and then drain the consumer
		<-ctx.Done()
		log.Printf("Draining consumer(id=%d)...\n", id)
		consCtx.Drain()
		consCtx.Stop()
		log.Printf("Consumer closed(id=%d)\n", id)
		return err
	}
}

func publishEvents(ctx context.Context, js jetstream.JetStream, subject string, delay time.Duration, maxMsgs int) func() error {
	return func() error {
		for i := 0; ; i++ {
			if maxMsgs > 0 && i >= maxMsgs {
				log.Println("published max messages.")
				return fmt.Errorf("hit published message limit")
			}
			select {
			case <-ctx.Done():
				log.Println("publisher complete.")
				return ctx.Err()
			default:
				log.Println("Publishing msg: ", i)
				msg := fmt.Sprintf("msg %d", i)
				sub := fmt.Sprintf("%v.%d", subject, i)
				if _, err := js.Publish(ctx, sub, []byte(msg)); err != nil {
					log.Println("failed to publish message: ", err)
					return err
				}
				time.Sleep(delay)
			}
		}
	}
}

func main() {
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	time.Sleep(10 * time.Second)
	log.Println("Starting connection...")

	nc, err := nats.Connect(config.NatsClusterURL, nats.Timeout(10*time.Second))
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Drain()

	// Wait until the NATS connection is established before continuing
	for !nc.IsConnected() {
		time.Sleep(50 * time.Millisecond)
	}

	// Listen for SIGTERM and cancel the context when received
	ctx, shutdown := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer shutdown()

	log.Println("Connected to NATS cluster at ", config.NatsClusterURL)

	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatalf("Failed to open Jetstream: %v", err)
	}

	log.Println("Jetstream context established")

	s, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      config.StreamName, // multiple replicas for the same service should use the same stream name to ensure even distribution of messages
		Subjects:  []string{config.Topic + ".*"},
		Retention: jetstream.WorkQueuePolicy, // ensures the msg is removed from the stream after an ACK
	})
	if err != nil {
		log.Fatalf("Failed to connect to stream: %v", err)
	}

	log.Println("Jetstream CreateStream")

	cons, err := s.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:   config.DurableConsumerName,
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		log.Fatal("Failed to create consumer: ", err)
	}

	g, groupCtx := errgroup.WithContext(ctx)

	for n := 1; n <= config.NumPublishers; n++ {
		g.Go(publishEvents(groupCtx, js, config.Topic, config.PublishDelay, config.MaxPublishedMsgs))
	}

	for n := 1; n <= config.NumConsumers; n++ {
		g.Go(consumeEvents(groupCtx, cons, n))
	}

	<-groupCtx.Done()
	log.Println("shutdown sequence...")

	g.Wait()

	log.Println("Shutdown complete.")
}
