package mcpserver

import (
	"fmt"

	"coros-fit-mcp/internal/service/activitysummary"
	"coros-fit-mcp/internal/service/coros"
	"coros-fit-mcp/internal/service/personalmetrics"

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
		Name:        "get_runner_profile",
		Description: "获取高驰个人基础训练资料，包括身高体重、阈值心率和阈值配速。",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, handleRunnerProfile)

	mcpServer.AddTool(mcp.Tool{
		Name:        "get_training_zones",
		Description: "获取高驰心率区间和配速区间。",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, handleTrainingZones)

	mcpServer.AddTool(mcp.Tool{
		Name:        "get_training_dashboard",
		Description: "获取高驰训练看板，包括跑步能力、恢复、HRV 和个人纪录。",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, handleTrainingDashboard)

	mcpServer.AddTool(mcp.Tool{
		Name:        "get_training_load_status",
		Description: "获取高驰训练负荷状态，包括 ATI、CTI、负荷比和疲劳状态。",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, handleTrainingLoadStatus)

	mcpServer.AddTool(mcp.Tool{
		Name:        "get_recent_activities",
		Description: "获取高驰最近运动列表。",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "可选，返回最近多少条活动；未传时返回全部可用活动。",
				},
			},
		},
	}, handleRecentActivities)

	mcpServer.AddTool(mcp.Tool{
		Name:        "get_weekly_summary",
		Description: "获取高驰本周训练汇总。",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, handleWeeklySummary)

	mcpServer.AddTool(mcp.Tool{
		Name:        "get_training_trends",
		Description: "获取高驰多日训练趋势，包括负荷、VO2 Max 和跑步能力变化。",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"days": map[string]interface{}{
					"type":        "number",
					"description": "可选，返回最近多少天趋势；未传时返回全部可用趋势。",
				},
			},
		},
	}, handleTrainingTrends)

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
		Description: "按北京时间返回某一天的跑步摘要列表；未传 date 时默认当天，返回当日汇总和每次跑步摘要。",
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

func handleRunnerProfile(args map[string]interface{}) (*mcp.CallToolResult, error) {
	profile, err := personalmetrics.GetRunnerProfile(coros.NewCorosService())
	if err != nil {
		return nil, err
	}

	text, err := personalmetrics.ResponseText("高驰跑者资料", profile)
	if err != nil {
		return nil, err
	}

	return textToolResult(text), nil
}

func handleTrainingZones(args map[string]interface{}) (*mcp.CallToolResult, error) {
	zones, err := personalmetrics.GetTrainingZones(coros.NewCorosService())
	if err != nil {
		return nil, err
	}

	text, err := personalmetrics.ResponseText("高驰训练区间", zones)
	if err != nil {
		return nil, err
	}

	return textToolResult(text), nil
}

func handleTrainingDashboard(args map[string]interface{}) (*mcp.CallToolResult, error) {
	dashboard, err := personalmetrics.GetTrainingDashboard(coros.NewCorosService())
	if err != nil {
		return nil, err
	}

	text, err := personalmetrics.ResponseText("高驰训练看板", dashboard)
	if err != nil {
		return nil, err
	}

	return textToolResult(text), nil
}

func handleTrainingLoadStatus(args map[string]interface{}) (*mcp.CallToolResult, error) {
	status, err := personalmetrics.GetTrainingLoadStatus(coros.NewCorosService())
	if err != nil {
		return nil, err
	}

	text, err := personalmetrics.ResponseText("高驰训练负荷状态", status)
	if err != nil {
		return nil, err
	}

	return textToolResult(text), nil
}

func handleRecentActivities(args map[string]interface{}) (*mcp.CallToolResult, error) {
	limit := optionalIntArg(args, "limit")
	activities, err := personalmetrics.GetRecentActivities(coros.NewCorosService(), limit)
	if err != nil {
		return nil, err
	}

	text, err := personalmetrics.ResponseText("高驰最近运动", activities)
	if err != nil {
		return nil, err
	}

	return textToolResult(text), nil
}

func handleWeeklySummary(args map[string]interface{}) (*mcp.CallToolResult, error) {
	summary, err := personalmetrics.GetWeeklySummary(coros.NewCorosService())
	if err != nil {
		return nil, err
	}

	text, err := personalmetrics.ResponseText("高驰本周训练汇总", summary)
	if err != nil {
		return nil, err
	}

	return textToolResult(text), nil
}

func handleTrainingTrends(args map[string]interface{}) (*mcp.CallToolResult, error) {
	days := optionalIntArg(args, "days")
	trends, err := personalmetrics.GetTrainingTrends(coros.NewCorosService(), days)
	if err != nil {
		return nil, err
	}

	text, err := personalmetrics.ResponseText("高驰训练趋势", trends)
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
	raw, ok := args[key]
	if !ok {
		return 0
	}
	value, ok := raw.(float64)
	if !ok {
		return 0
	}
	if value <= 0 {
		return 0
	}
	return int(value)
}
