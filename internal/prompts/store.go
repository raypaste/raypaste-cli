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

	"raypaste-cli/internal/config"
	"raypaste-cli/internal/llm"
	"raypaste-cli/internal/prompts/defaults"
	"raypaste-cli/pkg/types"

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

	s.prompts[defaults.MetaPromptName] = metaPrompt
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

// Render renders a prompt template with the given output length
func (s *Store) Render(name string, length types.OutputLength) (string, error) {
	prompt, err := s.Get(name)
	if err != nil {
		return "", err
	}

	directive, ok := prompt.LengthDirectives[string(length)]
	if !ok {
		// Fall back to default directive
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
