/*
Copyright © 2026 Raypaste
*/
package config

import (
	"testing"
)

func TestResolveModel(t *testing.T) {
	customModels := map[string]Model{
		"custom": {
			ID:       "custom/model",
			Provider: "custom",
			Tier:     "fast",
		},
	}

	tests := []struct {
		name    string
		alias   string
		want    string
		wantErr bool
	}{
		{"default model", "cerebras-llama-8b", "meta-llama/llama-3.1-8b-instruct", false},
		{"default openai model", "openai-gpt5-nano", "openai/gpt-5-nano", false},
		{"custom model", "custom", "custom/model", false},
		{"direct ID", "provider/direct-model", "provider/direct-model", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveModel(tt.alias, customModels)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.ID != tt.want {
				t.Errorf("ResolveModel() ID = %v, want %v", got.ID, tt.want)
			}
		})
	}
}

func TestGetModelID(t *testing.T) {
	customModels := map[string]Model{
		"custom": {
			ID:       "custom/model",
			Provider: "custom",
			Tier:     "fast",
		},
	}

	tests := []struct {
		name    string
		alias   string
		want    string
		wantErr bool
	}{
		{"default model", "cerebras-llama-8b", "meta-llama/llama-3.1-8b-instruct", false},
		{"default openai model", "openai-gpt5-nano", "openai/gpt-5-nano", false},
		{"custom model", "custom", "custom/model", false},
		{"direct ID", "provider/model", "provider/model", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetModelID(tt.alias, customModels)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetModelID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetModelID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListModels(t *testing.T) {
	customModels := map[string]Model{
		"custom1": {ID: "custom/model1"},
		"custom2": {ID: "custom/model2"},
	}

	aliases := ListModels(customModels)

	// Should have custom models + default models
	expectedMin := len(customModels) + len(DefaultModels)
	if len(aliases) < expectedMin {
		t.Errorf("ListModels() returned %d aliases, expected at least %d", len(aliases), expectedMin)
	}

	// Check that custom models are included
	hasCustom1 := false
	for _, alias := range aliases {
		if alias == "custom1" {
			hasCustom1 = true
			break
		}
	}
	if !hasCustom1 {
		t.Error("ListModels() should include custom model 'custom1'")
	}
}
