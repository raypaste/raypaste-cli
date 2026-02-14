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

	"raypaste-cli/pkg/types"
)

// processStreamingResponse processes Server-Sent Events (SSE) from the streaming response
func processStreamingResponse(body io.Reader, callback func(string) error) error {
	scanner := bufio.NewScanner(body)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if line == "" {
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

		// Extract content delta
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			if err := callback(chunk.Choices[0].Delta.Content); err != nil {
				return fmt.Errorf("callback error: %w", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %w", err)
	}

	return nil
}
