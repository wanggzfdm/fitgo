package coros

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"coros-fit-mcp/pkg/config"
)

// TestLiveDebugPersonalMetricsEndpoints is a manual debug test for real COROS APIs.
//
// It is skipped by default. To run it:
//
//	COROS_LIVE_DEBUG=1 go test ./internal/service/coros -run TestLiveDebugPersonalMetricsEndpoints -v
//
// Optional env vars:
//
//	COROS_CONFIG_PATH=/path/to/config.json
//	COROS_DEBUG_WRITE=1   # write raw responses to /tmp/coros-*.json
func TestLiveDebugPersonalMetricsEndpoints(t *testing.T) {
	if os.Getenv("COROS_LIVE_DEBUG") != "1" {
		t.Skip("set COROS_LIVE_DEBUG=1 to run live COROS debug test")
	}

	ensureDebugConfigPath(t)

	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	service := &corosService{}
	token, err := service.Login()
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	t.Logf("login ok, token prefix=%s...", prefix(token, 8))

	if service.loginData != nil {
		body, _ := json.MarshalIndent(service.loginData, "", "  ")
		t.Logf("login data snapshot:\n%s", string(body))
		maybeWriteDebugFile(t, "/tmp/coros-login-data.json", body)
	}

	endpoints := []struct {
		name   string
		method string
		url    string
		body   interface{}
	}{
		{name: "account_query", method: "GET", url: cfg.Coros.Address + "/account/query", body: nil},
		{name: "dashboard_query", method: "GET", url: cfg.Coros.Address + "/dashboard/query", body: nil},
		{name: "dashboard_detail_query", method: "GET", url: cfg.Coros.Address + "/dashboard/detail/query", body: nil},
		{name: "activity_query", method: "GET", url: cfg.Coros.Address + "/activity/query?size=1&pageNumber=1&modeList=100", body: nil},
	}

	for _, endpoint := range endpoints {
		responseBody, err := service.doAuthenticatedRequest(endpoint.method, endpoint.url, endpoint.body)
		if err != nil {
			t.Logf("%s error: %v", endpoint.name, err)
			continue
		}

		formatted := formatJSON(responseBody)
		t.Logf("%s response:\n%s", endpoint.name, formatted)
		maybeWriteDebugFile(t, "/tmp/"+endpoint.name+".json", []byte(formatted))
	}
}

func formatJSON(body []byte) string {
	var parsed interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return string(body)
	}
	pretty, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return string(body)
	}
	return string(pretty)
}

func maybeWriteDebugFile(t *testing.T, path string, body []byte) {
	t.Helper()
	if os.Getenv("COROS_DEBUG_WRITE") != "1" {
		return
	}
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Logf("write debug file %s failed: %v", path, err)
		return
	}
	t.Logf("wrote debug file: %s", path)
}

func prefix(value string, n int) string {
	if len(value) <= n {
		return value
	}
	return fmt.Sprintf("%s", value[:n])
}

func ensureDebugConfigPath(t *testing.T) {
	t.Helper()
	if os.Getenv("COROS_CONFIG_PATH") != "" {
		return
	}

	candidates := []string{
		"configs/config.json",
		"../../configs/config.json",
		"../../../configs/config.json",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			abs, absErr := filepath.Abs(candidate)
			if absErr != nil {
				t.Fatalf("resolve config path failed: %v", absErr)
			}
			if err := os.Setenv("COROS_CONFIG_PATH", abs); err != nil {
				t.Fatalf("set COROS_CONFIG_PATH failed: %v", err)
			}
			t.Logf("using config file: %s", abs)
			return
		}
	}

	t.Fatalf("could not find config file; set COROS_CONFIG_PATH manually")
}
