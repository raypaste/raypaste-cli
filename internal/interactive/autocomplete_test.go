package interactive

import (
	"strings"
	"testing"
)

func TestCompleteLineCommandPrefixMatching(t *testing.T) {
	commandNames := slashCommandAutocompleteNames(interactiveSlashCommands)
	modelNames := []string{"z-model", "A-model"}
	promptNames := []string{"beta", "Alpha"}

	tests := []struct {
		name            string
		input           string
		wantPrefix      string
		wantSuggestions []string
	}{
		{
			name:            "slash shows command suggestions",
			input:           "/",
			wantPrefix:      "/",
			wantSuggestions: []string{"/clear", "/length", "/model", "/copy", "/prompt", "/help", "/quit", "/exit"},
		},
		{
			name:            "prefix filters model command",
			input:           "/mo",
			wantPrefix:      "/mo",
			wantSuggestions: []string{"/model"},
		},
		{
			name:            "case-insensitive command matching",
			input:           "/PR",
			wantPrefix:      "/PR",
			wantSuggestions: []string{"/prompt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSuggestions, gotPrefix := completeLine(tt.input, commandNames, modelNames, promptNames)
			if gotPrefix != tt.wantPrefix {
				t.Fatalf("completeLine() prefix = %q, want %q", gotPrefix, tt.wantPrefix)
			}
			if strings.Join(gotSuggestions, ",") != strings.Join(tt.wantSuggestions, ",") {
				t.Fatalf(
					"completeLine() suggestions = %v, want %v",
					gotSuggestions,
					tt.wantSuggestions,
				)
			}
		})
	}
}

func TestCompleteLineModelSuggestions(t *testing.T) {
	commandNames := slashCommandAutocompleteNames(interactiveSlashCommands)
	modelNames := sortedCaseInsensitive([]string{"zeta-model", "Alpha-Model", "beta-model"})
	promptNames := []string{"metaprompt"}

	tests := []struct {
		name            string
		input           string
		wantPrefix      string
		wantSuggestions []string
	}{
		{
			name:            "model command with trailing space suggests all sorted models",
			input:           "/model ",
			wantPrefix:      "",
			wantSuggestions: []string{"Alpha-Model", "beta-model", "zeta-model"},
		},
		{
			name:            "model alias command supports completions",
			input:           "/m be",
			wantPrefix:      "be",
			wantSuggestions: []string{"beta-model"},
		},
		{
			name:            "case-insensitive model filtering",
			input:           "/model AL",
			wantPrefix:      "AL",
			wantSuggestions: []string{"Alpha-Model"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSuggestions, gotPrefix := completeLine(tt.input, commandNames, modelNames, promptNames)
			if gotPrefix != tt.wantPrefix {
				t.Fatalf("completeLine() prefix = %q, want %q", gotPrefix, tt.wantPrefix)
			}
			if strings.Join(gotSuggestions, ",") != strings.Join(tt.wantSuggestions, ",") {
				t.Fatalf(
					"completeLine() suggestions = %v, want %v",
					gotSuggestions,
					tt.wantSuggestions,
				)
			}
		})
	}
}

func TestCompleteLinePromptSuggestions(t *testing.T) {
	commandNames := slashCommandAutocompleteNames(interactiveSlashCommands)
	modelNames := []string{"model-a"}
	promptNames := sortedCaseInsensitive([]string{"zebra", "AlphaPrompt", "betaPrompt"})

	tests := []struct {
		name            string
		input           string
		wantPrefix      string
		wantSuggestions []string
	}{
		{
			name:            "prompt command with trailing space suggests all sorted prompts",
			input:           "/prompt ",
			wantPrefix:      "",
			wantSuggestions: []string{"AlphaPrompt", "betaPrompt", "zebra"},
		},
		{
			name:            "prompt alias command supports completions",
			input:           "/p be",
			wantPrefix:      "be",
			wantSuggestions: []string{"betaPrompt"},
		},
		{
			name:            "case-insensitive prompt filtering",
			input:           "/prompt AL",
			wantPrefix:      "AL",
			wantSuggestions: []string{"AlphaPrompt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSuggestions, gotPrefix := completeLine(tt.input, commandNames, modelNames, promptNames)
			if gotPrefix != tt.wantPrefix {
				t.Fatalf("completeLine() prefix = %q, want %q", gotPrefix, tt.wantPrefix)
			}
			if strings.Join(gotSuggestions, ",") != strings.Join(tt.wantSuggestions, ",") {
				t.Fatalf(
					"completeLine() suggestions = %v, want %v",
					gotSuggestions,
					tt.wantSuggestions,
				)
			}
		})
	}
}

func TestCompleteLineDeterministicOrdering(t *testing.T) {
	commandNames := slashCommandAutocompleteNames(interactiveSlashCommands)
	modelNames := sortedCaseInsensitive([]string{"z-model", "A-model", "b-model"})
	promptNames := sortedCaseInsensitive([]string{"z-prompt", "A-prompt", "b-prompt"})

	firstModels, _ := completeLine("/model ", commandNames, modelNames, promptNames)
	secondModels, _ := completeLine("/model ", commandNames, modelNames, promptNames)
	if strings.Join(firstModels, ",") != strings.Join(secondModels, ",") {
		t.Fatalf("model suggestions are not deterministic: %v vs %v", firstModels, secondModels)
	}

	firstPrompts, _ := completeLine("/prompt ", commandNames, modelNames, promptNames)
	secondPrompts, _ := completeLine("/prompt ", commandNames, modelNames, promptNames)
	if strings.Join(firstPrompts, ",") != strings.Join(secondPrompts, ",") {
		t.Fatalf("prompt suggestions are not deterministic: %v vs %v", firstPrompts, secondPrompts)
	}
}
