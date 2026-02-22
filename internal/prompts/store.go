/*
Copyright © 2026 Raypaste
*/
package prompts

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/raypaste/raypaste-cli/internal/config"
	"github.com/raypaste/raypaste-cli/internal/llm"
	"github.com/raypaste/raypaste-cli/internal/prompts/defaults"
	"github.com/raypaste/raypaste-cli/pkg/types"

	"gopkg.in/yaml.v3"
)

// Prompt represents a prompt template
type Prompt struct {
	Name             string            `yaml:"name"`
	Description      string            `yaml:"description"`
	System           string            `yaml:"system"`
	LengthDirectives map[string]string `yaml:"length_directives"`
}

// Store manages prompt templates
type Store struct {
	prompts map[string]*Prompt
}

// NewStore creates a new prompt store and loads all prompts
func NewStore() (*Store, error) {
	s := &Store{
		prompts: make(map[string]*Prompt),
	}

	// Load built-in prompts
	if err := s.loadBuiltInPrompts(); err != nil {
		return nil, fmt.Errorf("failed to load built-in prompts: %w", err)
	}

	// Load user prompts
	if err := s.loadUserPrompts(); err != nil {
		// Don't fail if user prompts can't be loaded, just warn
		fmt.Fprintf(os.Stderr, "Warning: failed to load user prompts: %v\n", err)
	}

	return s, nil
}

// loadBuiltInPrompts loads the default built-in prompts
func (s *Store) loadBuiltInPrompts() error {
	// Create default meta-prompt
	metaPrompt := &Prompt{
		Name:        defaults.MetaPromptName,
		Description: defaults.MetaPromptDescription,
		System:      defaults.MetaPromptTemplate,
		LengthDirectives: map[string]string{
			string(types.OutputLengthShort):  llm.LengthParams[types.OutputLengthShort].Directive,
			string(types.OutputLengthMedium): llm.LengthParams[types.OutputLengthMedium].Directive,
			string(types.OutputLengthLong):   llm.LengthParams[types.OutputLengthLong].Directive,
		},
	}

	// Create bulletlist prompt (only supports short and medium)
	bulletListPrompt := &Prompt{
		Name:        defaults.BulletListName,
		Description: defaults.BulletListDescription,
		System:      defaults.BulletListTemplate,
		LengthDirectives: map[string]string{
			string(types.OutputLengthShort):  llm.LengthParams[types.OutputLengthShort].Directive,
			string(types.OutputLengthMedium): llm.LengthParams[types.OutputLengthMedium].Directive,
			// Note: long mode intentionally not supported for bulletlist
		},
	}

	s.prompts[defaults.MetaPromptName] = metaPrompt
	s.prompts[defaults.BulletListName] = bulletListPrompt
	return nil
}

// loadUserPrompts loads user-defined prompts from ~/.raypaste/prompts/
func (s *Store) loadUserPrompts() error {
	promptsDir, err := config.GetPromptsDir()
	if err != nil {
		return err
	}

	// Check if directory exists
	if _, err := os.Stat(promptsDir); os.IsNotExist(err) {
		return nil // No user prompts yet
	}

	// Read all .yaml files
	files, err := filepath.Glob(filepath.Join(promptsDir, "*.yaml"))
	if err != nil {
		return fmt.Errorf("failed to list prompt files: %w", err)
	}

	// Also check for .yml files
	ymlFiles, err := filepath.Glob(filepath.Join(promptsDir, "*.yml"))
	if err != nil {
		return fmt.Errorf("failed to list prompt files: %w", err)
	}
	files = append(files, ymlFiles...)

	// Load each file
	for _, file := range files {
		if err := s.loadPromptFile(file); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load prompt file %s: %v\n", file, err)
		}
	}

	return nil
}

// loadPromptFile loads a single prompt file
func (s *Store) loadPromptFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var prompt Prompt
	if err := yaml.Unmarshal(data, &prompt); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	if prompt.Name == "" {
		return fmt.Errorf("prompt name is required")
	}

	s.prompts[prompt.Name] = &prompt
	return nil
}

