package llm

import (
	"context"
	"errors"
	"io"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIAdapter implements Adapter using the OpenAI Chat Completions API.
// It also works with any OpenAI-compatible endpoint (OpenRouter, etc.) by
// setting a custom BaseURL.
type OpenAIAdapter struct {
	client       *openai.Client
	defaultModel string
}

// NewOpenAI creates an OpenAI adapter. baseURL may be empty for the default
// OpenAI endpoint. defaultModel is used when the caller passes an empty model.
func NewOpenAI(apiKey, baseURL, defaultModel string) *OpenAIAdapter {
	cfg := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	if defaultModel == "" {
		defaultModel = "gpt-4o-mini"
	}
	return &OpenAIAdapter{
		client:       openai.NewClientWithConfig(cfg),
		defaultModel: defaultModel,
	}
}

func (a *OpenAIAdapter) resolveModel(model string) string {
	if model != "" {
		return model
	}
	return a.defaultModel
}

func (a *OpenAIAdapter) toOpenAIMessages(messages []ChatMessage) []openai.ChatCompletionMessage {
	out := make([]openai.ChatCompletionMessage, len(messages))
	for i, m := range messages {
		out[i] = openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}
	return out
}

// Stream opens a streaming chat completion and returns a channel of events.
func (a *OpenAIAdapter) Stream(ctx context.Context, messages []ChatMessage, model string) (<-chan StreamEvent, error) {
	model = a.resolveModel(model)

	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: a.toOpenAIMessages(messages),
		Stream:   true,
	}

	stream, err := a.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, err
	}

	ch := make(chan StreamEvent, 64)
	go func() {
		defer close(ch)
		defer stream.Close()

		var lastUsage *Usage
		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				ch <- StreamEvent{Done: true, Usage: lastUsage}
				return
			}
			if err != nil {
				ch <- StreamEvent{Done: true, Error: err, Usage: lastUsage}
				return
			}

			// Capture usage from the final chunk if the provider sends it.
			if resp.Usage != nil {
				lastUsage = &Usage{
					PromptTokens:     resp.Usage.PromptTokens,
					CompletionTokens: resp.Usage.CompletionTokens,
					TotalTokens:      resp.Usage.TotalTokens,
				}
			}

			if len(resp.Choices) > 0 && resp.Choices[0].Delta.Content != "" {
				ch <- StreamEvent{Content: resp.Choices[0].Delta.Content}
			}
		}
	}()

	return ch, nil
}

// Complete performs a non-streaming chat completion.
func (a *OpenAIAdapter) Complete(ctx context.Context, messages []ChatMessage, model string) (*CompletionResult, error) {
	model = a.resolveModel(model)

	resp, err := a.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    model,
		Messages: a.toOpenAIMessages(messages),
	})
	if err != nil {
		return nil, err
	}

	result := &CompletionResult{}
	if len(resp.Choices) > 0 {
		result.Content = resp.Choices[0].Message.Content
	}
	result.Usage = Usage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}
	return result, nil
}
