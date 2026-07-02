package main

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

type mockAcknowledger struct {
	ackCount atomic.Int32
}

func (m *mockAcknowledger) Ack(tag uint64, multiple bool) error {
	m.ackCount.Add(1)
	return nil
}

func (m *mockAcknowledger) Nack(tag uint64, multiple bool, requeue bool) error { return nil }
func (m *mockAcknowledger) Reject(tag uint64, requeue bool) error              { return nil }

type mockPublisher struct {
	published []StatusMessage
}

func (m *mockPublisher) Publish(msg StatusMessage) error {
	m.published = append(m.published, msg)
	return nil
}

func makeDelivery(ack *mockAcknowledger, body []byte) amqp.Delivery {
	return amqp.Delivery{Acknowledger: ack, Body: body}
}

func validEnvelope(url string) []byte {
	return []byte(`{"payload":{"key":"value"},"meta":{"trace_id":"trace-uuid","log_id":"log-uuid","destination":{"type":"http","config":{"url":"` + url + `"}}}}`)
}

func TestProcess_acksAndPublishesDispatchedOnSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ack := &mockAcknowledger{}
	pub := &mockPublisher{}
	consumer := &Consumer{deliverer: NewDeliverer(), publisher: pub}

	consumer.process(makeDelivery(ack, validEnvelope(server.URL)))

	if n := ack.ackCount.Load(); n != 1 {
		t.Errorf("Ack called %d times, want 1", n)
	}
	if len(pub.published) != 1 {
		t.Fatalf("published %d messages, want 1", len(pub.published))
	}
	if pub.published[0].Status != "dispatched" {
		t.Errorf("status = %v, want dispatched", pub.published[0].Status)
	}
}

func TestProcess_acksAndPublishesErroredOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ack := &mockAcknowledger{}
	pub := &mockPublisher{}
	consumer := &Consumer{deliverer: NewDeliverer(), publisher: pub}

	consumer.process(makeDelivery(ack, validEnvelope(server.URL)))

	if n := ack.ackCount.Load(); n != 1 {
		t.Errorf("Ack called %d times, want 1", n)
	}
	if len(pub.published) != 1 {
		t.Fatalf("published %d messages, want 1", len(pub.published))
	}
	if pub.published[0].Status != "errored" {
		t.Errorf("status = %v, want errored", pub.published[0].Status)
	}
	if pub.published[0].Error == nil {
		t.Error("Error is nil, want non-nil")
	}
}

func TestProcess_acksAndDiscardsOnMalformedEnvelope(t *testing.T) {
	ack := &mockAcknowledger{}
	pub := &mockPublisher{}
	consumer := &Consumer{deliverer: NewDeliverer(), publisher: pub}

	consumer.process(makeDelivery(ack, []byte("not json")))

	if n := ack.ackCount.Load(); n != 1 {
		t.Errorf("Ack called %d times, want 1", n)
	}
	if len(pub.published) != 0 {
		t.Errorf("published %d messages, want 0", len(pub.published))
	}
}

func TestProcess_acksAndDiscardsOnWrongDestinationType(t *testing.T) {
	ack := &mockAcknowledger{}
	pub := &mockPublisher{}
	consumer := &Consumer{deliverer: NewDeliverer(), publisher: pub}

	body := []byte(`{"payload":{},"meta":{"trace_id":"t","log_id":"l","destination":{"type":"kafka","config":{"url":"http://example.com"}}}}`)

	consumer.process(makeDelivery(ack, body))

	if n := ack.ackCount.Load(); n != 1 {
		t.Errorf("Ack called %d times, want 1", n)
	}
	if len(pub.published) != 0 {
		t.Errorf("published %d messages, want 0", len(pub.published))
	}
}

func TestProcess_setsTraceAndLogIDOnStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ack := &mockAcknowledger{}
	pub := &mockPublisher{}
	consumer := &Consumer{deliverer: NewDeliverer(), publisher: pub}

	consumer.process(makeDelivery(ack, validEnvelope(server.URL)))

	if len(pub.published) != 1 {
		t.Fatalf("published %d messages, want 1", len(pub.published))
	}
	if pub.published[0].TraceID != "trace-uuid" {
		t.Errorf("TraceID = %v, want trace-uuid", pub.published[0].TraceID)
	}
	if pub.published[0].LogID != "log-uuid" {
		t.Errorf("LogID = %v, want log-uuid", pub.published[0].LogID)
	}
}
