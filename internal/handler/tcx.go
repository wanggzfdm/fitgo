package handler

import (
	"encoding/json"
	"fitgo/internal/service/tcx"
	"fmt"
	"net/http"
)

type TCXHandler struct {
	tcxService tcx.TCXService
}

func NewTCXHandler(service tcx.TCXService) *TCXHandler {
	return &TCXHandler{
		tcxService: service,
	}
}

func (h *TCXHandler) Home(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "FitGo TCX Processor API",
		"version": "1.0.0",
		"routes":  []string{"/upload", "/summary/{id}", "/summaries"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *TCXHandler) UploadTCX(w http.ResponseWriter, r *http.Request) {
	// 只允许POST方法
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析表单数据，最大内存32MB
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	// 获取上传的文件
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 验证文件扩展名
	if len(header.Filename) < 4 || header.Filename[len(header.Filename)-4:] != ".tcx" {
		http.Error(w, "Only .tcx files are allowed", http.StatusBadRequest)
		return
	}

	// 调用服务处理文件
	summary, err := h.tcxService.UploadTCX(file, header.Filename)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to process TCX file: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(summary)
}

func (h *TCXHandler) GetTCXSummary(w http.ResponseWriter, r *http.Request) {
	// 从URL参数获取ID
	id := r.PathValue("id") // Go 1.22+ 的新特性
	if id == "" {
		// 兼容旧版本 Go
		// 从 URL 查询参数获取 ID
		id = r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Missing ID parameter", http.StatusBadRequest)
			return
		}
	}

	summary, err := h.tcxService.GetTCXSummary(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get TCX summary: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (h *TCXHandler) ListTCXSummaries(w http.ResponseWriter, r *http.Request) {
	summaries, err := h.tcxService.ListTCXSummaries()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list TCX summaries: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}
