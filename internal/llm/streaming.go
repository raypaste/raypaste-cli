/*
Copyright © 2026 Raypaste
*/
package llm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/raypaste/raypaste-cli/pkg/types"
)

type streamChunkCompat struct {
	Choices []streamChoiceCompat `json:"choices"`
	Error   *types.StreamError   `json:"error,omitempty"`
}

type streamChoiceCompat struct {
	FinishReason string              `json:"finish_reason,omitempty"`
	Delta        streamMessageCompat `json:"delta"`
	Message      streamMessageCompat `json:"message,omitempty"`
}

type streamMessageCompat struct {
	Content json.RawMessage `json:"content,omitempty"`
}

// processStreamingResponse processes Server-Sent Events (SSE) from the streaming response.
//
// SSE comments (lines starting with ":") such as ": OPENROUTER PROCESSING" are
// safely ignored per the SSE spec. Mid-stream errors from OpenRouter are detected
// via the error field or finish_reason "error" in the chunk.
//
// See: https://openrouter.ai/docs/api/reference/streaming
func processStreamingResponse(body io.Reader, callback func(string) error) error {
	scanner := bufio.NewScanner(body)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and SSE comments (e.g. ": OPENROUTER PROCESSING")
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// SSE format: "data: {...}"
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Extract JSON data
		data := strings.TrimPrefix(line, "data: ")

		// Check for done signal
		if data == "[DONE]" {
			break
		}

		// Parse chunk using a compatibility shape so non-string content payloads
		// (array/object) do not cause us to drop valid chunks.
		var chunk streamChunkCompat
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			// Skip malformed chunks
			continue
		}

		// Check for mid-stream error from OpenRouter.
		// When an error occurs after tokens have been sent, the API sends a chunk
		// with an error field and finish_reason "error".
		// See: https://openrouter.ai/docs/api/reference/streaming#handling-errors-during-streaming
		if chunk.Error != nil {
			return fmt.Errorf("stream error from API: %s", chunk.Error.Message)
		}

		for _, choice := range chunk.Choices {
			// Check for error termination via finish_reason
			if choice.FinishReason == "error" {
				return fmt.Errorf("stream terminated with error finish_reason")
			}

			// Prefer delta content. Some providers send content in message.content.
			content := extractStreamContent(choice.Delta.Content)
			if content == "" {
				content = extractStreamContent(choice.Message.Content)
			}

			if content != "" {
				if err := callback(content); err != nil {
					return fmt.Errorf("callback error: %w", err)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %w", err)
	}

	return nil
}

func extractStreamContent(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var value interface{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return ""
	}

	return extractStreamContentFromValue(value)
}

func extractStreamContentFromValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []interface{}:
		var b strings.Builder
		for _, item := range v {
			b.WriteString(extractStreamContentFromValue(item))
		}
		return b.String()
	case map[string]interface{}:
		// Handle content part objects like:
		// {"type":"output_text","text":"..."} and nested wrappers.
		var b strings.Builder
		if text, ok := v["text"]; ok {
			b.WriteString(extractStreamContentFromValue(text))
		}
		if content, ok := v["content"]; ok {
			b.WriteString(extractStreamContentFromValue(content))
		}
		if outputText, ok := v["output_text"]; ok {
			b.WriteString(extractStreamContentFromValue(outputText))
		}
		if valueField, ok := v["value"]; ok {
			b.WriteString(extractStreamContentFromValue(valueField))
		}
		return b.String()
	default:
		return ""
	}
}
