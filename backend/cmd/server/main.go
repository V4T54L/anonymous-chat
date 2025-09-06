package main

import (
	"log"
	"net/http"
	"time"
	"chat-app/server/internal/application"
	"chat-app/server/internal/infrastructure/auth"
	"chat-app/server/internal/infrastructure/persistence/inmemory"
	"chat-app/server/internal/infrastructure/transport/http"
	"chat-app/server/internal/infrastructure/transport/websocket"
)

func main() {
	// Configuration (in a real app, this would come from a file or env vars)
	jwtSecret := "a_very_secret_key_that_should_be_long_and_random"
	serverAddr := ":8080"

	// Setup Dependencies (Dependency Injection)
	// Infrastructure Layer
	userRepo := inmemory.NewInMemoryUserRepository()
	groupRepo := inmemory.NewInMemoryGroupRepository()
	jwtService := auth.NewJWTService(jwtSecret, 24*time.Hour)

	// Application Layer
	chatService := application.NewChatService(userRepo, groupRepo)

	// WebSocket Hub
	hub := websocket.NewHub(chatService)
	go hub.Run()

	// Transport Layer (HTTP Router)
	router := http.NewRouter(hub, jwtService, chatService)

	log.Printf("Server starting on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, router); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
