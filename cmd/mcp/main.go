package main

import (
	"log"
	"os"

	"coros-fit-mcp/internal/mcpserver"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	if err := server.ServeStdio(mcpserver.New()); err != nil {
		log.Printf("mcp server error: %v\n", err)
		os.Exit(1)
	}
}
