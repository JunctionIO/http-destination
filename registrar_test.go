package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegistrar_sendsNameAndQueue(t *testing.T) {
	var got map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &got)
		w.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	NewRegistrar(server.URL, "token").Register()

	if got["name"] != "http" {
		t.Errorf("name = %v, want http", got["name"])
	}

	if got["queue"] != "http" {
		t.Errorf("queue = %v, want http", got["queue"])
	}
}

func TestRegistrar_setsAuthHeader(t *testing.T) {
	var token string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token = r.Header.Get("X-Junction-Token")
		w.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	NewRegistrar(server.URL, "worker-jwt").Register()

	if token != "worker-jwt" {
		t.Errorf("X-Junction-Token = %v, want worker-jwt", token)
	}
}

func TestRegistrar_returnsNilOn2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	if err := NewRegistrar(server.URL, "token").Register(); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestRegistrar_returnsErrorOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	defer server.Close()

	if err := NewRegistrar(server.URL, "token").Register(); err == nil {
		t.Error("expected error on 500, got nil")
	}
}

func TestRegistrar_returnsErrorOnNetworkFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	if err := NewRegistrar(server.URL, "token").Register(); err == nil {
		t.Error("expected error on network failure, got nil")
	}
}
