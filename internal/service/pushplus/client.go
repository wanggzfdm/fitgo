package pushplus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"fitgo/pkg/config"
)

const defaultBaseURL = "https://www.pushplus.plus/send"

type Client struct {
	cfg        config.PushPlusConfig
	httpClient *http.Client
}

type sendRequest struct {
	Token    string `json:"token"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Template string `json:"template"`
	Topic    string `json:"topic,omitempty"`
}

type sendResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

func NewClient(cfg *config.PushPlusConfig) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("pushplus config is nil")
	}

	resolved := *cfg
	if resolved.Token == "" {
		resolved.Token = strings.TrimSpace(os.Getenv("PUSHPLUS_TOKEN"))
	}
	if resolved.Template == "" {
		resolved.Template = "markdown"
	}
	if resolved.Timeout <= 0 {
		resolved.Timeout = 10
	}

	return &Client{
		cfg: resolved,
		httpClient: &http.Client{
			Timeout: time.Duration(resolved.Timeout) * time.Second,
		},
	}, nil
}

func (c *Client) Send(ctx context.Context, title, content string) error {
	if !c.cfg.Enabled {
		return nil
	}
	if c.cfg.Token == "" {
		return fmt.Errorf("pushplus token is empty")
	}
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("pushplus content is empty")
	}
	if strings.TrimSpace(title) == "" {
		title = "运动分析报告"
	}

	retry := c.cfg.Retry
	if retry < 0 {
		retry = 0
	}
	interval := time.Duration(c.cfg.RetryIntervalMs) * time.Millisecond
	if interval <= 0 {
		interval = time.Second
	}

	var lastErr error
	for attempt := 0; attempt <= retry; attempt++ {
		lastErr = c.sendOnce(ctx, title, content)
		if lastErr == nil {
			return nil
		}
		if attempt < retry {
			time.Sleep(interval)
		}
	}
	return lastErr
}

func (c *Client) sendOnce(ctx context.Context, title, content string) error {
	reqBody := sendRequest{
		Token:    c.cfg.Token,
		Title:    title,
		Content:  content,
		Template: c.cfg.Template,
	}
	if strings.TrimSpace(c.cfg.Topic) != "" {
		reqBody.Topic = c.cfg.Topic
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal pushplus request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, defaultBaseURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create pushplus request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send pushplus request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read pushplus response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pushplus response status %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	var parsed sendResponse
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return fmt.Errorf("parse pushplus response: %w", err)
	}
	if parsed.Code != 200 {
		return fmt.Errorf("pushplus response error: %s", parsed.Msg)
	}
	return nil
}
