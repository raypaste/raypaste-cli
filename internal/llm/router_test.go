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
	}

	tests := []struct {
		name       string
		modelAlias string
		length     types.OutputLength
		wantMaxTok int
		wantErr    bool
	}{
		{"short length", "test-model", types.OutputLengthShort, 300, false},
		{"medium length", "test-model", types.OutputLengthMedium, 800, false},
		{"long length", "test-model", types.OutputLengthLong, 1500, false},
		{"invalid length", "test-model", types.OutputLength("invalid"), 0, true},
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
				if req.Model != "test/model" {
					t.Errorf("BuildRequest() Model = %v, want test/model", req.Model)
				}
				if len(req.Messages) != 2 {
					t.Errorf("BuildRequest() Messages length = %v, want 2", len(req.Messages))
				}
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
