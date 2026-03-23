package router

import (
	"fitgo/internal/handler"
	"net/http"
)

// SetupMCPRoutes 设置MCP相关路由
func SetupMCPRoutes(mux *http.ServeMux, mcpHandler *handler.MCPHandler) {
	mux.HandleFunc("GET /mcp/streamhttp", mcpHandler.StreamHTTP)
}

// SetupTcxRoutes 设置TCX相关路由
func SetupTcxRoutes(mux *http.ServeMux, tcxHandler *handler.TCXHandler) {
	mux.HandleFunc("GET /", tcxHandler.Home)
	mux.HandleFunc("POST /upload/tcx", tcxHandler.UploadTCX)
	mux.HandleFunc("GET /summary/{id}", tcxHandler.GetTCXSummary)
	mux.HandleFunc("GET /summaries", tcxHandler.ListTCXSummaries)
}

// SetCorosRoutes 设置Coros相关路由
func SetCorosRoutes(mux *http.ServeMux, corosHandler *handler.CorosHandler) {
	mux.HandleFunc("GET /coros/login", corosHandler.Login)
	mux.HandleFunc("GET /coros/sports/summary", corosHandler.SportsSummary)
	mux.HandleFunc("GET /coros/active", corosHandler.ActivityList)
	mux.HandleFunc("GET /coros/ai/summary", corosHandler.GetAiSportsSummary)
	mux.HandleFunc("GET /coros/ai/latest", corosHandler.GetLatestAiSportsSummary)
}

// SetAgentRoutes 设置AI智能体相关路由
func SetAgentRoutes(mux *http.ServeMux, agentHandler *handler.AIAgentHandler) {
	mux.HandleFunc("POST /ai/agent/analyze", agentHandler.Analyze)
}
