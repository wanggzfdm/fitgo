package coros

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestActivityListRefreshesExpiredTokenOnBusinessError(t *testing.T) {
	var loginCalls int
	var activityCalls int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/account/login":
			loginCalls++
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"result":  "0000",
				"message": "OK",
				"data": map[string]interface{}{
					"accessToken": fmt.Sprintf("fresh-token-%d", loginCalls),
				},
			})
		case "/activity/query":
			activityCalls++
			token := r.Header.Get("accesstoken")
			if token != "fresh-token-1" {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"result":  "1019",
					"message": "Access token is invalid",
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"result":  "0000",
				"message": "OK",
				"data": map[string]interface{}{
					"dataList": []interface{}{},
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	t.Setenv("COROS_TOKEN_CACHE_PATH", filepath.Join(t.TempDir(), "coros_token.json"))
	configPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("COROS_CONFIG_PATH", configPath)

	service := &corosService{}
	service.token = "stale-token"
	service.tokenExpire = time.Now().Add(time.Hour)

	writeTestConfig(t, configPath, ts.URL)

	list, err := service.ActivityList(1, 1, 100)
	if err != nil {
		t.Fatalf("ActivityList returned error: %v", err)
	}
	if _, ok := list["data"].(map[string]interface{}); !ok {
		t.Fatalf("expected data in response, got %#v", list)
	}
	if loginCalls != 1 {
		t.Fatalf("expected 1 login call, got %d", loginCalls)
	}
	if activityCalls != 2 {
		t.Fatalf("expected 2 activity calls, got %d", activityCalls)
	}
}

func TestLoginForceRefreshSkipsDiskCache(t *testing.T) {
	var loginCalls int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/account/login":
			loginCalls++
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"result":  "0000",
				"message": "OK",
				"data": map[string]interface{}{
					"accessToken": fmt.Sprintf("fresh-login-token-%d", loginCalls),
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	cachePath := filepath.Join(t.TempDir(), "coros_token.json")
	t.Setenv("COROS_TOKEN_CACHE_PATH", cachePath)
	configPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("COROS_CONFIG_PATH", configPath)
	writeTestConfig(t, configPath, ts.URL)

	if err := saveTokenCache(tokenCache{
		Account:     "runner@example.com",
		AccessToken: "stale-disk-token",
		ExpiresAt:   time.Now().Add(time.Hour),
	}); err != nil {
		t.Fatalf("saveTokenCache failed: %v", err)
	}

	service := &corosService{}
	token, err := service.login(true)
	if err != nil {
		t.Fatalf("login(true) returned error: %v", err)
	}
	if token != "fresh-login-token-1" {
		t.Fatalf("expected fresh login token, got %s", token)
	}
	if loginCalls != 1 {
		t.Fatalf("expected 1 login call, got %d", loginCalls)
	}
}

func TestAccountQueryFallsBackToLoginPayload(t *testing.T) {
	var loginCalls int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/account/login":
			loginCalls++
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"result":  "0000",
				"message": "OK",
				"data": map[string]interface{}{
					"accessToken": "login-token",
					"weight":      70.0,
					"stature":     178.0,
					"birthday":    19990202,
					"countryCode": "CN",
					"maxHr":       195,
					"rhr":         49,
					"hrZoneType":  3,
					"zoneData": map[string]interface{}{
						"lthr": 172,
						"ltsp": 307,
					},
				},
			})
		case "/account/query":
			t.Fatalf("account/query should not be called when login payload already contains profile data")
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	configPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("COROS_CONFIG_PATH", configPath)
	t.Setenv("COROS_TOKEN_CACHE_PATH", filepath.Join(t.TempDir(), "coros_token.json"))
	writeTestConfig(t, configPath, ts.URL)

	service := &corosService{}
	data, err := service.AccountQuery()
	if err != nil {
		t.Fatalf("AccountQuery returned error: %v", err)
	}
	if loginCalls != 1 {
		t.Fatalf("expected 1 login call, got %d", loginCalls)
	}
	if data.Weight != 70 || data.ZoneData.LTHR != 172 || data.ZoneData.LTSP != 307 {
		t.Fatalf("unexpected account query data: %#v", data)
	}
}

func writeTestConfig(t *testing.T, path, address string) {
	t.Helper()
	content := []byte(fmt.Sprintf(`{"coros":{"account":"runner@example.com","accountType":2,"p1":"p1","p2":"p2","address":"%s"}}`, address))
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config failed: %v", err)
	}
}
