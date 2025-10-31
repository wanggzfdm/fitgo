package service

import (
	"context"
	"fitgo/internal/service/ai/client"
	"fitgo/internal/service/ai/gemini"
	qwenservice "fitgo/internal/service/ai/qwen"
	"fitgo/pkg/config"
	"fmt"
)

// AIService AI 服务
type AIService struct {
	client client.AIClient
}

// NewAIService 创建 AI 服务
func NewAIService(cfg *config.AIConfig) (*AIService, error) {
	if cfg == nil {
		return nil, fmt.Errorf("AI 配置不能为空")
	}

	var aiClient client.AIClient
	var err error

	switch cfg.Provider {
	case "qwen":
		aiClient, err = qwenservice.NewClient(cfg)
		if err != nil {
			return nil, fmt.Errorf("创建千问客户端失败: %v", err)
		}
	case "gemini":
		aiClient, err = gemini.NewClient(cfg)
		if err != nil {
			return nil, fmt.Errorf("创建 Gemini 客户端失败: %v", err)
		}
	default:
		return nil, fmt.Errorf("不支持的AI提供者: %s", cfg.Provider)
	}

	return &AIService{client: aiClient}, nil
}

// Chat 发送聊天消息
func (s *AIService) Chat(ctx context.Context, messages []client.ChatMessage) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("AI 客户端未初始化")
	}
	return s.client.Chat(ctx, messages)
}
