package httpapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"banking-chatbot/internal/assistant"
	webui "banking-chatbot/web"
)

// ChatGenerator is the assistant capability the HTTP handler needs.
// Keeping this interface small avoids tying the handler to a specific provider.
type ChatGenerator interface {
	Generate(ctx context.Context, conversationID, prompt string) (assistant.Result, error)
}

type chatRequest struct {
	ConversationID string `json:"conversation_id"`
	Message        string `json:"message"`
}

type chatResponse struct {
	Reply  string `json:"reply"`
	Source string `json:"source"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// NewHandler wires the assistant into the HTTP routes.
func NewHandler(generator ChatGenerator) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		chat(w, r, generator)
	})
	mux.HandleFunc("/", home)
	return mux
}

func chat(w http.ResponseWriter, r *http.Request, generator ChatGenerator) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	var request chatRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "request body must be valid JSON"})
		return
	}

	if strings.TrimSpace(request.Message) == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "message is required"})
		return
	}
	if strings.TrimSpace(request.ConversationID) == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "conversation_id is required"})
		return
	}

	result, err := generator.Generate(r.Context(), request.ConversationID, request.Message)
	if err != nil {
		log.Printf("assistant generation failed: %v", err)
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{
			Error: "The assistant is temporarily unavailable. Please try again.",
		})
		return
	}

	if os.Getenv("LOG_AI_RESPONSE") == "true" {
		log.Printf("AI model response: %s", result.Reply)
	}

	writeJSON(w, http.StatusOK, chatResponse{Reply: result.Reply, Source: result.Source})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	page, err := webui.Files.ReadFile("index.html")
	if err != nil {
		http.Error(w, "chat page is unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(page)
}