// Get retrieves a prompt by name
func (s *Store) Get(name string) (*Prompt, error) {
	prompt, ok := s.prompts[name]
	if !ok {
		return nil, fmt.Errorf("prompt not found: %s", name)
	}
	return prompt, nil
}

// Render renders a prompt template with the given output length and project context.
// The context string is injected into the template via {{.Context}}.
func (s *Store) Render(name string, length types.OutputLength, context string) (string, error) {
	prompt, err := s.Get(name)
	if err != nil {
		return "", err
	}

	directive, ok := prompt.LengthDirectives[string(length)]
	if !ok {
		// Check if this prompt has any length directives defined
		if len(prompt.LengthDirectives) > 0 {
			// If the prompt has specific length directives but this length isn't supported,
			// return an error instead of falling back
			return "", fmt.Errorf("prompt '%s' does not support output length '%s'", name, length)
		}
		// Fall back to default directive for prompts without specific directives
		directive = llm.LengthParams[length].Directive
	}

	// Parse template
	tmpl, err := template.New("prompt").Parse(prompt.System)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Render template
	var buf bytes.Buffer
	data := map[string]string{
		"LengthDirective": directive,
		"Context":         context,
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return buf.String(), nil
}

// List returns a list of all available prompt names
func (s *Store) List() []string {
	names := make([]string, 0, len(s.prompts))
	for name := range s.prompts {
		names = append(names, name)
	}
	return names
}

// SavePrompt saves a custom prompt to the user's prompts directory
func (s *Store) SavePrompt(prompt *Prompt) error {
	if prompt.Name == "" {
		return fmt.Errorf("prompt name is required")
	}

	promptsDir, err := config.GetPromptsDir()
	if err != nil {
		return err
	}

	// Ensure the prompts directory exists
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		return fmt.Errorf("failed to create prompts directory: %w", err)
	}

	// Marshal the prompt to YAML
	data, err := yaml.Marshal(prompt)
	if err != nil {
		return fmt.Errorf("failed to marshal prompt to YAML: %w", err)
	}

	// Write to file
	filename := filepath.Join(promptsDir, prompt.Name+".yaml")
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write prompt file: %w", err)
	}

	// Add to the store's prompts map
	s.prompts[prompt.Name] = prompt

	return nil
}

// DeletePrompt removes a custom prompt from the user's prompts directory
func (s *Store) DeletePrompt(name string) error {
	// Check if prompt exists
	prompt, ok := s.prompts[name]
	if !ok {
		return fmt.Errorf("prompt not found: %s", name)
	}

	// Check if it's a built-in prompt (cannot delete built-ins)
	if prompt.Name == defaults.MetaPromptName || prompt.Name == defaults.BulletListName {
		return fmt.Errorf("cannot delete built-in prompt: %s", name)
	}

	promptsDir, err := config.GetPromptsDir()
	if err != nil {
		return err
	}

	// Try to delete the file (may not exist if loaded from elsewhere)
	filename := filepath.Join(promptsDir, name+".yaml")
	if _, err := os.Stat(filename); err == nil {
		if err := os.Remove(filename); err != nil {
			return fmt.Errorf("failed to delete prompt file: %w", err)
		}
	}

	// Also try .yml extension
	filenameYml := filepath.Join(promptsDir, name+".yml")
	if _, err := os.Stat(filenameYml); err == nil {
		if err := os.Remove(filenameYml); err != nil {
			return fmt.Errorf("failed to delete prompt file: %w", err)
		}
	}

	// Remove from the store's prompts map
	delete(s.prompts, name)

	return nil
}

// IsBuiltIn checks if a prompt is a built-in prompt
func (s *Store) IsBuiltIn(name string) bool {
	return name == defaults.MetaPromptName || name == defaults.BulletListName
}
