package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

func mustEnv(key string) string {
	v := os.Getenv(key)

	if v == "" {
		slog.Error("missing required environment variable", "key", key)
		os.Exit(1)
	}

	return v
}

func main() {
	rabbitURL := mustEnv("RABBITMQ_URL")
	apiURL := mustEnv("JUNCTION_API_URL")
	workerToken := mustEnv("JUNCTION_WORKER_TOKEN")

	if err := NewRegistrar(apiURL, workerToken).Register(); err != nil {
		slog.Error("registration failed", "error", err)
		os.Exit(1)
	}

	slog.Info("registered http destination type")

	conn, err := amqp.Dial(rabbitURL)

	if err != nil {
		slog.Error("failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}

	defer conn.Close()

	consumerCh, err := conn.Channel()

	if err != nil {
		slog.Error("failed to open consumer channel", "error", err)
		os.Exit(1)
	}

	defer consumerCh.Close()

	publisherCh, err := conn.Channel()

	if err != nil {
		slog.Error("failed to open publisher channel", "error", err)
		os.Exit(1)
	}

	defer publisherCh.Close()

	publisher, err := NewPublisher(publisherCh)

	if err != nil {
		slog.Error("failed to initialize publisher", "error", err)
		os.Exit(1)
	}

	consumer := NewConsumer(consumerCh, NewDeliverer(), publisher)

	errs := make(chan error, 1)

	go func() {
		errs <- consumer.Run()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		slog.Info("shutting down")
		conn.Close()
	case err := <-errs:
		slog.Error("consumer error", "error", err)
		os.Exit(1)
	}
}
