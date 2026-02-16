/*
Copyright © 2026 Raypaste
*/
package types

// OutputLength represents the desired response length from the LLM
type OutputLength string

const (
	OutputLengthShort  OutputLength = "short"
	OutputLengthMedium OutputLength = "medium"
	OutputLengthLong   OutputLength = "long"
)

// Message represents a chat message in the OpenRouter API format
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest represents a request to the OpenRouter API
type CompletionRequest struct {
	Model               string    `json:"model"`
	Messages            []Message `json:"messages"`
	MaxTokens           int       `json:"max_tokens,omitempty"`
	MaxCompletionTokens int       `json:"max_completion_tokens,omitempty"`
	ReasoningEffort     string    `json:"reasoning_effort,omitempty"`
	Temperature         float64   `json:"temperature,omitempty"`
	Stream              bool      `json:"stream,omitempty"`
}

// TokenUsage represents token usage statistics from the API
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Choice represents a choice in a completion response
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// CompletionResponse represents a response from the OpenRouter API
type CompletionResponse struct {
	ID      string     `json:"id"`
	Model   string     `json:"model"`
	Choices []Choice   `json:"choices"`
	Usage   TokenUsage `json:"usage"`
}

// Delta represents a delta update in a streaming response
type Delta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// StreamChoice represents a choice in a streaming response chunk
type StreamChoice struct {
	Index        int    `json:"index"`
	Delta        Delta  `json:"delta"`
	FinishReason string `json:"finish_reason,omitempty"`
}

// StreamChunk represents a streaming response chunk from OpenRouter.
// When a mid-stream error occurs, OpenRouter sends a chunk with an Error field
// and finish_reason "error" in the choices array.
// See: https://openrouter.ai/docs/api/reference/streaming#handling-errors-during-streaming
type StreamChunk struct {
	ID      string         `json:"id"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
	Error   *StreamError   `json:"error,omitempty"`
}

// StreamError represents an error delivered mid-stream via SSE from OpenRouter.
type StreamError struct {
	Code    interface{} `json:"code"` // may be int or string depending on error type
	Message string      `json:"message"`
}

// APIError represents an error from the OpenRouter API
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// ErrorResponse represents an error response from the OpenRouter API
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// LengthParams holds parameters for a specific output length
type LengthParams struct {
	MaxTokens int
	Directive string
}
