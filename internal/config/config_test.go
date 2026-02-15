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

func TestConfigGetDefaultModel(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{"custom model", Config{DefaultModel: "custom-model"}, "custom-model"},
		{"empty falls back", Config{DefaultModel: ""}, "cerebras-llama-8b"},
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
