package main

import (
	"fitgo/internal/handler"
	"fitgo/internal/middleware"
	"fitgo/internal/service/coros"
	"fitgo/internal/service/tcx"
	"fitgo/pkg/config"
	"fmt"
	"net/http"
	"os"

	"fitgo/router"
)

func main() {
	// Load configuration with default paths
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 创建服务实例
	tcxService := tcx.NewTCXService()
	corosService := coros.NewCorosService()

	// 创建处理器
	tcxHandler := handler.NewTCXHandler(tcxService)
	corosHandler := handler.NewCorosHandler(corosService)

	// 创建 ServeMux
	mux := http.NewServeMux()

	// 设置路由
	router.SetupTcxRoutes(mux, tcxHandler)
	router.SetCorosRoutes(mux, corosHandler)

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
