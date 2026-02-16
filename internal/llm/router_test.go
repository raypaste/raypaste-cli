/*
Copyright © 2026 Raypaste
*/
package llm

import (
	"testing"

	"github.com/raypaste/raypaste-cli/internal/config"
	"github.com/raypaste/raypaste-cli/pkg/types"
)

func TestGetLengthDirective(t *testing.T) {
	tests := []struct {
		name    string
		length  types.OutputLength
		wantErr bool
	}{
		{"short", types.OutputLengthShort, false},
		{"medium", types.OutputLengthMedium, false},
		{"long", types.OutputLengthLong, false},
		{"invalid", types.OutputLength("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLengthDirective(tt.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLengthDirective() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Error("GetLengthDirective() returned empty directive")
			}
		})
	}
}

func TestBuildRequest(t *testing.T) {
	customModels := map[string]config.Model{
		"test-model": {
			ID:       "test/model",
			Provider: "test",
			Tier:     "fast",
		},
		"test-gpt5": {
			ID:       "openai/gpt-5-nano",
			Provider: "openai",
			Tier:     "fast",
		},
	}

	tests := []struct {
		name       string
		modelAlias string
		wantModel  string
		length     types.OutputLength
		wantMaxTok int
		wantMaxCmp int
		wantEffort string
		wantErr    bool
	}{
		{"short length", "test-model", "test/model", types.OutputLengthShort, 300, 0, "", false},
		{"medium length", "test-model", "test/model", types.OutputLengthMedium, 800, 0, "", false},
		{"long length", "test-model", "test/model", types.OutputLengthLong, 1500, 0, "", false},
		{"gpt5 medium length", "test-gpt5", "openai/gpt-5-nano", types.OutputLengthMedium, 0, 800, "minimal", false},
		{"invalid length", "test-model", "", types.OutputLength("invalid"), 0, 0, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := BuildRequest(
				tt.modelAlias,
				"system prompt",
				"user prompt",
				tt.length,
				0.7,
				false,
				customModels,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if req.MaxTokens != tt.wantMaxTok {
					t.Errorf("BuildRequest() MaxTokens = %v, want %v", req.MaxTokens, tt.wantMaxTok)
				}
				if req.MaxCompletionTokens != tt.wantMaxCmp {
					t.Errorf("BuildRequest() MaxCompletionTokens = %v, want %v", req.MaxCompletionTokens, tt.wantMaxCmp)
				}
				if req.ReasoningEffort != tt.wantEffort {
					t.Errorf("BuildRequest() ReasoningEffort = %q, want %q", req.ReasoningEffort, tt.wantEffort)
				}
				if req.Model != tt.wantModel {
					t.Errorf("BuildRequest() Model = %v, want %v", req.Model, tt.wantModel)
				}
				if len(req.Messages) != 2 {
					t.Errorf("BuildRequest() Messages length = %v, want 2", len(req.Messages))
				}
			}
		})
	}
}

func TestIsGPT5Model(t *testing.T) {
	tests := []struct {
		modelID string
		want    bool
	}{
		{"openai/gpt-5-nano", true},
		{"openai/gpt-5", true},
		{"gpt-5-nano", true},
		{"openai/gpt-4.1-nano", false},
		{"meta-llama/llama-3.1-8b-instruct", false},
	}

	for _, tt := range tests {
		t.Run(tt.modelID, func(t *testing.T) {
			if got := isGPT5Model(tt.modelID); got != tt.want {
				t.Errorf("isGPT5Model() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLengthParams(t *testing.T) {
	// Verify all expected lengths have parameters
	lengths := []types.OutputLength{
		types.OutputLengthShort,
		types.OutputLengthMedium,
		types.OutputLengthLong,
	}

	for _, length := range lengths {
		t.Run(string(length), func(t *testing.T) {
			params, ok := LengthParams[length]
			if !ok {
				t.Errorf("LengthParams missing entry for %s", length)
				return
			}
			if params.MaxTokens <= 0 {
				t.Errorf("LengthParams[%s].MaxTokens = %v, want > 0", length, params.MaxTokens)
			}
			if params.Directive == "" {
				t.Errorf("LengthParams[%s].Directive is empty", length)
			}
		})
	}
}
