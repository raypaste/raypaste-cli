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
			got, err := store.Render(tt.prompt, tt.length, "")
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
			got, err := store.Render("bulletlist", tt.length, "")
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

func TestStoreIsBuiltIn(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	tests := []struct {
		name     string
		prompt   string
		expected bool
	}{
		{"metaprompt is built-in", "metaprompt", true},
		{"bulletlist is built-in", "bulletlist", true},
		{"non-existent is not built-in", "nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := store.IsBuiltIn(tt.prompt)
			if got != tt.expected {
				t.Errorf("Store.IsBuiltIn(%s) = %v, want %v", tt.prompt, got, tt.expected)
			}
		})
	}
}

func TestSavePromptValidation(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// Test that saving a prompt without a name fails
	prompt := &Prompt{
		Name:        "",
		Description: "Test description",
		System:      "Test system prompt",
	}

	err = store.SavePrompt(prompt)
	if err == nil {
		t.Error("SavePrompt() should fail when name is empty")
	}
}

func TestDeletePromptBuiltIn(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// Test that deleting a built-in prompt fails
	err = store.DeletePrompt("metaprompt")
	if err == nil {
		t.Error("DeletePrompt(metaprompt) should fail for built-in prompts")
	}

	err = store.DeletePrompt("bulletlist")
	if err == nil {
		t.Error("DeletePrompt(bulletlist) should fail for built-in prompts")
	}
}

func TestDeletePromptNonExistent(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// Test that deleting a non-existent prompt fails
	err = store.DeletePrompt("nonexistent-prompt")
	if err == nil {
		t.Error("DeletePrompt(nonexistent-prompt) should fail")
	}
}

func TestIsNumericDirective(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"300", true},
		{"1500", true},
		{"0", true},
		{"", false},
		{"Keep it short", false},
		{"300 tokens", false},
		{"abc", false},
		{"3.14", false},
		{"-100", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isNumericDirective(tt.input); got != tt.want {
				t.Errorf("isNumericDirective(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetMaxTokensOverride(t *testing.T) {
	store := &Store{prompts: make(map[string]*Prompt)}
	store.prompts["numeric-prompt"] = &Prompt{
		Name:   "numeric-prompt",
		System: "test {{.LengthDirective}}",
		LengthDirectives: map[string]string{
			"short":  "200",
			"medium": "500",
			"long":   "Generate a comprehensive prompt",
		},
	}
	store.prompts["text-prompt"] = &Prompt{
		Name:   "text-prompt",
		System: "test {{.LengthDirective}}",
		LengthDirectives: map[string]string{
			"short": "Be concise",
		},
	}

	tests := []struct {
		name   string
		prompt string
		length types.OutputLength
		want   int
	}{
		{"numeric short", "numeric-prompt", types.OutputLengthShort, 200},
		{"numeric medium", "numeric-prompt", types.OutputLengthMedium, 500},
		{"text long", "numeric-prompt", types.OutputLengthLong, 0},
		{"text directive", "text-prompt", types.OutputLengthShort, 0},
		{"nonexistent prompt", "nonexistent", types.OutputLengthShort, 0},
		{"unsupported length", "text-prompt", types.OutputLengthMedium, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := store.GetMaxTokensOverride(tt.prompt, tt.length)
			if got != tt.want {
				t.Errorf("GetMaxTokensOverride(%q, %q) = %v, want %v", tt.prompt, tt.length, got, tt.want)
			}
		})
	}
}

func TestRenderNumericDirectiveInjectsEmpty(t *testing.T) {
	store := &Store{prompts: make(map[string]*Prompt)}
	store.prompts["sql"] = &Prompt{
		Name:   "sql",
		System: "Act as SQL expert. {{.LengthDirective}} Output SQL only.",
		LengthDirectives: map[string]string{
			"short":  "200",
			"medium": "500",
		},
	}

	rendered, err := store.Render("sql", types.OutputLengthShort, "")
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	// Numeric directives must NOT appear in the rendered text
	if rendered != "Act as SQL expert.  Output SQL only." {
		t.Errorf("Render() with numeric directive should inject empty string, got: %q", rendered)
	}
}
