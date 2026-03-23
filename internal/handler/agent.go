package handler

import (
	"encoding/json"
	"net/http"

	"fitgo/internal/agent/manager"
)

// AIAgentHandler handles AI agent related requests
type AIAgentHandler struct {
	agentMgr *manager.Manager
}

// NewAIAgentHandler creates a new AIAgentHandler
func NewAIAgentHandler(agentMgr *manager.Manager) *AIAgentHandler {
	return &AIAgentHandler{
		agentMgr: agentMgr,
	}
}

// Analyze handles the AI analysis request
func (h *AIAgentHandler) Analyze(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.agentMgr.Process(r.Context(), "sports-analyzer", input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    result,
	})
}
