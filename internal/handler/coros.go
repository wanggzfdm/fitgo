package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	appanalysis "fitgo/internal/application/analysis"
	appcoros "fitgo/internal/application/coros"
	"fitgo/internal/service/pushplus"
	"fitgo/pkg/config"
)

type CorosHandler struct {
	corosQuery    *appcoros.QueryService
	analysisQuery *appanalysis.QueryService
}

func NewCorosHandler(corosQuery *appcoros.QueryService, analysisQuery *appanalysis.QueryService) *CorosHandler {
	return &CorosHandler{
		corosQuery:    corosQuery,
		analysisQuery: analysisQuery,
	}
}

func (h *CorosHandler) Login(w http.ResponseWriter, r *http.Request) {
	_, _ = h.corosQuery.Login(r.Context())
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
	result, err := h.corosQuery.SportsSummary(r.Context(), labelId, sportType)
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
	result, err := h.corosQuery.ActivityList(r.Context(), size, pageNumber, 1)
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

	format := r.URL.Query().Get("format")
	// 调用分析器
	result, err := h.analysisQuery.Analyze(r.Context(), labelId, sportType, format)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.tryPushAnalysis(r.Context(), sportType, result)

	// 直接返回分析结果
	if strings.EqualFold(format, "md") || strings.EqualFold(format, "markdown") {
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	} else if strings.EqualFold(format, "raw") || strings.EqualFold(format, "json") {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	} else {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

// GetLatestAiSportsSummary 获取最新运动数据的AI分析结果
// @Summary 获取最新运动数据的AI分析
// @Description 自动获取最新运动记录并返回AI分析报告
// @Tags AI分析
// @Accept  json
// @Produce text/html; charset=utf-8
// @Param   format  query    string     false       "输出格式: html|md|raw"
// @Success 200 {string} string "成功返回AI分析结果"
// @Failure 500 {string} string "服务器内部错误"
// @Router /coros/ai/latest [get]
func (h *CorosHandler) GetLatestAiSportsSummary(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")

	result, err := h.corosQuery.ActivityList(r.Context(), 1, 1, 1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	labelId, sportType, err := extractLatestActivity(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	analysis, err := h.analysisQuery.Analyze(r.Context(), labelId, sportType, format)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.tryPushAnalysis(r.Context(), sportType, analysis)

	if strings.EqualFold(format, "md") || strings.EqualFold(format, "markdown") {
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	} else if strings.EqualFold(format, "raw") || strings.EqualFold(format, "json") {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	} else {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(analysis))
}

func (h *CorosHandler) tryPushAnalysis(ctx context.Context, sportType, content string) {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		log.Printf("pushplus skipped: load config failed: %v", err)
		return
	}
	if !cfg.PushPlus.Enabled {
		return
	}

	client, err := pushplus.NewClient(&cfg.PushPlus)
	if err != nil {
		log.Printf("pushplus skipped: init client failed: %v", err)
		return
	}

	title := "运动分析报告"
	if sportType != "" {
		title = fmt.Sprintf("运动分析报告 - %s", sportType)
	}

	timeout := cfg.PushPlus.Timeout
	if timeout <= 0 {
		timeout = 10
	}
	pushCtx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	if err := client.Send(pushCtx, title, content); err != nil {
		log.Printf("pushplus send failed: %v", err)
	}
}

func extractLatestActivity(payload map[string]interface{}) (string, string, error) {
	root := payload
	if data, ok := payload["data"].(map[string]interface{}); ok {
		root = data
	}

	if labelId, sportType, ok := readActivityFields(root); ok {
		return labelId, sportType, nil
	}

	list := extractList(root)
	for _, item := range list {
		activity, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if labelId, sportType, ok := readActivityFields(activity); ok {
			return labelId, sportType, nil
		}
	}

	return "", "", fmt.Errorf("无法解析最新运动记录")
}

func extractList(data map[string]interface{}) []interface{} {
	keys := []string{"dataList", "list", "records", "items", "activityList"}
	for _, key := range keys {
		if list, ok := data[key].([]interface{}); ok {
			return list
		}
	}
	return nil
}

func readActivityFields(data map[string]interface{}) (string, string, bool) {
	labelId := firstString(data, "labelId", "label_id", "activityId", "activity_id", "id")
	sportType := firstString(data, "sportType", "sport_type", "sportTypeId", "sport_type_id", "sportId", "sport_id")
	if labelId == "" || sportType == "" {
		return "", "", false
	}
	return labelId, sportType, true
}

func firstString(data map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val, ok := data[key]; ok {
			if str := toString(val); str != "" {
				return str
			}
		}
	}
	return ""
}

func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		if v == 0 {
			return ""
		}
		return strconv.FormatInt(int64(v), 10)
	case int:
		if v == 0 {
			return ""
		}
		return strconv.Itoa(v)
	case int64:
		if v == 0 {
			return ""
		}
		return strconv.FormatInt(v, 10)
	case json.Number:
		return v.String()
	default:
		return ""
	}
}
