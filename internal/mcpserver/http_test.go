package mcpserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStreamableHTTPInitialize(t *testing.T) {
	remoteServer := NewRemoteServer(New(), "http://localhost:9093")
	reqBody := []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}`)

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	remoteServer.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload["jsonrpc"] != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %#v", payload["jsonrpc"])
	}

	if payload["id"].(float64) != 1 {
		t.Fatalf("expected id 1, got %#v", payload["id"])
	}
}

func TestStreamableHTTPRejectsGet(t *testing.T) {
	remoteServer := NewRemoteServer(New(), "http://localhost:9093")
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	rec := httptest.NewRecorder()

	remoteServer.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestServerRegistersPersonalMetricsTools(t *testing.T) {
	remoteServer := NewRemoteServer(New(), "http://localhost:9093")
	reqBody := []byte(`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`)
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	remoteServer.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload struct {
		Result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode tools list response: %v", err)
	}

	names := map[string]bool{}
	for _, tool := range payload.Result.Tools {
		names[tool.Name] = true
	}

	for _, name := range []string{
		"get_runner_profile",
		"get_training_zones",
		"get_training_dashboard",
		"get_training_load_status",
		"get_recent_activities",
		"get_weekly_summary",
		"get_training_trends",
	} {
		if !names[name] {
			t.Fatalf("expected tool %s to be registered", name)
		}
	}
}
