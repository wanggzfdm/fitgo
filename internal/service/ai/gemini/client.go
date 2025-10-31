package gemini

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"fitgo/internal/service/ai/client"
	"fitgo/pkg/config"

	"github.com/sashabaranov/go-openai"
)

// Client 实现 AIClient 接口
type Client struct {
	client *openai.Client
	config *config.AIConfig
}

// NewClient 创建新的 Gemini 客户端
func NewClient(cfg *config.AIConfig) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("AI 配置不能为空")
	}

	openAIConfig := openai.DefaultConfig(cfg.Config.APIKey)

	// 使用配置中的 base_url，如果未设置则使用默认值
	baseURL := "https://generativelanguage.googleapis.com/v1beta"
	if cfg.Config.BaseURL != "" {
		baseURL = cfg.Config.BaseURL
	}
	openAIConfig.BaseURL = baseURL

	// 设置超时
	if cfg.Config.Timeout > 0 {
		openAIConfig.HTTPClient = &http.Client{
			Timeout: time.Duration(cfg.Config.Timeout) * time.Second,
		}
	}

	return &Client{
		client: openai.NewClientWithConfig(openAIConfig),
		config: cfg,
	}, nil
}

// Chat 实现 AIClient 接口
func (g *Client) Chat(ctx context.Context, messages []client.ChatMessage) (string, error) {
	// 转换消息格式
	openAIMessages := make([]openai.ChatCompletionMessage, 0, len(messages))
	for _, msg := range messages {
		openAIMessages = append(openAIMessages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 发送请求
	resp, err := g.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    g.config.Config.Model,
			Messages: openAIMessages,
		},
	)
	if err != nil {
		return "", fmt.Errorf("调用 Gemini API 失败: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("Gemini 返回空响应")
	}

	return resp.Choices[0].Message.Content, nil
}

// 确保 Client 实现了 AIClient 接口
var _ client.AIClient = (*Client)(nil)
