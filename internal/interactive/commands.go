package interactive

import (
	"fmt"
	"os"
	"strings"

	"github.com/raypaste/raypaste-cli/internal/clipboard"
	"github.com/raypaste/raypaste-cli/internal/config"
	"github.com/raypaste/raypaste-cli/internal/output"
)

type slashCommandHelpEntry struct {
	Usage       string
	Description string
}

type slashCommandSpec struct {
	Primary      string
	Aliases      []string
	HelpEntries  []slashCommandHelpEntry
	Autocomplete []string
}

var interactiveSlashCommands = []slashCommandSpec{
	{
		Primary: "/clear",
		HelpEntries: []slashCommandHelpEntry{
			{Usage: "/clear", Description: "Clear the screen"},
		},
	},
	{
		Primary: "/length",
		Aliases: []string{"/l"},
		HelpEntries: []slashCommandHelpEntry{
			{Usage: "/length", Description: "Show current length and list of available lengths"},
			{Usage: "/length [name]", Description: "Change output length to provided length"},
		},
	},
	{
		Primary: "/model",
		Aliases: []string{"/m"},
		HelpEntries: []slashCommandHelpEntry{
			{Usage: "/model", Description: "Show current model and list of available models"},
			{Usage: "/model [name]", Description: "Switch model to provided model"},
		},
	},
	{
		Primary: "/copy",
		Aliases: []string{"/c"},
		HelpEntries: []slashCommandHelpEntry{
			{Usage: "/copy", Description: "Copy last response to clipboard"},
		},
	},
	{
		Primary: "/prompt",
		Aliases: []string{"/p"},
		HelpEntries: []slashCommandHelpEntry{
			{Usage: "/prompt", Description: "Show current prompt and list of available prompts"},
			{Usage: "/prompt [name]", Description: "Switch prompt template to provided prompt"},
		},
	},
	{
		Primary: "/help",
		HelpEntries: []slashCommandHelpEntry{
			{Usage: "/help", Description: "Show this help"},
		},
	},
	{
		Primary: "/quit",
		Aliases: []string{"/exit"},
		HelpEntries: []slashCommandHelpEntry{
			{Usage: "/quit or /exit", Description: "Exit REPL"},
		},
		Autocomplete: []string{"/quit", "/exit"},
	},
}

var slashCommandLookup = buildSlashCommandLookup(interactiveSlashCommands)

// handleSlashCommand processes a slash command and returns true if the REPL should exit.
func handleSlashCommand(line string, state *State, models map[string]config.Model) bool {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return false
	}

	command := normalizeSlashCommand(parts[0])
	args := parts[1:]

	switch command {
	case "/quit":
		return true

	case "/clear":
		clearScreen()
		printWelcome(state)

	case "/help":
		printHelp()

	case "/length":
		if len(args) == 0 {
			fmt.Printf("Current length: %s\n", output.Bold(output.Yellow(string(state.Length))))
			fmt.Printf("Usage: %s\n", output.Cyan("/length <short|medium|long>"))
			return false
		}
		length, err := config.ValidateOutputLength(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", output.Red(err.Error()))
			return false
		}
		state.Length = length
		fmt.Printf("Length set to: %s\n", output.Bold(output.Yellow(string(length))))

	case "/model":
		if len(args) == 0 {
			fmt.Printf("Current model: %s\n", output.Bold(output.Blue(state.Model)))
			availableModels := config.ListModels(models)
			coloredModels := make([]string, len(availableModels))
			for i, m := range availableModels {
				coloredModels[i] = output.Blue(m)
			}
			fmt.Printf("Available models: %s\n", strings.Join(coloredModels, ", "))
			fmt.Printf("Usage: %s\n", output.Cyan("/model <alias>"))
			return false
		}
		state.Model = args[0]
		fmt.Printf("Model set to: %s\n", output.Bold(output.Blue(state.Model)))

	case "/prompt":
		if len(args) == 0 {
			fmt.Printf("Current prompt: %s\n", output.Bold(output.Green(state.PromptName)))
			prompts := state.Store.List()
			coloredPrompts := make([]string, len(prompts))
			for i, p := range prompts {
				coloredPrompts[i] = output.Green(p)
			}
			fmt.Printf("Available prompts: %s\n", strings.Join(coloredPrompts, ", "))
			fmt.Printf("Usage: %s\n", output.Cyan("/prompt <name>"))
			return false
		}
		// Verify prompt exists
		if _, err := state.Store.Get(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", output.Red(err.Error()))
			return false
		}
		state.PromptName = args[0]
		fmt.Printf("Prompt set to: %s\n", output.Bold(output.Green(state.PromptName)))

	case "/copy":
		if state.LastResponse == "" {
			fmt.Println(output.Yellow("No response to copy"))
			return false
		}
		if warning := clipboard.CopyWithWarning(state.LastResponse); warning != "" {
			fmt.Fprintln(os.Stderr, output.Yellow(warning))
		} else {
			fmt.Println(output.CopiedMessage())
		}

	default:
		fmt.Printf("Unknown command: %s (type %s for help)\n", output.Red(command), output.Bold(output.Green("/help")))
	}

	return false
}

func printHelp() {
	// Calculate max usage length for description text right-alignment
	maxUsageLen := 0
	for _, command := range interactiveSlashCommands {
		for _, entry := range command.HelpEntries {
			if len(entry.Usage) > maxUsageLen {
				maxUsageLen = len(entry.Usage)
			}
		}
	}

	fmt.Println(output.Bold("Available commands:"))
	fmt.Println()
	for _, command := range interactiveSlashCommands {
		for _, entry := range command.HelpEntries {
			padding := maxUsageLen - len(entry.Usage)
			fmt.Printf("  %s%s - %s\n", (output.Cyan(entry.Usage)), strings.Repeat(" ", padding), entry.Description)
		}
	}
	fmt.Println("\nKeyboard shortcuts:")
	fmt.Printf("  %s  - Cancel current generation\n", output.BoldYellow("Ctrl+C"))
	fmt.Printf("  %s  - Exit REPL\n", output.BoldRed("Ctrl+D"))
	fmt.Println()
}

func buildSlashCommandLookup(commands []slashCommandSpec) map[string]string {
	lookup := make(map[string]string)
	for _, command := range commands {
		primary := strings.ToLower(command.Primary)
		lookup[primary] = command.Primary
		for _, alias := range command.Aliases {
			lookup[strings.ToLower(alias)] = command.Primary
		}
	}
	return lookup
}

func normalizeSlashCommand(command string) string {
	normalized := strings.ToLower(command)
	if primary, ok := slashCommandLookup[normalized]; ok {
		return primary
	}
	return normalized
}
