package main

import (
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

const statusQueue = "junction.status"

type PublisherInterface interface {
	Publish(msg StatusMessage) error
}

type Publisher struct {
	ch *amqp.Channel
}

func NewPublisher(ch *amqp.Channel) (*Publisher, error) {
	_, err := ch.QueueDeclare(statusQueue, true, false, false, false, nil)

	if err != nil {
		return nil, err
	}

	return &Publisher{ch: ch}, nil
}

func (p *Publisher) Publish(msg StatusMessage) error {
	body, err := json.Marshal(msg)

	if err != nil {
		return err
	}

	return p.ch.Publish(
		"",          // default exchange
		statusQueue, // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
