/*
Copyright © 2026 Raypaste
*/
package prompts

import (
	"testing"

	"raypaste-cli/pkg/types"
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
