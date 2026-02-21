/*
Copyright © 2026 Raypaste
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/raypaste/raypaste-cli/internal/config"
	"github.com/raypaste/raypaste-cli/internal/output"

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
