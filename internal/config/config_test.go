/*
Copyright © 2026 Raypaste
*/
package config

import (
	"os"
	"testing"

	"github.com/raypaste/raypaste-cli/pkg/types"
)

func TestValidateOutputLength(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    types.OutputLength
		wantErr bool
	}{
		{"valid short", "short", types.OutputLengthShort, false},
		{"valid medium", "medium", types.OutputLengthMedium, false},
		{"valid long", "long", types.OutputLengthLong, false},
		{"invalid", "invalid", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateOutputLength(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOutputLength() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateOutputLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigGetAPIKey(t *testing.T) {
	tests := []struct {
		name   string
		cfg    Config
		envKey string
		want   string
	}{
		{"from config", Config{APIKey: "config-key"}, "", "config-key"},
		{"from env", Config{APIKey: ""}, "env-key", "env-key"},
		{"config priority", Config{APIKey: "config-key"}, "env-key", "config-key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envKey != "" {
				if err := os.Setenv("RAYPASTE_API_KEY", tt.envKey); err != nil {
					t.Fatalf("Failed to set env var: %v", err)
				}
				defer func() {
					_ = os.Unsetenv("RAYPASTE_API_KEY")
				}()
			}

			got := tt.cfg.GetAPIKey()
			if got != tt.want {
				t.Errorf("GetAPIKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveProviderKey(t *testing.T) {
	// Clear env vars that could interfere
	for _, env := range []string{"OPENROUTER_API_KEY", "CEREBRAS_API_KEY", "RAYPASTE_API_KEY"} {
		t.Setenv(env, "")
	}

	tests := []struct {
		name         string
		cfg          Config
		modelAlias   string
		wantProvider string
		wantKey      string
		wantErr      bool
	}{
		{
			name:         "cerebras model with cerebras key goes direct",
			cfg:          Config{CerebrasAPIKey: "csk-123", Models: map[string]Model{}},
			modelAlias:   "cerebras-llama-8b",
			wantProvider: "cerebras",
			wantKey:      "csk-123",
		},
		{
			name:         "cerebras model with only openrouter key falls back to openrouter",
			cfg:          Config{OpenRouterAPIKey: "sk-or-abc", Models: map[string]Model{}},
			modelAlias:   "cerebras-llama-8b",
			wantProvider: "openrouter",
			wantKey:      "sk-or-abc",
		},
		{
			name:         "cerebras model with both keys prefers direct",
			cfg:          Config{CerebrasAPIKey: "csk-123", OpenRouterAPIKey: "sk-or-abc", Models: map[string]Model{}},
			modelAlias:   "cerebras-llama-8b",
			wantProvider: "cerebras",
			wantKey:      "csk-123",
		},
		{
			name:         "openai model with openrouter key uses openrouter",
			cfg:          Config{OpenRouterAPIKey: "sk-or-abc", Models: map[string]Model{}},
			modelAlias:   "openai-gpt5-nano",
			wantProvider: "openrouter",
			wantKey:      "sk-or-abc",
		},
		{
			name:       "openai model with only cerebras key errors",
			cfg:        Config{CerebrasAPIKey: "csk-123", Models: map[string]Model{}},
			modelAlias: "openai-gpt5-nano",
			wantErr:    true,
		},
		{
			name:         "legacy api_key with sk-or prefix falls back to openrouter",
			cfg:          Config{APIKey: "sk-or-legacy", Models: map[string]Model{}},
			modelAlias:   "cerebras-llama-8b",
			wantProvider: "openrouter",
			wantKey:      "sk-or-legacy",
		},
		{
			name:       "legacy api_key without sk-or prefix does not fall back",
			cfg:        Config{APIKey: "some-other-key", Models: map[string]Model{}},
			modelAlias: "cerebras-llama-8b",
			wantErr:    true,
		},
		{
			name:       "no keys configured errors",
			cfg:        Config{Models: map[string]Model{}},
			modelAlias: "cerebras-llama-8b",
			wantErr:    true,
		},
		{
			name:         "unknown provider model with openrouter key uses openrouter",
			cfg:          Config{OpenRouterAPIKey: "sk-or-abc", Models: map[string]Model{}},
			modelAlias:   "some-provider/some-model",
			wantProvider: "openrouter",
			wantKey:      "sk-or-abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pk, err := tt.cfg.ResolveProviderKey(tt.modelAlias)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveProviderKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if pk.Provider != tt.wantProvider {
				t.Errorf("ResolveProviderKey() Provider = %q, want %q", pk.Provider, tt.wantProvider)
			}
			if pk.APIKey != tt.wantKey {
				t.Errorf("ResolveProviderKey() APIKey = %q, want %q", pk.APIKey, tt.wantKey)
			}
		})
	}
}

func TestHasAnyAPIKey(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want bool
	}{
		{"no keys", Config{}, false},
		{"openrouter key", Config{OpenRouterAPIKey: "sk-or-abc"}, true},
		{"cerebras key", Config{CerebrasAPIKey: "csk-123"}, true},
		{"legacy key", Config{APIKey: "some-key"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars
			for _, env := range []string{"OPENROUTER_API_KEY", "CEREBRAS_API_KEY", "RAYPASTE_API_KEY"} {
				t.Setenv(env, "")
			}
			if got := tt.cfg.HasAnyAPIKey(); got != tt.want {
				t.Errorf("HasAnyAPIKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOpenRouterAPIKey(t *testing.T) {
	tests := []struct {
		name   string
		cfg    Config
		envKey string
		want   string
	}{
		{"from config", Config{OpenRouterAPIKey: "config-key"}, "", "config-key"},
		{"from env", Config{}, "env-key", "env-key"},
		{"config priority", Config{OpenRouterAPIKey: "config-key"}, "env-key", "config-key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("OPENROUTER_API_KEY", tt.envKey)
			got := tt.cfg.GetOpenRouterAPIKey()
			if got != tt.want {
				t.Errorf("GetOpenRouterAPIKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCerebrasAPIKey(t *testing.T) {
	tests := []struct {
		name   string
		cfg    Config
		envKey string
		want   string
	}{
		{"from config", Config{CerebrasAPIKey: "config-key"}, "", "config-key"},
		{"from env", Config{}, "env-key", "env-key"},
		{"config priority", Config{CerebrasAPIKey: "config-key"}, "env-key", "config-key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CEREBRAS_API_KEY", tt.envKey)
			got := tt.cfg.GetCerebrasAPIKey()
			if got != tt.want {
				t.Errorf("GetCerebrasAPIKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigGetDefaultModel(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{"custom model", Config{DefaultModel: "custom-model"}, "custom-model"},
		{"empty falls back", Config{DefaultModel: ""}, "cerebras-gpt-oss-120b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.GetDefaultModel()
			if got != tt.want {
				t.Errorf("GetDefaultModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigGetDefaultLength(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want types.OutputLength
	}{
		{"custom length", Config{DefaultLength: types.OutputLengthShort}, types.OutputLengthShort},
		{"empty falls back", Config{DefaultLength: ""}, types.OutputLengthMedium},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.GetDefaultLength()
			if got != tt.want {
				t.Errorf("GetDefaultLength() = %v, want %v", got, tt.want)
			}
		})
	}
}
