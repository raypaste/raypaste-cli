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

	"raypaste-cli/internal/clipboard"
	"raypaste-cli/internal/config"
	"raypaste-cli/internal/llm"
	"raypaste-cli/internal/prompts"
	"raypaste-cli/pkg/types"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

// interactiveCmd represents the interactive command
var interactiveCmd = &cobra.Command{
	Use:     "interactive",
	Aliases: []string{"i", "repl"},
	Short:   "Start an interactive REPL session",
	Long: `Start an interactive REPL session with streaming output.

The interactive mode provides a REPL (Read-Eval-Print Loop) where you can
continuously generate prompts with streaming output.

Slash commands:
  /length <short|medium|long>  - Change output length
  /model <alias>               - Switch model
  /copy                        - Copy last response to clipboard
  /prompt <name>               - Switch prompt template
  /help                        - Show help
  /quit or /exit               - Exit REPL

Keyboard shortcuts:
  Ctrl+C  - Cancel current generation
  Ctrl+D  - Exit REPL

Example:
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

	fmt.Println("\nGoodbye!")
	return nil
}

func printWelcome(state *replState) {
	fmt.Println("raypaste interactive mode")
	fmt.Printf("Model: %s | Length: %s | Prompt: %s\n", state.model, state.length, state.promptName)
	fmt.Println("Type /help for commands, /quit to exit")
	fmt.Println()
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
			fmt.Printf("Current length: %s\n", state.length)
			fmt.Println("Usage: /length <short|medium|long>")
			return false
		}
		length, err := config.ValidateOutputLength(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return false
		}
		state.length = length
		fmt.Printf("Length set to: %s\n", length)

	case "/model", "/m":
		if len(args) == 0 {
			fmt.Printf("Current model: %s\n", state.model)
			fmt.Println("Usage: /model <alias>")
			return false
		}
		state.model = args[0]
		fmt.Printf("Model set to: %s\n", state.model)

	case "/prompt", "/p":
		if len(args) == 0 {
			fmt.Printf("Current prompt: %s\n", state.promptName)
			fmt.Println("Available prompts:", strings.Join(state.store.List(), ", "))
			fmt.Println("Usage: /prompt <name>")
			return false
		}
		// Verify prompt exists
		if _, err := state.store.Get(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return false
		}
		state.promptName = args[0]
		fmt.Printf("Prompt set to: %s\n", state.promptName)

	case "/copy", "/c":
		if state.lastResponse == "" {
			fmt.Println("No response to copy")
			return false
		}
		if warning := clipboard.CopyWithWarning(state.lastResponse); warning != "" {
			fmt.Fprintln(os.Stderr, warning)
		} else {
			fmt.Println("✓ Copied to clipboard")
		}

	default:
		fmt.Printf("Unknown command: %s (type /help for help)\n", command)
	}

	return false
}

func printHelp() {
	fmt.Println("\nAvailable commands:")
	fmt.Println("  /length <short|medium|long>  - Change output length")
	fmt.Println("  /model <alias>               - Switch model")
	fmt.Println("  /prompt <name>               - Switch prompt template")
	fmt.Println("  /copy                        - Copy last response to clipboard")
	fmt.Println("  /help                        - Show this help")
	fmt.Println("  /quit or /exit               - Exit REPL")
	fmt.Println("\nKeyboard shortcuts:")
	fmt.Println("  Ctrl+C  - Cancel current generation")
	fmt.Println("  Ctrl+D  - Exit REPL")
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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Stream response
	fmt.Println() // New line before output
	err = state.client.StreamComplete(ctx, req, func(token string) error {
		fmt.Print(token)
		responseBuilder.WriteString(token)
		return nil
	})

	if err != nil {
		fmt.Println() // Ensure newline after error
		return fmt.Errorf("streaming failed: %w", err)
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
			fmt.Fprintln(os.Stderr, "✓ Copied to clipboard")
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
