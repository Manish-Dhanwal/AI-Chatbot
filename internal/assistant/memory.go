package assistant

import (
	"context"
	"sync"

	"banking-chatbot/internal/ai"
)

const maxRecentConversationMessages = 20 // Keep the latest 10 user-and-reply turns.

// ConversationStore loads and saves recent messages for one conversation.
type ConversationStore interface {
	History(ctx context.Context, conversationID string) ([]ai.Message, error)
	SaveTurn(ctx context.Context, conversationID, customerMessage, assistantReply string) error
}

// InMemoryConversationStore keeps conversation history only while this server runs.
type InMemoryConversationStore struct {
	mu            sync.RWMutex
	conversations map[string][]ai.Message
}

func NewInMemoryConversationStore() *InMemoryConversationStore {
	return &InMemoryConversationStore{conversations: make(map[string][]ai.Message)}
}

func (s *InMemoryConversationStore) History(ctx context.Context, conversationID string) ([]ai.Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	history := s.conversations[conversationID]
	return append([]ai.Message(nil), history...), nil
}

func (s *InMemoryConversationStore) SaveTurn(ctx context.Context, conversationID, customerMessage, assistantReply string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	history := s.conversations[conversationID]
	history = append(history,
		ai.Message{Role: "user", Content: customerMessage},
		ai.Message{Role: "assistant", Content: assistantReply},
	)
	if len(history) > maxRecentConversationMessages {
		history = history[len(history)-maxRecentConversationMessages:]
	}
	s.conversations[conversationID] = history
	return nil
}
