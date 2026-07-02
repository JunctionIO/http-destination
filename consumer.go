package main

import (
	"encoding/json"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const deliveryQueue = "junction.destinations.http"

type Consumer struct {
	ch        *amqp.Channel
	deliverer *Deliverer
	publisher PublisherInterface
}

func NewConsumer(ch *amqp.Channel, deliverer *Deliverer, publisher PublisherInterface) *Consumer {
	return &Consumer{ch: ch, deliverer: deliverer, publisher: publisher}
}

func (c *Consumer) Run() error {
	_, err := c.ch.QueueDeclare(deliveryQueue, true, false, false, false, nil)

	if err != nil {
		return err
	}

	msgs, err := c.ch.Consume(deliveryQueue, "", false, false, false, false, nil)

	if err != nil {
		return err
	}

	for msg := range msgs {
		c.process(msg)
	}

	return nil
}

func (c *Consumer) process(msg amqp.Delivery) {
	var env Envelope

	if err := json.Unmarshal(msg.Body, &env); err != nil {
		slog.Error("malformed envelope", "error", err)
		msg.Ack(false)
		return
	}

	if env.Meta.Destination.Type != "http" {
		slog.Error("unexpected destination type", "type", env.Meta.Destination.Type)
		msg.Ack(false)
		return
	}

	status, errStr := c.deliverer.Deliver(env.Meta.Destination.Config, env.Payload)

	statusMsg := StatusMessage{
		TraceID:     env.Meta.TraceID,
		LogID:       env.Meta.LogID,
		Status:      status,
		AttemptedAt: time.Now().UTC().Format(time.RFC3339),
		Error:       errStr,
	}

	if err := c.publisher.Publish(statusMsg); err != nil {
		slog.Error("failed to publish status", "error", err, "trace_id", env.Meta.TraceID)
	}

	msg.Ack(false)
}
