/*
Copyright © 2026 Raypaste
*/
package llm

import (
	"fmt"
	"strings"

	"github.com/raypaste/raypaste-cli/internal/config"
	"github.com/raypaste/raypaste-cli/pkg/types"
)

// LengthParams maps output lengths to their parameters
var LengthParams = map[types.OutputLength]types.LengthParams{
	types.OutputLengthShort: {
		MaxTokens: 550,
		Directive: "Keep the generated prompt concise — under 150 words. Focus on the core instruction only.",
	},
	types.OutputLengthMedium: {
		MaxTokens: 850,
		Directive: "Generate a moderately detailed prompt (~200-350 words) with context, constraints, and desired output format.",
	},
	types.OutputLengthLong: {
		MaxTokens: 1600,
		Directive: "Generate a comprehensive prompt (400-600+ words) including examples, edge cases, tone guidance, and detailed formatting instructions.",
	},
}

// BuildRequest builds a completion request with the given parameters
func BuildRequest(modelAlias, systemPrompt, userPrompt string, length types.OutputLength, temperature float64, stream bool, customModels map[string]config.Model) (types.CompletionRequest, error) {
	modelID, err := config.GetModelID(modelAlias, customModels)
	if err != nil {
		return types.CompletionRequest{}, fmt.Errorf("failed to resolve model: %w", err)
	}

	lengthParams, ok := LengthParams[length]
	if !ok {
		return types.CompletionRequest{}, fmt.Errorf("invalid output length: %s", length)
	}

	messages := []types.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: userPrompt,
		},
	}

	req := types.CompletionRequest{
		Model:       modelID,
		Messages:    messages,
		MaxTokens:   lengthParams.MaxTokens,
		Temperature: temperature,
		Stream:      stream,
	}

	// GPT-5 models account for reasoning tokens inside completion tokens.
	// Using max_completion_tokens and lower reasoning effort avoids empty
	// visible outputs caused by reasoning consuming the full budget.
	if isGPT5Model(modelID) {
		req.MaxTokens = 0
		req.MaxCompletionTokens = lengthParams.MaxTokens
		req.ReasoningEffort = "minimal"
	}

	return req, nil
}

func isGPT5Model(modelID string) bool {
	modelID = strings.ToLower(modelID)
	return strings.HasPrefix(modelID, "openai/gpt-5") || strings.HasPrefix(modelID, "gpt-5")
}

// GetLengthDirective returns the directive for a given output length
func GetLengthDirective(length types.OutputLength) (string, error) {
	params, ok := LengthParams[length]
	if !ok {
		return "", fmt.Errorf("invalid output length: %s", length)
	}
	return params.Directive, nil
}
