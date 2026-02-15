/*
Copyright © 2026 Raypaste
*/
package llm

import (
	"fmt"

	"github.com/raypaste/raypaste-cli/internal/config"
	"github.com/raypaste/raypaste-cli/pkg/types"
)

// LengthParams maps output lengths to their parameters
var LengthParams = map[types.OutputLength]types.LengthParams{
	types.OutputLengthShort: {
		MaxTokens: 300,
		Directive: "Keep the generated prompt concise — under 100 words. Focus on the core instruction only.",
	},
	types.OutputLengthMedium: {
		MaxTokens: 800,
		Directive: "Generate a moderately detailed prompt (~150-300 words) with context, constraints, and desired output format.",
	},
	types.OutputLengthLong: {
		MaxTokens: 1500,
		Directive: "Generate a comprehensive prompt (300-500+ words) including examples, edge cases, tone guidance, and detailed formatting instructions.",
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

	return req, nil
}

// GetLengthDirective returns the directive for a given output length
func GetLengthDirective(length types.OutputLength) (string, error) {
	params, ok := LengthParams[length]
	if !ok {
		return "", fmt.Errorf("invalid output length: %s", length)
	}
	return params.Directive, nil
}
