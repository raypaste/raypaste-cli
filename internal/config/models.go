/*
Copyright © 2026 Raypaste
*/
package config

import "fmt"

// Model represents an LLM model configuration
type Model struct {
	ID       string `yaml:"id" mapstructure:"id"`
	Provider string `yaml:"provider" mapstructure:"provider"`
	Tier     string `yaml:"tier" mapstructure:"tier"`
}

// DefaultModels contains the built-in model registry
var DefaultModels = map[string]Model{
	"cerebras-llama-8b": {
		ID:       "meta-llama/llama-3.1-8b-instruct",
		Provider: "cerebras",
		Tier:     "fast",
	},
	"cerebras-gpt-oss-120b": {
		ID:       "openai/gpt-oss-120b",
		Provider: "cerebras",
		Tier:     "balanced",
	},
	"openai-gpt5-nano": {
		ID:       "openai/gpt-5-nano",
		Provider: "openai",
		Tier:     "fast",
	},
}

// ResolveModel resolves a model alias to a Model struct
// If the alias is not found in the registry, it treats it as a direct OpenRouter model ID
func ResolveModel(alias string, customModels map[string]Model) (Model, error) {
	// Check custom models first
	if model, ok := customModels[alias]; ok {
		return model, nil
	}

	// Check default models
	if model, ok := DefaultModels[alias]; ok {
		return model, nil
	}

	// If not found, treat as direct OpenRouter model ID
	return Model{
		ID:       alias,
		Provider: "unknown",
		Tier:     "unknown",
	}, nil
}

// ListModels returns a list of all available model aliases
func ListModels(customModels map[string]Model) []string {
	aliases := make([]string, 0, len(DefaultModels)+len(customModels))

	// Add custom models first
	for alias := range customModels {
		aliases = append(aliases, alias)
	}

	// Add default models
	for alias := range DefaultModels {
		// Skip if already in custom models
		if _, ok := customModels[alias]; !ok {
			aliases = append(aliases, alias)
		}
	}

	return aliases
}

// GetModelID returns the OpenRouter model ID for a given alias
func GetModelID(alias string, customModels map[string]Model) (string, error) {
	model, err := ResolveModel(alias, customModels)
	if err != nil {
		return "", err
	}

	if model.ID == "" {
		return "", fmt.Errorf("model %s has no ID", alias)
	}

	return model.ID, nil
}
