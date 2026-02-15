/*
Copyright © 2026 Raypaste
*/
package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/raypaste/raypaste-cli/internal/clipboard"
	"github.com/raypaste/raypaste-cli/internal/config"
	"github.com/raypaste/raypaste-cli/internal/llm"
	"github.com/raypaste/raypaste-cli/internal/output"
	"github.com/raypaste/raypaste-cli/internal/prompts"
	"github.com/raypaste/raypaste-cli/pkg/types"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

// interactiveCmd represents the interactive command
var interactiveCmd = &cobra.Command{
	Use:     "interactive",
	Aliases: []string{"i", "repl"},
	Short:   "Start an interactive REPL session",
	Long: output.Bold("Start an interactive REPL session") + output.Cyan(" with streaming output.") + `

The interactive mode provides a REPL (Read-Eval-Print Loop) where you can
continuously generate prompts with streaming output.

` + output.Bold("Slash commands:") + `
  ` + output.Green("/length") + `                       - Show current length and list of available lengths
  ` + output.Green("/length [name]") + `                - Change output length to provided length
	` + output.Green("/model") + `                        - Show current model and list of available models
  ` + output.Green("/model [name]") + `                 - Switch model to provided model
  ` + output.Green("/copy") + `                         - Copy last response to clipboard
	` + output.Green("/prompt") + `                       - Show current prompt and list of available prompts
  ` + output.Green("/prompt [name]") + `         			  - Switch prompt template to provided prompt
  ` + output.Green("/help") + `                         - Show help
  ` + output.Green("/quit") + ` or ` + output.Green("/exit") + `                - Exit REPL

` + output.Bold("Keyboard shortcuts:") + `
  ` + output.Yellow("Ctrl+C") + `  - Cancel current generation
  ` + output.Yellow("Ctrl+D") + `  - Exit REPL

` + output.Bold("Example:") + `
  raypaste interactive
  raypaste i`,
	RunE: runInteractive,
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}

type replState struct {
	model        string
	length       types.OutputLength
	promptName   string
	lastResponse string
	store        *prompts.Store
	client       *llm.Client
}

func runInteractive(cmd *cobra.Command, args []string) error {
	state := &replState{
		model:      modelFlag,
		promptName: promptFlag,
	}

	if state.model == "" {
		state.model = cfg.GetDefaultModel()
	}

	length, err := config.ValidateOutputLength(lengthFlag)
	if err != nil {
		return err
	}
	state.length = length

	state.store, err = prompts.NewStore()
	if err != nil {
		return fmt.Errorf("failed to load prompts: %w", err)
	}

	state.client = llm.NewClient(cfg.GetAPIKey())

	// Create readline instance
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "> ",
		HistoryFile:     getHistoryFile(),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return fmt.Errorf("failed to create readline: %w", err)
	}
	defer func() {
		_ = rl.Close()
	}()

	printWelcome(state)

	// Main REPL loop
	for {
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				continue
			} else if err == io.EOF {
				break
			}
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Handle slash commands
		if strings.HasPrefix(line, "/") {
			if shouldExit := handleSlashCommand(line, state); shouldExit {
				break
			}
			continue
		}

		// Generate response
		if err := generateStreaming(line, state); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}

	fmt.Println(output.Bold(output.Green("\nGoodbye!")))
	return nil
}

func printWelcome(state *replState) {
	for _, line := range formatWelcomeLines(state) {
		fmt.Print(line)
	}
}

func formatWelcomeLines(state *replState) []string {
	return []string{
		fmt.Sprintf("%s\n", output.Cyan("raypaste interactive mode")),
		fmt.Sprintf(
			"Model: %s | Length: %s | Prompt: %s\n",
			output.Bold(output.Blue(state.model)),
			output.Bold(output.Yellow(string(state.length))),
			output.Bold(output.Green(state.promptName)),
		),
		fmt.Sprintf(
			"Type %s for commands, %s to exit\n",
			output.Bold(output.Green("/help")),
			output.Bold(output.Red("/quit")),
		),
		"\n",
	}
}

