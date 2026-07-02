package main

import "encoding/json"

type Envelope struct {
	Payload json.RawMessage `json:"payload"`
	Meta    Meta            `json:"meta"`
}

type Meta struct {
	TraceID     string      `json:"trace_id"`
	LogID       string      `json:"log_id"`
	Destination Destination `json:"destination"`
}

type Destination struct {
	Type   string     `json:"type"`
	Config HTTPConfig `json:"config"`
}

type HTTPConfig struct {
	URL        string `json:"url"`
	AuthHeader string `json:"auth_header"`
	AuthToken  string `json:"auth_token"`
}

type StatusMessage struct {
	TraceID     string  `json:"trace_id"`
	LogID       string  `json:"log_id"`
	Status      string  `json:"status"`
	AttemptedAt string  `json:"attempted_at"`
	Error       *string `json:"error"`
}
