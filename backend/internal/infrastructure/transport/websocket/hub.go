package websocket

import (
	"context"
	"log"
	"sync"
	"time"

	"chat-app/server/internal/application"
)

const groupCleanupTimeout = 5 * time.Minute

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	clients     map[string]*Client          // Map userID to client
	groups      map[string]map[*Client]bool // Map groupID to set of clients
	register    chan *Client
	unregister  chan *Client
	chatService *application.ChatService
	mu          sync.RWMutex
}

func NewHub(chatService *application.ChatService) *Hub {
	return &Hub{
		clients:     make(map[string]*Client),
		groups:      make(map[string]map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		chatService: chatService,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			// Just register the connection for now. Authentication will assign the UserID.
			log.Println("Client connected")
		case client := <-h.unregister:
			h.handleUnregister(client)
		}
	}
}

func (h *Hub) handleMessage(client *Client, msg IncomingMessage) {
	// A giant switch statement is not ideal, but it's simple for this example.
	// A better approach would be a map of message types to handler functions.
	ctx := context.Background()

	switch msg.Type {
	case "authenticate":
		h.handleAuthenticate(client, msg.Payload)
	case "create_group":
		h.handleCreateGroup(client, msg.Payload)
	case "join_group":
		h.handleJoinGroup(client, msg.Payload)
	case "leave_group":
		h.handleLeaveGroup(client, msg.Payload)
	case "send_message":
		h.handleSendMessage(client, msg.Payload)
	case "key_exchange_offer":
		h.handleKeyExchange(client, msg.Payload, "key_exchange_answer")
	case "key_exchange_answer":
		h.handleKeyExchange(client, msg.Payload, "key_exchange_complete")
	case "update_profile":
		h.handleUpdateProfile(client, msg.Payload)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}
