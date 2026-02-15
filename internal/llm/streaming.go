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

		// Parse chunk
		var chunk types.StreamChunk
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

		if len(chunk.Choices) > 0 {
			// Check for error termination via finish_reason
			if chunk.Choices[0].FinishReason == "error" {
				return fmt.Errorf("stream terminated with error finish_reason")
			}

			// Extract content delta
			if chunk.Choices[0].Delta.Content != "" {
				if err := callback(chunk.Choices[0].Delta.Content); err != nil {
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
