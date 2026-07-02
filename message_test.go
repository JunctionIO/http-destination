package main

import (
	"encoding/json"
	"testing"
)

func TestEnvelope_unmarshal(t *testing.T) {
	raw := `{
		"payload": {"event": "do.something"},
		"meta": {
			"trace_id": "trace-uuid",
			"log_id": "log-uuid",
			"destination": {
				"type": "http",
				"config": {
					"url": "https://example.com/webhook",
					"auth_header": "Authorization",
					"auth_token": "Bearer token"
				}
			}
		}
	}`

	var env Envelope
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if env.Meta.TraceID != "trace-uuid" {
		t.Errorf("TraceID = %v, want trace-uuid", env.Meta.TraceID)
	}
	if env.Meta.LogID != "log-uuid" {
		t.Errorf("LogID = %v, want log-uuid", env.Meta.LogID)
	}
	if env.Meta.Destination.Type != "http" {
		t.Errorf("Type = %v, want http", env.Meta.Destination.Type)
	}
	if env.Meta.Destination.Config.URL != "https://example.com/webhook" {
		t.Errorf("URL = %v, want https://example.com/webhook", env.Meta.Destination.Config.URL)
	}
	if env.Meta.Destination.Config.AuthHeader != "Authorization" {
		t.Errorf("AuthHeader = %v, want Authorization", env.Meta.Destination.Config.AuthHeader)
	}
	if env.Meta.Destination.Config.AuthToken != "Bearer token" {
		t.Errorf("AuthToken = %v, want Bearer token", env.Meta.Destination.Config.AuthToken)
	}
}

func TestEnvelope_unmarshal_no_auth(t *testing.T) {
	raw := `{
		"payload": {"event": "do.something"},
		"meta": {
			"trace_id": "trace-uuid",
			"log_id": "log-uuid",
			"destination": {
				"type": "http",
				"config": {"url": "https://example.com/webhook"}
			}
		}
	}`

	var env Envelope
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if env.Meta.Destination.Config.AuthHeader != "" {
		t.Errorf("AuthHeader = %v, want empty", env.Meta.Destination.Config.AuthHeader)
	}
	if env.Meta.Destination.Config.AuthToken != "" {
		t.Errorf("AuthToken = %v, want empty", env.Meta.Destination.Config.AuthToken)
	}
}
