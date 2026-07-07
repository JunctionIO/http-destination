package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDeliver_returnsDispatchedOn2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	status, errStr := NewDeliverer().Deliver(HTTPConfig{URL: server.URL}, json.RawMessage(`{}`), "")

	if status != "dispatched" {
		t.Errorf("status = %v, want dispatched", status)
	}
	if errStr != nil {
		t.Errorf("errStr = %v, want nil", *errStr)
	}
}

func TestDeliver_returnsErroredOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	status, errStr := NewDeliverer().Deliver(HTTPConfig{URL: server.URL}, json.RawMessage(`{}`), "")

	if status != "errored" {
		t.Errorf("status = %v, want errored", status)
	}
	if errStr == nil {
		t.Fatal("errStr is nil, want non-nil")
	}
	if !strings.Contains(*errStr, "502") {
		t.Errorf("errStr = %v, want to contain 502", *errStr)
	}
}

func TestDeliver_setsContentType(t *testing.T) {
	var contentType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	NewDeliverer().Deliver(HTTPConfig{URL: server.URL}, json.RawMessage(`{}`), "")

	if contentType != "application/json" {
		t.Errorf("Content-Type = %v, want application/json", contentType)
	}
}

func TestDeliver_setsAuthHeaderWhenConfigured(t *testing.T) {
	var authHeader string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	NewDeliverer().Deliver(HTTPConfig{
		URL:        server.URL,
		AuthHeader: "Authorization",
		AuthToken:  "Bearer secret",
	}, json.RawMessage(`{}`), "")

	if authHeader != "Bearer secret" {
		t.Errorf("Authorization = %v, want Bearer secret", authHeader)
	}
}

func TestDeliver_doesNotSetAuthHeaderWhenEmpty(t *testing.T) {
	var authHeader string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	NewDeliverer().Deliver(HTTPConfig{URL: server.URL}, json.RawMessage(`{}`), "")

	if authHeader != "" {
		t.Errorf("Authorization = %v, want empty", authHeader)
	}
}

func TestDeliver_setsTraceIDHeaderWhenProvided(t *testing.T) {
	var traceID string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID = r.Header.Get("X-Junction-Trace-ID")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	NewDeliverer().Deliver(HTTPConfig{URL: server.URL}, json.RawMessage(`{}`), "trace-uuid")

	if traceID != "trace-uuid" {
		t.Errorf("X-Junction-Trace-ID = %v, want trace-uuid", traceID)
	}
}

func TestDeliver_doesNotSetTraceIDHeaderWhenEmpty(t *testing.T) {
	var hasHeader bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, hasHeader = r.Header["X-Junction-Trace-Id"]
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	NewDeliverer().Deliver(HTTPConfig{URL: server.URL}, json.RawMessage(`{}`), "")

	if hasHeader {
		t.Error("X-Junction-Trace-ID header was set, want absent")
	}
}

func TestDeliver_returnsErroredOnNetworkFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	status, errStr := NewDeliverer().Deliver(HTTPConfig{URL: server.URL}, json.RawMessage(`{}`), "")

	if status != "errored" {
		t.Errorf("status = %v, want errored", status)
	}
	if errStr == nil {
		t.Fatal("errStr is nil, want non-nil")
	}
}
