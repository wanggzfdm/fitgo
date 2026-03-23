package client

import "context"

type AIClient interface {
	Chat(ctx context.Context, message []ChatMessage) (string, error)
}

type ChatMessage struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}
