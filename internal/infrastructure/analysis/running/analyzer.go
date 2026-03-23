package running

import (
	"context"
	"fmt"
	"log"
	"strings"

	"fitgo/internal/infrastructure/ai/client"
	aiservice "fitgo/internal/infrastructure/ai/service"
	"fitgo/internal/infrastructure/coros"
	"fitgo/pkg/config"
)

// RunAnalyzer 分析运动数据并返回AI分析结果
func RunAnalyzer(userID, sportID string) (string, error) {
	return RunAnalyzerWithFormat(userID, sportID, "html")
}

// RunAnalyzerWithFormat supports output format selection (html|md).
func RunAnalyzerWithFormat(userID, sportID, format string) (string, error) {
	// 1. 加载配置
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		return "", fmt.Errorf("加载配置失败: %v", err)
	}

	// 2. 创建 AI 服务
	aiService, err := aiservice.NewAIService(&cfg.AI)
	if err != nil {
		return "", fmt.Errorf("创建 AI 服务失败: %v", err)
	}

	// 3. 获取运动概要
	sportsSummary, err := coros.NewCorosService().SportsSummary(context.Background(), userID, sportID)
	if err != nil {
		return "", fmt.Errorf("获取运动概要失败: %v", err)
	}

	// 4. 准备提示词
	prompt := fmt.Sprintf(`
你是一个专业的运动数据分析助手。请分析以下高驰运动数据：

【输出要求】
仅输出**严格 JSON**，不要 Markdown、不要解释、不要代码块。
数值字段必须直接使用输入里的**原始值**，不要做任何计算或单位转换。
无法从输入确认的字段请返回 null 或空字符串，不要臆测。

【JSON Schema】
{
  "training": {
    "type": "string|null",
    "date": "string|null",
    "location": "string|null"
  },
  "summary": {
    "total_distance_m": number|null,
    "total_time_cs": number|null,
    "total_pause_time_cs": number|null,
    "avg_pace_s_per_km": number|null,
    "best_pace_s_per_km": number|null,
    "insight": "string",
    "key_conclusion": "string"
  },
  "metrics": {
    "avg_heart_rate_bpm": number|null,
    "avg_cadence_spm": number|null,
    "avg_power_w": number|null,
    "training_load": number|null,
    "note": "string"
  },
  "pace_consistency": {
    "segments": [
      {
        "segment": "start|middle|end",
        "distance_m": number|null,
        "avg_pace_s_per_km": number|null,
        "avg_heart_rate_bpm": number|null,
        "note": "string"
      }
    ],
    "summary": "string"
  },
  "advice": [
    { "priority": number, "title": "string", "detail": "string" }
  ],
  "next_plan": {
    "goal": "string",
    "pace_range": "string",
    "focus": "string"
  }
}
segments 需要覆盖 start/middle/end 三段；无法判断时填 null 和空字符串。
advice 必须基于本次运动数据给出建议，避免通用模板话术，至少 3 条。
key_conclusion 与 next_plan 也必须基于本次运动数据给出，无法判断时返回空字符串。

运动数据：
每圈数据：%s
总结数据：%s`,
		sportsSummary.LapList, sportsSummary.Summary)
	// 5. 调用AI服务
	response, err := aiService.Chat(context.Background(), []client.ChatMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	})
	if err != nil {
		return "", fmt.Errorf("AI分析失败: %v", err)
	}

	rawJSON, err := extractJSON(response)
	if err != nil {
		return "", fmt.Errorf("AI分析失败: %v", err)
	}
	log.Printf("ai analysis json: %s", rawJSON)

	switch strings.ToLower(strings.TrimSpace(format)) {
	case "raw", "json":
		return rawJSON, nil
	case "md", "markdown":
		analysis, err := parseAIAnalysis(rawJSON)
		if err != nil {
			return "", fmt.Errorf("AI分析失败: %v", err)
		}
		markdown := renderAnalysisMarkdown(analysis, sportsSummary.LapList)
		log.Printf("ai analysis markdown: %s", markdown)
		return markdown, nil
	default:
		analysis, err := parseAIAnalysis(rawJSON)
		if err != nil {
			return "", fmt.Errorf("AI分析失败: %v", err)
		}
		html := renderAnalysisHTML(analysis, sportsSummary.LapList)
		log.Printf("ai analysis html: %s", html)
		return html, nil
	}
}
