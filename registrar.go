package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Registrar struct {
	apiURL string
	token  string
	client *http.Client
}

func NewRegistrar(apiURL, token string) *Registrar {
	return &Registrar{
		apiURL: apiURL,
		token:  token,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

var registrationPayload = map[string]any{
	"name":        "http",
	"queue":       "http",
	"description": "Delivers payload via HTTP POST",
	"config_schema": map[string]any{
		"url":         map[string]any{"required": true, "rules": []string{"string"}},
		"auth_header": map[string]any{"required": false, "rules": []string{"string"}},
		"auth_token":  map[string]any{"required": false, "rules": []string{"string"}},
	},
}

func (r *Registrar) Register() error {
	body, err := json.Marshal(registrationPayload)

	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, r.apiURL+"/system/destination-types/register", bytes.NewReader(body))

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Junction-Token", r.token)

	resp, err := r.client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("registration failed: HTTP %d", resp.StatusCode)
	}

	return nil
}
