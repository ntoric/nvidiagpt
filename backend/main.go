package main

import (
	"log"
	"net/http"
	"strings"

	"nvidiagpt/cache"
	"nvidiagpt/db"
	"nvidiagpt/handlers"
	"nvidiagpt/nvidia"
)

func main() {
	cfg := LoadConfig()

	// Connect to PostgreSQL
	database, err := db.New(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	log.Println("Connected to PostgreSQL")

	// Connect to Redis
	redisCache, err := cache.New(cfg.RedisHost, cfg.RedisPort)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisCache.Client.Close()
	log.Println("Connected to Redis")

	// Create NVIDIA client
	nvidiaClient := nvidia.New(cfg.NvidiaAPIKey, cfg.NvidiaAPIURL, cfg.NvidiaModel)

	// Create handler
	h := handlers.New(database, redisCache, nvidiaClient, cfg.AvailableModels)

	// Routes
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", h.Health)
	mux.HandleFunc("/api/models", h.Models)
	mux.HandleFunc("/api/conversations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListConversations(w, r)
		case http.MethodPost:
			h.CreateConversation(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/conversations/", func(w http.ResponseWriter, r *http.Request) {
		// /api/conversations/{id}  or /api/conversations/{id}/messages
		path := r.URL.Path
		if strings.HasSuffix(path, "/messages") || strings.HasSuffix(path, "/chat") {
			if strings.HasSuffix(path, "/chat") && r.Method == http.MethodPost {
				h.Chat(w, r)
				return
			}
			if strings.HasSuffix(path, "/messages") && r.Method == http.MethodGet {
				h.GetConversation(w, r)
				return
			}
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		switch r.Method {
		case http.MethodGet:
			h.GetConversation(w, r)
		case http.MethodDelete:
			h.DeleteConversation(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// CORS middleware
	handler := corsMiddleware(mux)

	addr := ":" + cfg.Port
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
