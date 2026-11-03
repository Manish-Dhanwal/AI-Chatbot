package ai

import "context"

// Message is one turn sent to a chat model, such as a system instruction or user message.
type Message struct {
	Role    string
	Content string
}

// Model generates a reply from an ordered conversation supplied by the application.
type Model interface {
	Generate(ctx context.Context, messages []Message) (string, error)
}

// Embedder turns text into a vector that can be compared by meaning.
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}
