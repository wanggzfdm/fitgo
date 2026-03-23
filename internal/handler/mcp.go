package handler

import (
	"encoding/json"
	"net/http"
)

// MCPHandler 处理MCP相关请求
type MCPHandler struct{}

// NewMCPHandler 创建新的MCP处理器
func NewMCPHandler() *MCPHandler {
	return &MCPHandler{}
}

// StreamHTTP 处理MCP StreamHTTP请求
func (h *MCPHandler) StreamHTTP(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 获取查询参数
	params := r.URL.Query()

	// 将参数转换为JSON格式
	response, err := json.Marshal(params)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	// 返回参数
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
