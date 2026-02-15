/*
Copyright © 2026 Raypaste
*/
package prompts

import (
	"testing"

	"github.com/raypaste/raypaste-cli/pkg/types"
)

func TestNewStore(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	if store == nil {
		t.Fatal("NewStore() returned nil")
	}

	if len(store.prompts) == 0 {
		t.Error("NewStore() should load at least built-in prompts")
	}
}

func TestStoreGet(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	tests := []struct {
		name    string
		prompt  string
		wantErr bool
	}{
		{"built-in metaprompt", "metaprompt", false},
		{"non-existent", "nonexistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.Get(tt.prompt)
			if (err != nil) != tt.wantErr {
				t.Errorf("Store.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("Store.Get() returned nil prompt")
			}
		})
	}
}

func TestStoreRender(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	tests := []struct {
		name    string
		prompt  string
		length  types.OutputLength
		wantErr bool
	}{
		{"metaprompt short", "metaprompt", types.OutputLengthShort, false},
		{"metaprompt medium", "metaprompt", types.OutputLengthMedium, false},
		{"metaprompt long", "metaprompt", types.OutputLengthLong, false},
		{"non-existent prompt", "nonexistent", types.OutputLengthMedium, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.Render(tt.prompt, tt.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("Store.Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Error("Store.Render() returned empty string")
			}
			if !tt.wantErr && got == "{{.LengthDirective}}" {
				t.Error("Store.Render() did not replace template variable")
			}
		})
	}
}

func TestStoreList(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	list := store.List()
	if len(list) == 0 {
		t.Error("Store.List() returned empty list")
	}

	// Check that metaprompt is in the list
	hasMetaprompt := false
	for _, name := range list {
		if name == "metaprompt" {
			hasMetaprompt = true
			break
		}
	}
	if !hasMetaprompt {
		t.Error("Store.List() should include 'metaprompt'")
	}
}

func TestBulletListPrompt(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	tests := []struct {
		name    string
		length  types.OutputLength
		wantErr bool
	}{
		{"bulletlist short (supported)", types.OutputLengthShort, false},
		{"bulletlist medium (supported)", types.OutputLengthMedium, false},
		{"bulletlist long (not supported)", types.OutputLengthLong, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.Render("bulletlist", tt.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("Store.Render(bulletlist, %s) error = %v, wantErr %v", tt.length, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got == "" {
					t.Error("Store.Render() returned empty string")
				}
				if got == "{{.LengthDirective}}" {
					t.Error("Store.Render() did not replace template variable")
				}
			}
		})
	}
}

func TestPromptLengthRestriction(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// Test that bulletlist has only short and medium directives
	prompt, err := store.Get("bulletlist")
	if err != nil {
		t.Fatalf("Failed to get bulletlist prompt: %v", err)
	}

	if len(prompt.LengthDirectives) != 2 {
		t.Errorf("bulletlist should have exactly 2 length directives, got %d", len(prompt.LengthDirectives))
	}

	if _, ok := prompt.LengthDirectives[string(types.OutputLengthShort)]; !ok {
		t.Error("bulletlist should support short length")
	}

	if _, ok := prompt.LengthDirectives[string(types.OutputLengthMedium)]; !ok {
		t.Error("bulletlist should support medium length")
	}

	if _, ok := prompt.LengthDirectives[string(types.OutputLengthLong)]; ok {
		t.Error("bulletlist should NOT support long length")
	}
}
