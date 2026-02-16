/*
Copyright © 2026 Raypaste
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/raypaste/raypaste-cli/internal/config"
	"github.com/raypaste/raypaste-cli/internal/interactive"
	"github.com/raypaste/raypaste-cli/internal/llm"
	"github.com/raypaste/raypaste-cli/internal/output"
	"github.com/raypaste/raypaste-cli/internal/projectcontext"
	"github.com/raypaste/raypaste-cli/internal/prompts"

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
  ` + output.Green("/clear") + `                        - Clear the screen
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

func runInteractive(cmd *cobra.Command, args []string) error {
	state := &interactive.State{
		Model:      modelFlag,
		PromptName: promptFlag,
	}

	if state.Model == "" {
		state.Model = cfg.GetDefaultModel()
	}

	length, err := config.ValidateOutputLength(lengthFlag)
	if err != nil {
		return err
	}
	state.Length = length

	state.Store, err = prompts.NewStore()
	if err != nil {
		return fmt.Errorf("failed to load prompts: %w", err)
	}

	workingDir, _ := os.Getwd()
	state.ProjCtx = projectcontext.Load(workingDir)
	state.Client = llm.NewClient(cfg.GetAPIKey())

	return interactive.Run(state, interactive.Options{
		Temperature: cfg.Temperature,
		Models:      cfg.Models,
		AutoCopy:    !noCopyFlag && !cfg.DisableCopy,
	})
}
