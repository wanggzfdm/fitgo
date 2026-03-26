package main

import (
	"log"
	"os"

	"coros-fit-mcp/internal/mcpserver"
)

func main() {
	addr := getenvDefault("MCP_SSE_ADDR", ":9093")
	baseURL := getenvDefault("MCP_BASE_URL", "http://127.0.0.1:9093")

	remoteServer := mcpserver.NewRemoteServer(mcpserver.New(), baseURL)

	log.Printf(
		"Starting %s remote server on %s (base URL %s, SSE /sse, Streamable HTTP /mcp)",
		mcpserver.ServerName,
		addr,
		baseURL,
	)
	if err := remoteServer.Start(addr); err != nil {
		log.Printf("mcp remote server error: %v\n", err)
		os.Exit(1)
	}
}

func getenvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
