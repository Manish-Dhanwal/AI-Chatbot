package assistant

import (
	"context"
	"errors"
	"strings"
	"testing"

	"banking-chatbot/internal/ai"
)

const testFAQ = "Overdraft fee: $20 per overdraft item."

type testEmbedder struct {
	vectors map[string][]float32
}

func (e testEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	vector, found := e.vectors[text]
	if !found {
		return nil, errors.New("unexpected text sent to test embedder")
	}
	return vector, nil
}

type recordingModel struct {
	calls    int
	messages []ai.Message
}

func (m *recordingModel) Generate(_ context.Context, messages []ai.Message) (string, error) {
	m.calls++
	m.messages = messages
	return "model answer", nil
}

func TestGenerateModelCallDependsOnFAQMatch(t *testing.T) {
	tests := []struct {
		name           string
		question       string
		questionVector []float32
		wantReply      string
		wantSource     string
		wantModelCalls int
	}{
		{
			name:           "does not call model when no FAQ matches",
			question:       "What time does the bank close?",
			questionVector: []float32{0, 1},
			wantReply:      "I’m sorry, but the FAQ does not specify that information.",
			wantSource:     "",
			wantModelCalls: 0,
		},
		{
			name:           "calls model when an FAQ matches",
			question:       "What is the overdraft fee?",
			questionVector: []float32{0.8, 0.6},
			wantReply:      "model answer",
			wantSource:     testFAQ,
			wantModelCalls: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model := &recordingModel{}
			embedder := testEmbedder{vectors: map[string][]float32{
				testFAQ:       {1, 0},
				test.question: test.questionVector,
			}}
			service, err := NewService(context.Background(), model, embedder, testFAQ, NewInMemoryConversationStore())
			if err != nil {
				t.Fatalf("NewService() error = %v", err)
			}

			result, err := service.Generate(context.Background(), "test-conversation", test.question)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			if result.Reply != test.wantReply {
				t.Errorf("reply = %q, want %q", result.Reply, test.wantReply)
			}
			if result.Source != test.wantSource {
				t.Errorf("source = %q, want %q", result.Source, test.wantSource)
			}
			if model.calls != test.wantModelCalls {
				t.Errorf("model calls = %d, want %d", model.calls, test.wantModelCalls)
			}
			if test.wantModelCalls > 0 {
				if len(model.messages) != 2 {
					t.Fatalf("model received %d messages, want system and user messages", len(model.messages))
				}
				userMessage := model.messages[1].Content
				if !strings.Contains(userMessage, testFAQ) {
					t.Errorf("model user message does not contain selected FAQ %q", testFAQ)
				}
				if !strings.Contains(userMessage, test.question) {
					t.Errorf("model user message does not contain customer question %q", test.question)
				}
			}
		})
	}
}