func handleSlashCommand(line string, state *replState) bool {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return false
	}

	command := parts[0]
	args := parts[1:]

	switch command {
	case "/quit", "/exit":
		return true

	case "/help":
		printHelp()

	case "/length", "/l":
		if len(args) == 0 {
			fmt.Printf("Current length: %s\n", output.Bold(output.Yellow(string(state.length))))
			fmt.Printf("Usage: %s\n", output.Cyan("/length <short|medium|long>"))
			return false
		}
		length, err := config.ValidateOutputLength(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", output.Red(err.Error()))
			return false
		}
		state.length = length
		fmt.Printf("Length set to: %s\n", output.Bold(output.Yellow(string(length))))

	case "/model", "/m":
		if len(args) == 0 {
			fmt.Printf("Current model: %s\n", output.Bold(output.Blue(state.model)))
			models := config.ListModels(cfg.Models)
			coloredModels := make([]string, len(models))
			for i, m := range models {
				coloredModels[i] = output.Blue(m)
			}
			fmt.Printf("Available models: %s\n", strings.Join(coloredModels, ", "))
			fmt.Printf("Usage: %s\n", output.Cyan("/model <alias>"))
			return false
		}
		state.model = args[0]
		fmt.Printf("Model set to: %s\n", output.Bold(output.Blue(state.model)))

	case "/prompt", "/p":
		if len(args) == 0 {
			fmt.Printf("Current prompt: %s\n", output.Bold(output.Green(state.promptName)))
			prompts := state.store.List()
			coloredPrompts := make([]string, len(prompts))
			for i, p := range prompts {
				coloredPrompts[i] = output.Green(p)
			}
			fmt.Printf("Available prompts: %s\n", strings.Join(coloredPrompts, ", "))
			fmt.Printf("Usage: %s\n", output.Cyan("/prompt <name>"))
			return false
		}
		// Verify prompt exists
		if _, err := state.store.Get(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", output.Red(err.Error()))
			return false
		}
		state.promptName = args[0]
		fmt.Printf("Prompt set to: %s\n", output.Bold(output.Green(state.promptName)))

	case "/copy", "/c":
		if state.lastResponse == "" {
			fmt.Println(output.Yellow("No response to copy"))
			return false
		}
		if warning := clipboard.CopyWithWarning(state.lastResponse); warning != "" {
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
	fmt.Println("\nAvailable commands:")
	fmt.Printf("  %s                       - Show current length and list of available lengths\n", output.Cyan("/length"))
	fmt.Printf("  %s [name]                - Change output length to provided length\n", output.Cyan("/length"))
	fmt.Printf("  %s                        - Show current model and list of available models\n", output.Cyan("/model"))
	fmt.Printf("  %s [name]                 - Switch model to provided model\n", output.Cyan("/model"))
	fmt.Printf("  %s                         - Copy last response to clipboard\n", output.Cyan("/copy"))
	fmt.Printf("  %s                       - Show current prompt and list of available prompts\n", output.Cyan("/prompt"))
	fmt.Printf("  %s [name]                - Switch prompt template to provided prompt\n", output.Cyan("/prompt"))
	fmt.Printf("  %s                         - Show this help\n", output.Cyan("/help"))
	fmt.Printf("  %s                - Exit REPL\n", output.Red("/quit or /exit"))
	fmt.Println("\nKeyboard shortcuts:")
	fmt.Printf("  %s  - Cancel current generation\n", output.BoldYellow("Ctrl+C"))
	fmt.Printf("  %s  - Exit REPL\n", output.BoldRed("Ctrl+D"))
	fmt.Println()
}

func generateStreaming(input string, state *replState) error {
	systemPrompt, err := state.store.Render(state.promptName, state.length)
	if err != nil {
		return fmt.Errorf("failed to render prompt: %w", err)
	}

	req, err := llm.BuildRequest(
		state.model,
		systemPrompt,
		input,
		state.length,
		cfg.Temperature,
		true, // streaming enabled
		cfg.Models,
	)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	// Reset last response
	state.lastResponse = ""
	var responseBuilder strings.Builder
	colorizer := output.NewStreamingColorizer()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Show progress indicator
	fmt.Fprintln(os.Stderr, output.GeneratingMessage())

	// Stream response
	fmt.Println() // New line before output
	err = state.client.StreamComplete(ctx, req, func(token string) error {
		colorizedToken := colorizer.ProcessToken(token)
		fmt.Print(colorizedToken)
		responseBuilder.WriteString(token)
		return nil
	})

	if err != nil {
		fmt.Println() // Ensure newline after error
		return fmt.Errorf("streaming failed: %w", err)
	}

	if trailing := colorizer.Finalize(); trailing != "" {
		fmt.Print(trailing)
	}

	// Store response
	state.lastResponse = responseBuilder.String()

	fmt.Println() // New line after output
	fmt.Println() // Extra line for spacing

	// Auto-copy if enabled
	if cfg.AutoCopy {
		if warning := clipboard.CopyWithWarning(state.lastResponse); warning != "" {
			fmt.Fprintln(os.Stderr, warning)
		} else {
			fmt.Fprintln(os.Stderr, output.CopiedMessage())
		}
	}

	return nil
}

func getHistoryFile() string {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return ""
	}
	return configDir + "/history"
}
