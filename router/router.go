package router

import (
	"fitgo/internal/handler"
	"net/http"
)

// SetupRoutes 设置路由
func SetupTcxRoutes(mux *http.ServeMux, tcxHandler *handler.TCXHandler) {
	// 注册路由
	mux.HandleFunc("GET /", tcxHandler.Home)
	mux.HandleFunc("POST /upload/tcx", tcxHandler.UploadTCX)
	mux.HandleFunc("GET /summary/{id}", tcxHandler.GetTCXSummary)
	mux.HandleFunc("GET /summaries", tcxHandler.ListTCXSummaries)

}

func SetCorosRoutes(mux *http.ServeMux, corosHandler *handler.CorosHandler) {
	mux.HandleFunc("GET /coros/login", corosHandler.Login)
	mux.HandleFunc("GET /coros/sports/summary", corosHandler.SportsSummary)
	mux.HandleFunc("GET /coros/active", corosHandler.ActivityList)
	mux.HandleFunc("GET /coros/ai/summary", corosHandler.GetAiSportsSummary)
}
