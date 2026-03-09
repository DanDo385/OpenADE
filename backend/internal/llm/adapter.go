package llm

import "context"

// ChatMessage is a single message in a conversation sent to an LLM.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// StreamEvent represents one event emitted during a streaming LLM response.
type StreamEvent struct {
	Content string // partial content token
	Done    bool   // true on the final event
	Usage   *Usage // populated on the final event (may be nil)
	Error   error  // non-nil if the stream encountered an error
}

// Usage holds token counts from a completed LLM call.
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// CompletionResult is the result of a non-streaming LLM call.
type CompletionResult struct {
	Content string
	Usage   Usage
}

// Adapter is the interface every LLM provider must implement.
type Adapter interface {
	// Stream sends messages to the LLM and returns a channel of streaming events.
	// The channel is closed when the stream is done or on error.
	Stream(ctx context.Context, messages []ChatMessage, model string) (<-chan StreamEvent, error)

	// Complete sends messages to the LLM and returns the full response at once.
	Complete(ctx context.Context, messages []ChatMessage, model string) (*CompletionResult, error)
}

// EstimateCost returns approximate cost in USD for the given model and token counts.
func EstimateCost(model string, promptTokens, completionTokens int) float64 {
	type pricing struct {
		input  float64 // per 1M tokens
		output float64 // per 1M tokens
	}
	prices := map[string]pricing{
		"gpt-4o-mini":   {0.15, 0.60},
		"gpt-4o":        {2.50, 10.00},
		"gpt-4-turbo":   {10.00, 30.00},
		"gpt-3.5-turbo": {0.50, 1.50},
	}
	p, ok := prices[model]
	if !ok {
		p = pricing{1.00, 3.00} // rough default for unknown models
	}
	return (float64(promptTokens)*p.input + float64(completionTokens)*p.output) / 1_000_000
}
