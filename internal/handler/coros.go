package handler

import (
	"encoding/json"
	"fitgo/internal/service/analyzer/running"
	"net/http"
	"strconv"

	"fitgo/internal/service/coros"
)

type CorosHandler struct {
	corosService coros.CorosService
}

func NewCorosHandler(service coros.CorosService) *CorosHandler {
	return &CorosHandler{
		corosService: service,
	}
}

func (h *CorosHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.corosService.Login()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Login successful"))
}

func (h *CorosHandler) SportsSummary(w http.ResponseWriter, r *http.Request) {
	// 从查询参数中获取 labelId 和 sportType
	labelId := r.URL.Query().Get("labelId")
	sportType := r.URL.Query().Get("sportType")

	// 验证必填参数
	if labelId == "" || sportType == "" {
		http.Error(w, "Both labelId and sportType are required", http.StatusBadRequest)
		return
	}

	// 调用服务层方法
	result, err := h.corosService.SportsSummary(labelId, sportType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *CorosHandler) ActivityList(w http.ResponseWriter, r *http.Request) {
	// 从查询参数中获取 size 和 pageNumber
	sizeStr := r.URL.Query().Get("size")
	pageNumberStr := r.URL.Query().Get("pageNumber")

	// 转换参数为整数
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		http.Error(w, "size 参数必须是整数", http.StatusBadRequest)
		return
	}

	pageNumber, err := strconv.Atoi(pageNumberStr)
	if err != nil {
		http.Error(w, "pageNumber 参数必须是整数", http.StatusBadRequest)
		return
	}

	// 调用服务层方法
	result, err := h.corosService.ActivityList(size, pageNumber, 1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetAiSportsSummary 获取运动数据的AI分析结果
// @Summary 获取运动数据的AI分析
// @Description 根据用户ID和运动ID获取AI分析的运动数据报告
// @Tags AI分析
// @Accept  json
// @Produce text/plain; charset=utf-8
// @Param   labelId    query    string     true        "运动记录ID"
// @Param   sportType  query    string     true        "运动类型"
// @Success 200 {string} string "成功返回AI分析结果"
// @Failure 400 {string} string "请求参数错误"
// @Failure 500 {string} string "服务器内部错误"
// @Router /coros/ai/summary [get]
func (h *CorosHandler) GetAiSportsSummary(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	labelId := r.URL.Query().Get("labelId")
	sportType := r.URL.Query().Get("sportType")

	// 验证必要参数
	if labelId == "" || sportType == "" {
		http.Error(w, "labelId and sportType are required", http.StatusBadRequest)
		return
	}

	// 调用分析器
	result, err := running.RunAnalyzer(labelId, sportType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 直接返回分析结果
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}
