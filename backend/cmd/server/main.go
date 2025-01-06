package main

import (
	"github.com/ndn/backend/internal/server"
	"log"
	"os"
)

func main() {
	// Create and start server
	srv, err := server.New()
	if err != nil {
		log.Printf("Failed to create server: %v\n", err)
		os.Exit(1)
	}

	if err := srv.Start(); err != nil {
		log.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}
