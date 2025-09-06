package http

import (
	"encoding/json"
	"net/http"
	"chat-app/server/internal/application"
	"chat-app/server/internal/infrastructure/auth"
	"chat-app/server/internal/infrastructure/transport/websocket"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// NewRouter sets up the application's HTTP routes.
func NewRouter(hub *websocket.Hub, jwtService *auth.JWTService, chatService *application.ChatService) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // Adjust for your client's URL
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// WebSocket connection endpoint
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})

	// REST API endpoints
	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/token", issueTokenHandler(jwtService, chatService))
		r.Get("/groups/search", searchGroupsHandler(chatService))
		// Note: Profile picture uploads would go here as a POST/PUT to /api/users/profile/picture
	})

	return r
}

type issueTokenRequest struct {
	UserID string `json:"userId"`
}

func issueTokenHandler(jwtService *auth.JWTService, chatService *application.ChatService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req issueTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		// Ensure user exists before issuing a token
		if _, err := chatService.GetUser(r.Context(), req.UserID); err != nil {
			http.Error(w, "User not found or not active", http.StatusNotFound)
			return
		}
		
		token, err := jwtService.GenerateToken(req.UserID)
		if err != nil {
			http.Error(w, "Could not generate token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	}
}

func searchGroupsHandler(chatService *application.ChatService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tagQuery := r.URL.Query().Get("tag")
		if tagQuery == "" {
			http.Error(w, "Missing 'tag' query parameter", http.StatusBadRequest)
			return
		}

		groups, err := chatService.FindGroupsByTag(r.Context(), tagQuery)
		if err != nil {
			http.Error(w, "Failed to search for groups", http.StatusInternalServerError)
			return
		}

		// We don't want to expose all member details in a public search
		type GroupSearchResult struct {
			ID                string `json:"id"`
			Name              string `json:"name"`
			ProfilePictureURL string `json:"profilePictureUrl"`
			JoinTag           string `json:"joinTag"`
			MemberCount       int    `json:"memberCount"`
		}

		results := make([]GroupSearchResult, len(groups))
		for i, g := range groups {
			results[i] = GroupSearchResult{
				ID:                g.ID,
				Name:              g.Name,
				ProfilePictureURL: g.ProfilePictureURL,
				JoinTag:           g.JoinTag,
				MemberCount:       len(g.Members),
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}
