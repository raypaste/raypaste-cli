package interactive

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/raypaste/raypaste-cli/internal/config"
	"github.com/raypaste/raypaste-cli/internal/llm"
	"github.com/raypaste/raypaste-cli/internal/output"
	"github.com/raypaste/raypaste-cli/internal/projectcontext"
	"github.com/raypaste/raypaste-cli/internal/prompts"
	"github.com/raypaste/raypaste-cli/pkg/types"

	"github.com/chzyer/readline"
)

// State holds the REPL session state.
type State struct {
	Model        string
	Length       types.OutputLength
	PromptName   string
	LastResponse string
	ProjCtx      projectcontext.Result
	Store        *prompts.Store
	Client       *llm.Client
}

// Options holds REPL configuration options.
type Options struct {
	Temperature float64
	Models      map[string]config.Model
	AutoCopy    bool
}

// readResult holds a single line read from readline.
type readResult struct {
	line string
	err  error
}

// Run starts the interactive REPL loop.
func Run(state *State, opts Options) error {
	ac := newAutoCompleter(state, opts)
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "> ",
		HistoryFile:     getHistoryFile(),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		AutoComplete:    ac,
		Painter:         newSuggestionPainter(ac),
	})
	if err != nil {
		return fmt.Errorf("failed to create readline: %w", err)
	}
	defer func() {
		_ = rl.Close()
	}()

	// Run readline in a dedicated goroutine so we can:
	// 1. Buffer pasted multi-line text into a single input
	// 2. Receive ^C (ErrInterrupt) during generation to cancel it
	lineCh := make(chan readResult, 512) // large buffer so paste lines queue up
	go func() {
		defer close(lineCh)
		for {
			line, lineErr := rl.Readline()
			lineCh <- readResult{line, lineErr}
			if lineErr != nil && lineErr != readline.ErrInterrupt {
				return // EOF or permanent error — stop reading
			}
		}
	}()

	printWelcome(state)

	// Main REPL loop
	for {
		result, ok := <-lineCh
		if !ok {
			break // readline goroutine exited
		}
		if result.err != nil {
			if result.err == readline.ErrInterrupt {
				// ^C while waiting for input — show new prompt and continue
				drainLines(lineCh)
				_, _ = fmt.Fprint(os.Stdout, "> ")
				continue
			}
			if result.err == io.EOF {
				break
			}
			return result.err
		}

		line := strings.TrimSpace(result.line)
		if line == "" {
			// Empty line — show prompt and continue
			_, _ = fmt.Fprint(os.Stdout, "> ")
			continue
		}

		// Collect remaining pasted lines that arrive rapidly after the first.
		// readline delivers pasted multi-line text one line at a time; we buffer
		// them into a single input to avoid firing N separate API calls.
		fullInput := collectPastedInput(lineCh, line)

		// Handle slash commands (only when input is a single-line slash command)
		if strings.HasPrefix(fullInput, "/") && !strings.Contains(fullInput, "\n") {
			if shouldExit := handleSlashCommand(fullInput, state, opts.Models); shouldExit {
				break
			}
			_, _ = fmt.Fprint(os.Stdout, "> ")
			continue
		}

		// Generate response with cancellation support.
		// We monitor lineCh for ^C (ErrInterrupt) during generation.
		cancelled := runGenerationWithCancel(fullInput, state, lineCh, opts)

		// Drain any buffered lines that queued up during streaming
		// (e.g. remaining paste lines after cancellation).
		drainLines(lineCh)

		if cancelled {
			fmt.Fprintln(os.Stderr, output.Yellow("\nGeneration cancelled"))
		}

		// Print the prompt to signal readiness for next input.
		// Since readline runs in a goroutine and we're reading from a channel,
		// we must explicitly print the prompt to provide visual feedback.
		_, _ = fmt.Fprint(os.Stdout, "> ")
	}

	fmt.Println(output.Bold(output.Green("\nGoodbye!")))
	return nil
}

// runGenerationWithCancel runs generateStreaming in a goroutine while monitoring
// lineCh for ^C interrupts. Returns true if generation was cancelled.
func runGenerationWithCancel(input string, state *State, lineCh <-chan readResult, opts Options) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Run generation in a goroutine so we can monitor for ^C
	errCh := make(chan error, 1)
	go func() {
		errCh <- generateStreaming(ctx, input, state, opts)
	}()

	// Wait for either generation to finish or ^C from readline
	for {
		select {
		case err := <-errCh:
			if err != nil {
				errMsg := err.Error()
				if strings.Contains(errMsg, "context canceled") || strings.Contains(errMsg, "context deadline exceeded") {
					return true
				}
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			return false

		case result, ok := <-lineCh:
			if !ok {
				// Channel closed (readline exited) — cancel generation
				cancel()
				<-errCh
				return true
			}
			if result.err == readline.ErrInterrupt {
				// ^C pressed — cancel the in-flight generation
				cancel()
				<-errCh // wait for generation to finish
				return true
			}
			if result.err == io.EOF {
				cancel()
				<-errCh
				return true
			}
			// Non-error line during generation — ignore it (could be leftover paste)
		}
	}
}

// collectPastedInput buffers additional lines that arrive almost immediately
// after the first line, indicating the user pasted multi-line text.
func collectPastedInput(lineCh <-chan readResult, firstLine string) string {
	const pasteTimeout = 80 * time.Millisecond

	lines := []string{firstLine}
	for {
		select {
		case result, ok := <-lineCh:
			if !ok {
				return strings.Join(lines, "\n")
			}
			if result.err != nil {
				// Interrupt/EOF during paste — return what we have
				return strings.Join(lines, "\n")
			}
			trimmed := strings.TrimSpace(result.line)
			lines = append(lines, trimmed) // keep empty lines for structure
		case <-time.After(pasteTimeout):
			return strings.Join(lines, "\n")
		}
	}
}

// drainLines discards any buffered lines in the channel (non-blocking).
func drainLines(lineCh <-chan readResult) {
	for {
		select {
		case _, ok := <-lineCh:
			if !ok {
				return
			}
		default:
			return
		}
	}
}

func printWelcome(state *State) {
	for _, line := range formatWelcomeLines(state) {
		fmt.Print(line)
	}
}

func formatWelcomeLines(state *State) []string {
	return []string{
		fmt.Sprintf("%s\n", output.Cyan("raypaste interactive mode")),
		fmt.Sprintf(
			"Model: %s | Length: %s | Prompt: %s\n",
			output.Bold(output.Blue(state.Model)),
			output.Bold(output.Yellow(string(state.Length))),
			output.Bold(output.Green(state.PromptName)),
		),
		fmt.Sprintf(
			"Type %s for commands, %s or %s to close raypaste\n",
			output.Bold(output.Green("/help")),
			output.Bold(output.Red("Ctrl+D")),
			output.Bold(output.Red("/quit")),
		),
		"\n",
	}
}

func getHistoryFile() string {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return ""
	}
	return configDir + "/history"
}

func clearScreen() {
	// Use ANSI escape codes to clear the screen and move cursor to top-left
	fmt.Print("\033[2J\033[H")
}
