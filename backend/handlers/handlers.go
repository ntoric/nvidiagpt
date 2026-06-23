package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"nvidiagpt/cache"
	"nvidiagpt/models"
	"nvidiagpt/nvidia"
)

type ModelCategory struct {
	Name   string   `json:"name"`
	Models []string `json:"models"`
}

type Handler struct {
	DB              *sql.DB
	Redis           *cache.Redis
	Nvidia          *nvidia.Client
	ModelCategories []ModelCategory
}

func New(db *sql.DB, redis *cache.Redis, nvidiaClient *nvidia.Client, modelCategories []ModelCategory) *Handler {
	return &Handler{DB: db, Redis: redis, Nvidia: nvidiaClient, ModelCategories: modelCategories}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) Models(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Count total unique models
	seen := make(map[string]bool)
	total := 0
	for _, cat := range h.ModelCategories {
		for _, m := range cat.Models {
			if !seen[m] {
				seen[m] = true
				total++
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"categories": h.ModelCategories,
		"total":      total,
	})
}

func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request) {
	convs, err := models.ListConversations(h.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if convs == nil {
		convs = []models.Conversation{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(convs)
}

func (h *Handler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		body.Title = "New Chat"
	}
	if body.Title == "" {
		body.Title = "New Chat"
	}

	conv, err := models.CreateConversation(h.DB, body.Title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(conv)
}

func (h *Handler) GetConversation(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	conv, err := models.GetConversation(h.DB, id)
	if err != nil {
		http.Error(w, "conversation not found", http.StatusNotFound)
		return
	}

	msgs, err := models.ListMessages(h.DB, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if msgs == nil {
		msgs = []models.Message{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"conversation": conv,
		"messages":     msgs,
	})
}

func (h *Handler) DeleteConversation(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := models.DeleteConversation(h.DB, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var body struct {
		Message string `json:"message"`
		Model   string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.Message) == "" {
		http.Error(w, "message is required", http.StatusBadRequest)
		return
	}

	// Verify conversation exists
	if _, err := models.GetConversation(h.DB, id); err != nil {
		http.Error(w, "conversation not found", http.StatusNotFound)
		return
	}

	// Save user message
	if _, err := models.CreateMessage(h.DB, id, "user", body.Message); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Load full conversation history
	msgs, err := models.ListMessages(h.DB, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var nvidiaMsgs []nvidia.Message
	for _, m := range msgs {
		nvidiaMsgs = append(nvidiaMsgs, nvidia.Message{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// Auto-title: if this is the first message, set conversation title
	if len(msgs) == 1 {
		title := body.Message
		if len(title) > 40 {
			title = title[:40] + "..."
		}
		_ = models.UpdateConversationTitle(h.DB, id, title)
	}

	// Stream tokens from NVIDIA API
	var fullResponse strings.Builder
	tokenCount := 0

	err = h.Nvidia.StreamChat(r.Context(), nvidiaMsgs, body.Model, func(content string) {
		fullResponse.WriteString(content)
		tokenCount++

		data, _ := json.Marshal(map[string]string{"content": content})
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	})

	if err != nil && err == context.Canceled {
		// Client disconnected (stop button) — save partial response
		assistantContent := fullResponse.String()
		if assistantContent != "" {
			_, _ = models.CreateMessage(h.DB, id, "assistant", assistantContent)
		}
		_ = models.UpdateConversationTimestamp(h.DB, id)
		return
	}

	if err != nil {
		errData, _ := json.Marshal(map[string]string{"error": err.Error()})
		fmt.Fprintf(w, "data: %s\n\n", errData)
		flusher.Flush()
		return
	}

	// Save assistant response
	assistantContent := fullResponse.String()
	if assistantContent != "" {
		if _, err := models.CreateMessage(h.DB, id, "assistant", assistantContent); err != nil {
			// Log but don't fail the stream
			fmt.Fprintf(w, "data: {\"error\":\"failed to save message\"}\n\n")
			flusher.Flush()
		}
	}

	// Send done event
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()

	_ = models.UpdateConversationTimestamp(h.DB, id)
}

func parseID(r *http.Request) (int, error) {
	// Extract ID from URL path: /api/conversations/{id}/...
	parts := strings.Split(r.URL.Path, "/")
	for i, p := range parts {
		if p == "conversations" && i+1 < len(parts) {
			var id int
			_, err := fmt.Sscanf(parts[i+1], "%d", &id)
			return id, err
		}
	}
	return 0, fmt.Errorf("no id found")
}
