package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	mcpgoserver "github.com/mark3labs/mcp-go/server"
)

type RemoteServer struct {
	server   *mcpgoserver.MCPServer
	baseURL  string
	sessions sync.Map
}

type sseSession struct {
	writer  http.ResponseWriter
	flusher http.Flusher
	done    chan struct{}
}

func NewRemoteServer(server *mcpgoserver.MCPServer, baseURL string) *RemoteServer {
	return &RemoteServer{
		server:  server,
		baseURL: baseURL,
	}
}

func (s *RemoteServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/sse", s.handleSSE)
	mux.HandleFunc("/message", s.handleMessage)
	mux.HandleFunc("/mcp", s.handleStreamableHTTP)
	mux.HandleFunc("/healthz", s.handleHealthz)
	return mux
}

func (s *RemoteServer) Start(addr string) error {
	return http.ListenAndServe(addr, s.Handler())
}

func (s *RemoteServer) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"name":   ServerName,
	})
}

func (s *RemoteServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	sessionID := uuid.New().String()
	session := &sseSession{
		writer:  w,
		flusher: flusher,
		done:    make(chan struct{}),
	}

	s.sessions.Store(sessionID, session)
	defer s.sessions.Delete(sessionID)

	messageEndpoint := fmt.Sprintf("%s/message?sessionId=%s", s.baseURL, sessionID)
	_, _ = fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", messageEndpoint)
	flusher.Flush()

	<-r.Context().Done()
	close(session.done)
}

func (s *RemoteServer) handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONRPCError(w, nil, mcp.INVALID_REQUEST, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		writeJSONRPCError(w, nil, mcp.INVALID_PARAMS, "Missing sessionId", http.StatusBadRequest)
		return
	}

	sessionI, ok := s.sessions.Load(sessionID)
	if !ok {
		writeJSONRPCError(w, nil, mcp.INVALID_PARAMS, "Invalid session ID", http.StatusBadRequest)
		return
	}
	session := sessionI.(*sseSession)

	rawMessage, err := decodeJSONBody(r)
	if err != nil {
		writeJSONRPCError(w, nil, mcp.PARSE_ERROR, "Parse error", http.StatusBadRequest)
		return
	}

	response := s.server.HandleMessage(r.Context(), rawMessage)
	if response == nil {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	eventData, err := json.Marshal(response)
	if err != nil {
		writeJSONRPCError(w, nil, mcp.INTERNAL_ERROR, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	_, _ = fmt.Fprintf(session.writer, "event: message\ndata: %s\n\n", eventData)
	session.flusher.Flush()

	writeJSON(w, response, http.StatusAccepted)
}

func (s *RemoteServer) handleStreamableHTTP(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w)

	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
	case http.MethodPost:
		rawMessage, err := decodeJSONBody(r)
		if err != nil {
			writeJSONRPCError(w, nil, mcp.PARSE_ERROR, "Parse error", http.StatusBadRequest)
			return
		}

		response := s.server.HandleMessage(r.Context(), rawMessage)
		if response == nil {
			w.WriteHeader(http.StatusAccepted)
			return
		}

		writeJSON(w, response, http.StatusOK)
	case http.MethodGet:
		writeJSONRPCError(w, nil, mcp.INVALID_REQUEST, "GET is not supported on /mcp; use POST", http.StatusMethodNotAllowed)
	default:
		writeJSONRPCError(w, nil, mcp.INVALID_REQUEST, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func decodeJSONBody(r *http.Request) (json.RawMessage, error) {
	defer r.Body.Close()

	var rawMessage json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawMessage); err != nil {
		return nil, err
	}

	return rawMessage, nil
}

func setCommonHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Mcp-Session-Id, Last-Event-ID")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
}

func writeJSON(w http.ResponseWriter, payload interface{}, statusCode int) {
	setCommonHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONRPCError(w http.ResponseWriter, id interface{}, code int, message string, statusCode int) {
	errPayload := mcp.JSONRPCError{
		JSONRPC: mcp.JSONRPC_VERSION,
		ID:      id,
		Error: struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    interface{} `json:"data,omitempty"`
		}{
			Code:    code,
			Message: message,
		},
	}

	writeJSON(w, errPayload, statusCode)
}

func ShutdownRemoteServer(ctx context.Context, srv *http.Server) error {
	if srv == nil {
		return nil
	}
	return srv.Shutdown(ctx)
}
