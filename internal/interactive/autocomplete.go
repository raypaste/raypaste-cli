package interactive

import (
	"sort"
	"strings"

	"github.com/raypaste/raypaste-cli/internal/config"
	"github.com/raypaste/raypaste-cli/internal/output"

	"github.com/chzyer/readline"
)

type autoCompleter struct {
	commandNames []string
	modelNames   func() []string
	promptNames  func() []string
}

func newAutoCompleter(state *State, opts Options) readline.AutoCompleter {
	return &autoCompleter{
		commandNames: slashCommandAutocompleteNames(interactiveSlashCommands),
		modelNames: func() []string {
			return sortedCaseInsensitive(config.ListModels(opts.Models))
		},
		promptNames: func() []string {
			return sortedCaseInsensitive(state.Store.List())
		},
	}
}

func (c *autoCompleter) Do(line []rune, pos int) ([][]rune, int) {
	if pos > len(line) {
		pos = len(line)
	}
	input := string(line[:pos])

	candidates, replacePrefix := completeLine(
		input,
		c.commandNames,
		c.modelNames(),
		c.promptNames(),
	)
	if len(candidates) == 0 {
		return nil, 0
	}

	offset := len([]rune(replacePrefix))
	newLine := make([][]rune, 0, len(candidates))
	for _, candidate := range candidates {
		candidateRunes := []rune(candidate)
		if offset > len(candidateRunes) {
			continue
		}
		newLine = append(newLine, candidateRunes[offset:])
	}

	return newLine, offset
}

func completeLine(input string, commandNames, modelNames, promptNames []string) ([]string, string) {
	trimmedLeft := strings.TrimLeft(input, " \t")
	if !strings.HasPrefix(trimmedLeft, "/") {
		return nil, ""
	}

	fields := strings.Fields(trimmedLeft)
	if len(fields) == 0 {
		return nil, ""
	}

	typedCommand := fields[0]
	command := normalizeSlashCommand(typedCommand)
	switch command {
	case "/model":
		if commandHasArguments(trimmedLeft) {
			prefix := argumentPrefix(trimmedLeft)
			return filterByPrefixCaseInsensitive(modelNames, prefix), prefix
		}
	case "/prompt":
		if commandHasArguments(trimmedLeft) {
			prefix := argumentPrefix(trimmedLeft)
			return filterByPrefixCaseInsensitive(promptNames, prefix), prefix
		}
	}

	return filterByPrefixCaseInsensitive(commandNames, typedCommand), typedCommand
}

func commandHasArguments(input string) bool {
	return strings.ContainsAny(input, " \t")
}

func argumentPrefix(input string) string {
	if strings.HasSuffix(input, " ") || strings.HasSuffix(input, "\t") {
		return ""
	}
	fields := strings.Fields(input)
	if len(fields) < 2 {
		return ""
	}
	return fields[len(fields)-1]
}

func filterByPrefixCaseInsensitive(candidates []string, prefix string) []string {
	normalizedPrefix := strings.ToLower(prefix)
	filtered := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if strings.HasPrefix(strings.ToLower(candidate), normalizedPrefix) {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

func sortedCaseInsensitive(values []string) []string {
	sorted := append([]string(nil), values...)
	sort.SliceStable(sorted, func(i, j int) bool {
		left := strings.ToLower(sorted[i])
		right := strings.ToLower(sorted[j])
		if left == right {
			return sorted[i] < sorted[j]
		}
		return left < right
	})
	return sorted
}

func slashCommandAutocompleteNames(commands []slashCommandSpec) []string {
	names := make([]string, 0, len(commands))
	for _, command := range commands {
		if len(command.Autocomplete) > 0 {
			names = append(names, command.Autocomplete...)
			continue
		}
		names = append(names, command.Primary)
	}
	return names
}

// suggestionPainter renders a dimmed completion preview after the cursor.
// It uses the same autocompleter as Tab completion for consistency.
type suggestionPainter struct {
	ac readline.AutoCompleter
}

func newSuggestionPainter(ac readline.AutoCompleter) readline.Painter {
	return &suggestionPainter{ac: ac}
}

func (p *suggestionPainter) Paint(line []rune, pos int) []rune {
	// Only show preview after typing at least 2 characters (e.g., "/h" not just "/")
	if len(line) < 2 {
		return line
	}
	suffixes, _ := p.ac.Do(line, pos)
	if len(suffixes) == 0 {
		return line
	}
	suffix := string(suffixes[0])
	if suffix == "" {
		return line
	}
	preview := output.SuggestionPreview(suffix)
	return append(line, []rune(preview)...)
}
