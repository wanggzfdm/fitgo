package ai_test

import (
	"context"
	"fmt"
	"testing"

	"fitgo/internal/service/ai/client"
	aiservice "fitgo/internal/service/ai/service"
	"fitgo/internal/service/analyzer/running"
	"fitgo/pkg/config"
)

func TestAIService(t *testing.T) {
	// 1. 加载配置
	cfg, err := config.LoadConfig("../../configs/config.json")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 2. 创建 AI 服务
	aiService, err := aiservice.NewAIService(&cfg.AI)
	if err != nil {
		t.Fatalf("创建AI服务失败: %v", err)
	}

	// 3. 使用 AI 服务
	response, err := aiService.Chat(context.Background(), []client.ChatMessage{
		{
			Role:    "user",
			Content: "我是高驰设备，我会给你高驰的运动概要，你帮我分析一下",
		},
	})
	if err != nil {
		t.Fatalf("AI聊天失败: %v", err)
	}

	fmt.Println("AI回复:", response)
}

func TestAIServiceSummary(t *testing.T) {
	// 测试用的用户ID和运动ID
	userID := "472913588747534541"
	sportID := "100"

	// 调用 RunAnalyzer 函数
	result, err := running.RunAnalyzer(userID, sportID)
	if err != nil {
		t.Fatalf("RunAnalyzer 执行失败: %v", err)
	}

	// 检查返回结果是否为空
	if result == "" {
		t.Error("预期得到非空的分析结果，但结果为空")
	}

	// 打印分析结果（可选，用于调试）
	t.Logf("运动数据分析结果：\n%s", result)
}

func TestGeminiClient(t *testing.T) {
	// 1. 加载配置
	cfg, err := config.LoadConfig("../../configs/config.json")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	aiService, err := aiservice.NewAIService(&cfg.AI)
	if err != nil {
		t.Fatalf("创建AI服务失败: %v", err)
	}

	response, err := aiService.Chat(context.Background(), []client.ChatMessage{
		{
			Role:    "user",
			Content: "你好，请用中文回答，你是谁？",
		},
	})
	if err != nil {
		t.Fatalf("AI聊天失败: %v", err)
	}

	t.Logf("AI回复: %s", response)
}
