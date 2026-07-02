package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const deliveryTimeout = 30 * time.Second

type Deliverer struct {
	client *http.Client
}

func NewDeliverer() *Deliverer {
	return &Deliverer{
		client: &http.Client{Timeout: deliveryTimeout},
	}
}

func (d *Deliverer) Deliver(config HTTPConfig, payload json.RawMessage) (status string, errStr *string) {
	req, err := http.NewRequest(http.MethodPost, config.URL, bytes.NewReader(payload))

	if err != nil {
		s := err.Error()
		return "errored", &s
	}

	req.Header.Set("Content-Type", "application/json")

	if config.AuthHeader != "" && config.AuthToken != "" {
		req.Header.Set(config.AuthHeader, config.AuthToken)
	}

	resp, err := d.client.Do(req)

	if err != nil {
		s := err.Error()
		return "errored", &s
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return "dispatched", nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	s := fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))

	return "errored", &s
}
