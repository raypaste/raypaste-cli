package interactive

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/raypaste/raypaste-cli/internal/clipboard"
	"github.com/raypaste/raypaste-cli/internal/llm"
	"github.com/raypaste/raypaste-cli/internal/output"
)

// generateStreaming generates a streaming response using the LLM client.
func generateStreaming(ctx context.Context, input string, state *State, opts Options) error {
	systemPrompt, err := state.Store.Render(state.PromptName, state.Length, state.ProjCtx.Content)
	if err != nil {
		return fmt.Errorf("failed to render prompt: %w", err)
	}

	req, err := llm.BuildRequest(
		state.Model,
		systemPrompt,
		input,
		state.Length,
		opts.Temperature,
		true, // streaming enabled
		opts.Models,
	)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	// Reset last response
	state.LastResponse = ""
	var responseBuilder strings.Builder
	colorizer := output.NewStreamingColorizer()

	// Show progress indicator
	fmt.Fprintln(os.Stderr, output.GeneratingMessage(state.Model, string(state.Length), state.ProjCtx.Filename))

	// Stream response
	fmt.Println() // New line before output
	startTime := time.Now()
	usage, err := state.Client.StreamComplete(ctx, req, func(token string) error {
		colorizedToken := colorizer.ProcessToken(token)
		fmt.Print(colorizedToken)
		responseBuilder.WriteString(token)
		return nil
	})
	durationMs := time.Since(startTime).Milliseconds()

	if err != nil {
		fmt.Println() // Ensure newline after error
		return fmt.Errorf("streaming failed: %w", err)
	}

	if trailing := colorizer.Finalize(); trailing != "" {
		fmt.Print(trailing)
	}

	// Store response
	state.LastResponse = responseBuilder.String()

	fmt.Println() // New line after output
	fmt.Println() // Extra line for spacing

	if opts.AutoCopy {
		if warning := clipboard.CopyWithWarning(state.LastResponse); warning != "" {
			fmt.Fprintln(os.Stderr, warning)
		} else {
			fmt.Fprintln(os.Stderr, output.CopiedMessage())
		}
	}

	// Display token usage and completion time
	fmt.Fprintln(os.Stderr, output.TokenUsageMessage(usage.PromptTokens, usage.CompletionTokens, durationMs))

	return nil
}
