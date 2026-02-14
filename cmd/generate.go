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

	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:     "generate [input]",
	Aliases: []string{"gen", "g"},
	Short:   "Generate an optimized prompt from your input",
	Long: `Generate an optimized prompt from your input text.

The generate command takes your input (from arguments or stdin) and generates
an optimized prompt using the specified model and output length.

Examples:
  raypaste generate "help me write a blog post" --length short --copy
  raypaste gen "analyze CSV data" -l long
  echo "my goal" | raypaste gen
  raypaste g "create a REST API" -m cerebras-llama-8b`,
	RunE: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Get input from args or stdin
	input, err := getInput(args)
	if err != nil {
		return fmt.Errorf("failed to get input: %w", err)
	}

	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("no input provided")
	}

	// Validate and get output length
	length, err := config.ValidateOutputLength(lengthFlag)
	if err != nil {
		return err
	}

	// Get model (use flag if set, otherwise config default)
	model := modelFlag
	if model == "" {
		model = cfg.GetDefaultModel()
	}

	store, err := prompts.NewStore()
	if err != nil {
		return fmt.Errorf("failed to load prompts: %w", err)
	}

	systemPrompt, err := store.Render(promptFlag, length)
	if err != nil {
		return fmt.Errorf("failed to render prompt: %w", err)
	}

	req, err := llm.BuildRequest(
		model,
		systemPrompt,
		input,
		length,
		cfg.Temperature,
		false, // no streaming for generate
		cfg.Models,
	)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	client := llm.NewClient(cfg.GetAPIKey())

	// Show progress indicator
	fmt.Fprintln(os.Stderr, "Generating...")
	fmt.Fprintln(os.Stderr, "")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := client.Complete(ctx, req)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	// Print result to stdout
	fmt.Println(result)

	// Copy to clipboard if requested
	shouldCopy := copyFlag || cfg.AutoCopy
	if shouldCopy {
		if warning := clipboard.CopyWithWarning(result); warning != "" {
			fmt.Fprintln(os.Stderr, warning)
		} else {
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "✓ Copied to clipboard")
		}
	}

	return nil
}

// getInput gets input from args or stdin
func getInput(args []string) (string, error) {
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}

	// Check if stdin has data
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	// If stdin is a pipe or file, read from it
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read stdin: %w", err)
		}
		return string(data), nil
	}

	return "", nil
}
