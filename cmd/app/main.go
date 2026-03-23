package main

import (
	"fmt"
	"net/http"
	"os"

	"fitgo/internal/agent/manager"
	"fitgo/internal/agent/sports_analyzer"
	appanalysis "fitgo/internal/application/analysis"
	appcoros "fitgo/internal/application/coros"
	apptcx "fitgo/internal/application/tcx"
	"fitgo/internal/handler"
	"fitgo/internal/infrastructure/analysis"
	"fitgo/internal/infrastructure/coros"
	"fitgo/internal/infrastructure/tcx"
	"fitgo/internal/middleware"
	"fitgo/pkg/config"
	"fitgo/router"
)

func main() {
	// Load configuration with default paths
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Infrastructure
	tcxRepo := tcx.NewMemoryRepository()
	corosGateway := coros.NewCorosService()
	analysisGateway := analysis.NewRunningAnalyzer()

	// Application services
	tcxCommandService := apptcx.NewCommandService(tcxRepo)
	tcxQueryService := apptcx.NewQueryService(tcxRepo)
	corosQueryService := appcoros.NewQueryService(corosGateway)
	analysisQueryService := appanalysis.NewQueryService(analysisGateway)

	// 创建处理器
	tcxHandler := handler.NewTCXHandler(tcxCommandService, tcxQueryService)
	corosHandler := handler.NewCorosHandler(corosQueryService, analysisQueryService)
	mcpHandler := handler.NewMCPHandler()

	// 创建 ServeMux
	mux := http.NewServeMux()

	// 设置路由
	router.SetupTcxRoutes(mux, tcxHandler)
	router.SetCorosRoutes(mux, corosHandler)
	router.SetupMCPRoutes(mux, mcpHandler)

	// 创建AI智能体处理器
	agentMgr := manager.New()
	sportsAgent := sportsanalyzer.New()
	if err := agentMgr.Register(sportsAgent); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register sports agent: %v\n", err)
		os.Exit(1)
	}
	agentHandler := handler.NewAIAgentHandler(agentMgr)
	router.SetAgentRoutes(mux, agentHandler)

	// 创建带 CORS 中间件的处理器
	handler := middleware.CORS(mux)

	// 启动服务器
	port := cfg.Server.Port
	fmt.Printf("Server starting on port %s\n", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		fmt.Fprintf(os.Stderr, "Server failed to start: %v\n", err)
		os.Exit(1)
	}
}
