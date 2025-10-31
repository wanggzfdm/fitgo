package qwen

import (
	"bytes"
	"context"
	"encoding/json"
	"fitgo/internal/service/ai/client"
	"fitgo/pkg/config"
	"fmt"
	"io"
	"net/http"
)

// QwenClient 千问客户端
type QwenClient struct {
	httpClient *http.Client
	config     *config.AIConfig
}

// NewClient 创建新的千问客户端
func NewClient(cfg *config.AIConfig) (*QwenClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("AI 配置不能为空")
	}

	return &QwenClient{
		httpClient: &http.Client{},
		config:     cfg,
	}, nil
}

// ChatRequest 千问API请求体
type ChatRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

// ChatResponse 千问API响应体
type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Chat 实现 AIClient 接口
func (q *QwenClient) Chat(ctx context.Context, messages []client.ChatMessage) (string, error) {
	// 构建请求体
	reqBody := struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}{
		Model: q.config.Config.Model,
	}

	for _, msg := range messages {
		reqBody.Messages = append(reqBody.Messages, struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求体失败: %v", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		q.config.Config.BaseURL+"/v1/chat/completions",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+q.config.Config.APIKey)

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("调用千问API失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("千问API返回错误: 状态码 %d, 响应: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析API响应失败: %v", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("千问API返回空响应")
	}

	return result.Choices[0].Message.Content, nil
}

// 确保 QwenClient 实现了 AIClient 接口
var _ client.AIClient = (*QwenClient)(nil)
