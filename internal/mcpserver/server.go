package mcpserver

import (
	"fmt"

	"coros-fit-mcp/internal/service/activitysummary"
	"coros-fit-mcp/internal/service/coros"
	"coros-fit-mcp/internal/service/personalmetrics"
	"coros-fit-mcp/internal/service/traininganalysis"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	ServerName    = "coros-fit-mcp"
	ServerVersion = "1.0.0"
)

func New() *server.MCPServer {
	mcpServer := server.NewMCPServer(
		ServerName,
		ServerVersion,
		server.WithToolCapabilities(true),
	)

	mcpServer.AddTool(mcp.Tool{
		Name:        "analyze_training_status",
		Description: "分析当前或指定日期的训练状态，并给出后续 3 天训练计划。",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"date": map[string]interface{}{
					"type":        "string",
					"description": "可选，格式 YYYY-MM-DD；不传时分析当前状态和最新活动。",
				},
			},
		},
	}, handleAnalyzeTrainingStatus)

	mcpServer.AddTool(mcp.Tool{
		Name:        "get_training_profile",
		Description: "聚合获取跑者基础资料、训练分区和训练看板，适合分析长期训练能力与分区设置。",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, handleTrainingProfile)

	mcpServer.AddTool(mcp.Tool{
		Name:        "get_training_context",
		Description: "聚合获取训练负荷、最近活动、本周汇总和趋势，适合分析近期运动状态。",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"recent_limit": map[string]interface{}{
					"type":        "number",
					"description": "可选，最近活动返回条数；默认 5。",
				},
				"days": map[string]interface{}{
					"type":        "number",
					"description": "可选，返回最近多少天趋势；默认 7。",
				},
			},
		},
	}, handleTrainingContext)

	mcpServer.AddTool(mcp.Tool{
		Name:        "get_latest_coros_activity_summary",
		Description: "获取高驰账号最新一条运动记录，并转换成统一的运动摘要 JSON。",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, handleLatestCorosActivitySummary)

	mcpServer.AddTool(mcp.Tool{
		Name:        "get_coros_daily_running_summaries",
		Description: "按北京时间返回某一天的跑步摘要列表；仅筛选户外跑步(100)和运动场跑步(103)；未传 date 时默认当天。",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"date": map[string]interface{}{
					"type":        "string",
					"description": "可选，格式 YYYY-MM-DD；未传默认北京时间当天。",
				},
			},
		},
	}, handleDailyCorosRunningSummaries)

	mcpServer.AddTool(mcp.Tool{
		Name:        "get_coros_daily_trail_running_summaries",
		Description: "按北京时间返回某一天的越野跑摘要列表；仅筛选越野跑(102)；未传 date 时默认当天。",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"date": map[string]interface{}{
					"type":        "string",
					"description": "可选，格式 YYYY-MM-DD；未传默认北京时间当天。",
				},
			},
		},
	}, handleDailyCorosTrailRunningSummaries)

	mcpServer.AddTool(mcp.Tool{
		Name:        "summarize_fit_file",
		Description: "读取本地 .fit 文件并转换成统一的运动摘要 JSON。",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "本地 .fit 文件绝对路径或相对路径",
				},
			},
		},
	}, handleSummarizeFITFile)

	return mcpServer
}

func handleLatestCorosActivitySummary(args map[string]interface{}) (*mcp.CallToolResult, error) {
	summary, err := activitysummary.SummarizeLatestCorosActivity(coros.NewCorosService())
	if err != nil {
		return nil, err
	}

	text, err := activitysummary.ResponseText(*summary)
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []interface{}{
			mcp.TextContent{
				Type: "text",
				Text: text,
			},
		},
	}, nil
}

func handleAnalyzeTrainingStatus(args map[string]interface{}) (*mcp.CallToolResult, error) {
	date, _ := args["date"].(string)
	analysis, err := traininganalysis.AnalyzeTrainingStatus(coros.NewCorosService(), date)
	if err != nil {
		return nil, err
	}

	text, err := traininganalysis.ResponseText(*analysis)
	if err != nil {
		return nil, err
	}

	return textToolResult(text), nil
}

func handleTrainingProfile(args map[string]interface{}) (*mcp.CallToolResult, error) {
	bundle, err := personalmetrics.GetTrainingProfileBundle(coros.NewCorosService())
	if err != nil {
		return nil, err
	}

	text, err := personalmetrics.ResponseText("高驰训练档案", bundle)
	if err != nil {
		return nil, err
	}

	return textToolResult(text), nil
}

func handleTrainingContext(args map[string]interface{}) (*mcp.CallToolResult, error) {
	recentLimit := optionalIntArgDefault(args, "recent_limit", 5)
	trendDays := optionalIntArgDefault(args, "days", 7)
	bundle, err := personalmetrics.GetTrainingContextBundle(coros.NewCorosService(), recentLimit, trendDays)
	if err != nil {
		return nil, err
	}

	text, err := personalmetrics.ResponseText("高驰近期训练状态", bundle)
	if err != nil {
		return nil, err
	}

	return textToolResult(text), nil
}

func handleDailyCorosRunningSummaries(args map[string]interface{}) (*mcp.CallToolResult, error) {
	date, _ := args["date"].(string)

	response, err := activitysummary.SummarizeCorosDailyActivities(coros.NewCorosService(), date)
	if err != nil {
		return nil, err
	}

	text, err := activitysummary.DailyResponseText(*response)
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []interface{}{
			mcp.TextContent{
				Type: "text",
				Text: text,
			},
		},
	}, nil
}

func handleDailyCorosTrailRunningSummaries(args map[string]interface{}) (*mcp.CallToolResult, error) {
	date, _ := args["date"].(string)

	response, err := activitysummary.SummarizeCorosDailyTrailRunningActivities(coros.NewCorosService(), date)
	if err != nil {
		return nil, err
	}

	text, err := activitysummary.DailyResponseText(*response)
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []interface{}{
			mcp.TextContent{
				Type: "text",
				Text: text,
			},
		},
	}, nil
}

func handleSummarizeFITFile(args map[string]interface{}) (*mcp.CallToolResult, error) {
	rawPath, ok := args["path"].(string)
	if !ok || rawPath == "" {
		return nil, fmt.Errorf("path 参数必填")
	}

	summary, err := activitysummary.SummarizeFITPath(rawPath)
	if err != nil {
		return nil, err
	}

	text, err := activitysummary.ResponseText(*summary)
	if err != nil {
		return nil, err
	}

	return textToolResult(text), nil
}

func textToolResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []interface{}{
			mcp.TextContent{
				Type: "text",
				Text: text,
			},
		},
	}
}

func optionalIntArg(args map[string]interface{}, key string) int {
	return optionalIntArgDefault(args, key, 0)
}

func optionalIntArgDefault(args map[string]interface{}, key string, fallback int) int {
	raw, ok := args[key]
	if !ok {
		return fallback
	}
	value, ok := raw.(float64)
	if !ok {
		return fallback
	}
	if value <= 0 {
		return fallback
	}
	return int(value)
}
