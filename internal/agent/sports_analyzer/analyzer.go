package sportsanalyzer

import (
	"context"
	"encoding/json"
	"fmt"

	"fitgo/internal/agent/core"
)

// SportsAnalyzer 运动分析智能体
type SportsAnalyzer struct {
	*core.BaseAgent
}

func New() *SportsAnalyzer {
	return &SportsAnalyzer{BaseAgent: core.NewBaseAgent("sports-analyzer", "运动分析专家", "分析运动数据并提供专业建议的智能体")}
}

// Handle 处理运动分析请求
func (a *SportsAnalyzer) Handle(ctx context.Context, input interface{}) (interface{}, error) {
	var req map[string]interface{}
	switch v := input.(type) {
	case []byte:
		if err := json.Unmarshal(v, &req); err != nil {
			return nil, fmt.Errorf("invalid input format: %w", err)
		}
	case string:
		if err := json.Unmarshal([]byte(v), &req); err != nil {
			return nil, fmt.Errorf("invalid input format: %w", err)
		}
	case map[string]interface{}:
		req = v
	default:
		return nil, fmt.Errorf("unsupported input type: %T", input)
	}

	// 简单示例：根据输入生成分析
	metrics := map[string]interface{}{
		"duration":   req["duration"],
		"distance":   req["distance"],
		"calories":   req["calories"],
		"avg_pace":   req["avg_pace"],
		"heart_rate": req["heart_rate"],
		"elevation":  req["elevation"],
	}

	analysis := map[string]interface{}{
		"summary":       "这是一次出色的训练！",
		"performance":   "优秀",
		"recovery_time": "24-48小时",
		"suggestions": []string{
			"保持当前训练强度",
			"注意补充水分和营养",
			"确保有足够的休息时间",
		},
		"metrics": metrics,
	}

	return analysis, nil
}
