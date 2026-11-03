// Package assistant contains the banking assistant's behavior and supplied knowledge.
package assistant

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"banking-chatbot/internal/ai"
)

const systemInstruction = `You are a helpful banking assistant for a learning demo.
Answer clearly and briefly.
Use only facts explicitly stated in the supplied Northstar Bank FAQ entry.
Do not add or guess fees, delivery times, eligibility rules, or other details that the FAQ does not state.
If the FAQ does not answer part of the question, say that the FAQ does not specify it.
This demo cannot access customer accounts or perform banking actions.
Never say or imply that you have ordered, replaced, frozen, cancelled, or initiated anything.`

// Service combines assistant behavior, FAQ search, and model generation.
type Service struct {
	model         ai.Model
	embedder      ai.Embedder
	conversations ConversationStore
	faqEntries    []faqEntry
}

// Result contains the model's reply and the FAQ text used to ground it.
// Source is empty when no FAQ matches and the model is not called.
type Result struct {
	Reply  string
	Source string
}

// NewService embeds each FAQ entry once so later questions can be searched by meaning.
func NewService(ctx context.Context, model ai.Model, embedder ai.Embedder, bankingFAQ string, conversations ConversationStore) (*Service, error) {
	if model == nil {
		return nil, errors.New("AI model is required")
	}
	if embedder == nil {
		return nil, errors.New("embedding model is required")
	}
	if conversations == nil {
		return nil, errors.New("conversation store is required")
	}

	service := &Service{model: model, embedder: embedder, conversations: conversations}
	for _, text := range strings.Split(bankingFAQ, "\n\n") {
		text = strings.TrimSpace(text)
		if text == "" || strings.EqualFold(text, "Northstar Bank (fictional learning data only)") {
			continue
		}

		vector, err := embedder.Embed(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("embed FAQ entry: %w", err)
		}
		service.faqEntries = append(service.faqEntries, faqEntry{text: text, vector: vector})
	}
	if len(service.faqEntries) == 0 {
		return nil, errors.New("banking FAQ is required")
	}

	return service, nil
}

// Generate uses recent conversation messages and the closest FAQ to answer a question.
func (s *Service) Generate(ctx context.Context, conversationID, customerMessage string) (Result, error) {
	if strings.TrimSpace(conversationID) == "" {
		return Result{}, errors.New("conversation ID is required")
	}
	if strings.TrimSpace(customerMessage) == "" {
		return Result{}, errors.New("customer message is required")
	}

	history, err := s.conversations.History(ctx, conversationID)
	if err != nil {
		return Result{}, fmt.Errorf("load conversation history: %w", err)
	}
	retrievalQuery := buildRetrievalQuery(history, customerMessage)
	questionVector, err := s.embedder.Embed(ctx, retrievalQuery)
	if err != nil {
		return Result{}, fmt.Errorf("embed customer question: %w", err)
	}

	faq, score, found := retrieveFAQ(questionVector, s.faqEntries)
	if os.Getenv("LOG_RETRIEVAL") == "true" {
		if found {
			log.Printf("FAQ retrieval score=%.3f selected: %s", score, faq.text)
		} else {
			log.Printf("FAQ retrieval score=%.3f: no relevant entry", score)
		}
	}
	if !found {
		// Do not ask the model to answer without a matching approved FAQ entry.
		reply := "I’m sorry, but the FAQ does not specify that information."
		if err := s.conversations.SaveTurn(ctx, conversationID, customerMessage, reply); err != nil {
			return Result{}, fmt.Errorf("save conversation turn: %w", err)
		}
		return Result{Reply: reply}, nil
	}

	userContent := "Customer question:\n" + customerMessage
	userContent = fmt.Sprintf("Approved bank FAQ entry:\n<faq>\n%s\n</faq>\n\n%s", faq.text, userContent)

	messages := []ai.Message{{Role: "system", Content: systemInstruction}}
	messages = append(messages, history...)
	messages = append(messages, ai.Message{Role: "user", Content: userContent})
	reply, err := s.model.Generate(ctx, messages)
	if err != nil {
		return Result{}, err
	}
	if err := s.conversations.SaveTurn(ctx, conversationID, customerMessage, reply); err != nil {
		return Result{}, fmt.Errorf("save conversation turn: %w", err)
	}
	return Result{Reply: reply, Source: faq.text}, nil
}

// buildRetrievalQuery adds the last two conversation turns so short or repeated
// follow-ups can still find the relevant FAQ.
func buildRetrievalQuery(history []ai.Message, customerMessage string) string {
	start := len(history) - 4
	if start < 0 {
		start = 0
	}

	var query strings.Builder
	for _, message := range history[start:] {
		query.WriteString(message.Content)
		query.WriteByte('\n')
	}
	query.WriteString(customerMessage)
	return query.String()
}
