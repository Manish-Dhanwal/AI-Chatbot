// Package ai contains clients for external and local AI model providers.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	ollamaChatEndpoint  = "http://localhost:11434/api/chat"
	ollamaEmbedEndpoint = "http://localhost:11434/api/embed"
	embeddingModel      = "nomic-embed-text"
	maxResponseSize     = 1 << 20
)

// OllamaClient sends conversations to a local Ollama server.
type OllamaClient struct {
	model      string
	httpClient *http.Client
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaChatResponse struct {
	Message ollamaMessage `json:"message"`
}

type ollamaEmbedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type ollamaEmbedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

// NewOllamaClient creates a reusable client for the given locally installed chat model.
func NewOllamaClient(model string) (*OllamaClient, error) {
	model = strings.TrimSpace(model)
	if model == "" {
		return nil, errors.New("Ollama model name is required")
	}

	return &OllamaClient{
		model: model,
		httpClient: &http.Client{
			Timeout: 3 * time.Minute,
		},
	}, nil
}

// Generate sends the supplied conversation to Ollama and returns its reply.
func (c *OllamaClient) Generate(ctx context.Context, messages []Message) (string, error) {
	if len(messages) == 0 {
		return "", errors.New("at least one message is required")
	}

	ollamaMessages := make([]ollamaMessage, 0, len(messages))
	for _, message := range messages {
		ollamaMessages = append(ollamaMessages, ollamaMessage{
			Role:    message.Role,
			Content: message.Content,
		})
	}

	requestBody := ollamaChatRequest{
		Model:    c.model,
		Messages: ollamaMessages,
		Stream:   false,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("encode Ollama request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ollamaChatEndpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create Ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send Ollama request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		message, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
		if len(message) == 0 {
			return "", fmt.Errorf("Ollama returned HTTP %d", resp.StatusCode)
		}
		return "", fmt.Errorf("Ollama returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(message)))
	}

	var result ollamaChatResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseSize)).Decode(&result); err != nil {
		return "", fmt.Errorf("decode Ollama response: %w", err)
	}
	if strings.TrimSpace(result.Message.Content) == "" {
		return "", errors.New("Ollama returned an empty response")
	}

	return result.Message.Content, nil
}

// Embed converts text into a vector using Ollama's local embedding model.
func (c *OllamaClient) Embed(ctx context.Context, text string) ([]float32, error) {
	if strings.TrimSpace(text) == "" {
		return nil, errors.New("text to embed is required")
	}

	body, err := json.Marshal(ollamaEmbedRequest{Model: embeddingModel, Input: text})
	if err != nil {
		return nil, fmt.Errorf("encode Ollama embedding request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ollamaEmbedEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create Ollama embedding request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send Ollama embedding request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		message, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
		return nil, fmt.Errorf("Ollama embedding API returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(message)))
	}

	var result ollamaEmbedResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseSize)).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode Ollama embedding response: %w", err)
	}
	if len(result.Embeddings) == 0 || len(result.Embeddings[0]) == 0 {
		return nil, errors.New("Ollama returned an empty embedding")
	}

	return result.Embeddings[0], nil
}
