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
	"github.com/raypaste/raypaste-cli/internal/projectcontext"
	"github.com/raypaste/raypaste-cli/internal/prompts"

	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	modelFlag  string
	lengthFlag string
	promptFlag string
	noCopyFlag bool
	cfg        *config.Config
)

// Version information (set via -ldflags during build)
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "raypaste [input]",
	Short: output.Cyan("Generate ") + output.Bold("AI-optimized prompts") + output.Cyan(" from your input"),
	Long: output.Bold("raypaste-cli") + output.Cyan(" - Ultra-fast AI revised meta prompts from your input text.") + `

A Cobra-based CLI that generates meta-prompts and general AI completions via OpenRouter,
with configurable output lengths and fast/small model routing.

` + output.Bold("Setup:") + `
  raypaste config set api-key ` + output.Green("sk-or-v1-...") + `        ` + output.Cyan("# Set your OpenRouter API key") + `
  raypaste config set default-model ` + output.Green("cerebras-llama-8b") + `  ` + output.Cyan("# Set default model") + `
  raypaste config ` + output.Green("get default-model") + `                     ` + output.Cyan("# View current settings") + `
  raypaste config prompt ` + output.Green("add my-prompt") + `                 ` + output.Cyan("# Add a custom prompt") + `

` + output.Bold("Examples:") + `
  raypaste "help me write a blog post" ` + output.Green("--length short") + `
  raypaste "analyze CSV data" ` + output.Green("-l long") + `
  echo "my goal" | raypaste
  raypaste interactive`,
	Args: func(cmd *cobra.Command, args []string) error {
		versionFlag, _ := cmd.Flags().GetBool("version")
		if versionFlag {
			return nil
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		versionFlag, _ := cmd.Flags().GetBool("version")
		if versionFlag {
			fmt.Printf("%s %s\n", output.Bold("raypaste-cli"), output.Cyan(Version))
			fmt.Printf("Git commit: %s\n", GitCommit)
			fmt.Printf("Build date: %s\n", BuildDate)
			return nil
		}
		return runGenerate(cmd, args)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Version flag
	rootCmd.Flags().BoolP("version", "v", false, "Print version information")

	// Persistent flags (available to all subcommands)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.raypaste/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&modelFlag, "model", "m", "", "Model alias or OpenRouter ID")
	rootCmd.PersistentFlags().StringVarP(&lengthFlag, "length", "l", "medium", "Output length: short|medium|long")
	rootCmd.PersistentFlags().StringVarP(&promptFlag, "prompt", "p", "metaprompt", "Prompt template name")
	rootCmd.PersistentFlags().BoolVar(&noCopyFlag, "no-copy", false, "Disable auto-copy to clipboard")
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	// Skip API key validation for config commands (they can set the key)
	if len(os.Args) > 1 && os.Args[1] == "config" {
		var err error
		cfg, err = config.LoadConfig(cfgFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		return
	}

	var err error
	cfg, err = config.LoadConfig(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Validate API key
	if cfg.GetAPIKey() == "" {
		fmt.Fprintln(os.Stderr, "Error: API key not found. Set RAYPASTE_API_KEY environment variable or add to config.yaml")
		os.Exit(1)
	}
}

// runGenerate handles the generation logic for raypaste "text"
func runGenerate(_ *cobra.Command, args []string) error {
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

	workingDir, _ := os.Getwd()
	projCtx := projectcontext.Load(workingDir)

	systemPrompt, err := store.Render(promptFlag, length, projCtx.Content)
	if err != nil {
		return fmt.Errorf("failed to render prompt: %w", err)
	}

	maxTokensOverride := store.GetMaxTokensOverride(promptFlag, length)

	req, err := llm.BuildRequest(
		model,
		systemPrompt,
		input,
		length,
		cfg.Temperature,
		false, // no streaming for generate
		cfg.Models,
		maxTokensOverride,
	)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	client := llm.NewClient(cfg.GetAPIKey())

	// Show progress indicator
	fmt.Fprintln(os.Stderr, output.GeneratingMessage(model, string(length), projCtx.Filename))
	fmt.Fprintln(os.Stderr, "")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	startTime := time.Now()
	result, usage, err := client.Complete(ctx, req)
	durationMs := time.Since(startTime).Milliseconds()
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	// Print result to stdout (colorize if markdown)
	fmt.Println(output.ColorizeMarkdown(result))

	// Copy to clipboard by default unless disabled
	shouldCopy := !noCopyFlag && !cfg.DisableCopy
	if shouldCopy {
		if warning := clipboard.CopyWithWarning(result); warning != "" {
			fmt.Fprintln(os.Stderr, warning)
		} else {
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, output.CopiedMessage())
		}
	}

	// Display token usage and completion time
	fmt.Fprintln(os.Stderr, output.TokenUsageMessage(usage.PromptTokens, usage.CompletionTokens, durationMs))

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
