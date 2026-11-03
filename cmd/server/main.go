package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"banking-chatbot/internal/ai"
	"banking-chatbot/internal/assistant"
	"banking-chatbot/internal/httpapi"
)

func main() {
	bankingFAQ, err := os.ReadFile("knowledge/bank_faq.txt")
	if err != nil {
		log.Fatalf("load banking FAQ: %v", err)
	}

	ollamaClient, err := ai.NewOllamaClient("gpt-oss:20b")
	if err != nil {
		log.Fatalf("initialize Ollama client: %v", err)
	}

	conversationStore := assistant.NewInMemoryConversationStore()
	assistantService, err := assistant.NewService(context.Background(), ollamaClient, ollamaClient, string(bankingFAQ), conversationStore)
	if err != nil {
		log.Fatalf("initialize banking assistant: %v", err)
	}

	server := &http.Server{
		Addr:              ":8080",
		Handler:           httpapi.NewHandler(assistantService),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Banking Chatbot API listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
