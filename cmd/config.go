/*
Copyright © 2026 Raypaste
*/
package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/raypaste/raypaste-cli/internal/config"
	"github.com/raypaste/raypaste-cli/internal/output"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage raypaste configuration",
	Long: output.Bold("Manage raypaste configuration") + output.Cyan(" settings.") + `

This command allows you to view and modify configuration values.
Configuration is stored in ~/.raypaste/config.yaml.

` + output.Bold("Available config keys:") + `
  ` + output.Green("api-key") + `        - OpenRouter API key
  ` + output.Green("default-model") + `  - Default model alias or OpenRouter ID
  ` + output.Green("default-length") + ` - Default output length (short|medium|long)
  ` + output.Green("disable-copy") + `   - Disable auto-copy to clipboard (true|false)
  ` + output.Green("temperature") + `    - Sampling temperature (0.0-2.0)

` + output.Bold("Examples:") + `
  raypaste config set api-key sk-or-v1-...
  raypaste config set default-model cerebras-llama-8b
  raypaste config set default-length short
  raypaste config set disable-copy true`,
}

// configSetCmd represents the config set command
var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long:  `Set a configuration value and save it to the config file.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.ToLower(args[0])
		value := args[1]

		// Reload config to ensure we're working with current values
		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		switch key {
		case "api-key":
			cfg.APIKey = value
			if err := cfg.SaveTo(cfgFile); err != nil {
				return err
			}
			fmt.Fprintln(os.Stderr, output.Green("✓")+" API key saved to config")

		case "default-model", "model":
			cfg.DefaultModel = value
			if err := cfg.SaveTo(cfgFile); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "%s Default model set to %s\n", output.Green("✓"), output.Cyan(value))

		case "default-length", "length":
			length, err := config.ValidateOutputLength(value)
			if err != nil {
				return err
			}
			cfg.DefaultLength = length
			if err := cfg.SaveTo(cfgFile); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "%s Default length set to %s\n", output.Green("✓"), output.Cyan(value))

		case "disable-copy":
			disable, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid value for disable-copy: %s (must be true or false)", value)
			}
			cfg.DisableCopy = disable
			if err := cfg.SaveTo(cfgFile); err != nil {
				return err
			}
			status := "enabled"
			if !disable {
				status = "disabled"
			}
			fmt.Fprintf(os.Stderr, "%s Auto-copy %s\n", output.Green("✓"), output.Cyan(status))

		case "temperature":
			temp, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("invalid value for temperature: %s (must be a number between 0.0 and 2.0)", value)
			}
			if temp < 0 || temp > 2.0 {
				return fmt.Errorf("temperature must be between 0.0 and 2.0")
			}
			cfg.Temperature = temp
			if err := cfg.SaveTo(cfgFile); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "%s Temperature set to %s\n", output.Green("✓"), output.Cyan(value))

		default:
			return fmt.Errorf("unknown config key: %s\nAvailable keys: api-key, default-model, default-length, disable-copy, temperature", key)
		}

		return nil
	},
}

// configGetCmd represents the config get command
var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Long:  `Get the current value of a configuration setting.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.ToLower(args[0])

		// Reload config to ensure we're working with current values
		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		switch key {
		case "api-key":
			if cfg.APIKey != "" {
				fmt.Println(cfg.APIKey)
			} else if envKey := os.Getenv("RAYPASTE_API_KEY"); envKey != "" {
				fmt.Println(envKey + " (from environment)")
			} else {
				fmt.Println("not set")
			}

		case "default-model", "model":
			fmt.Println(cfg.GetDefaultModel())

		case "default-length", "length":
			fmt.Println(cfg.GetDefaultLength())

		case "disable-copy":
			fmt.Println(cfg.DisableCopy)

		case "temperature":
			fmt.Println(cfg.Temperature)

		default:
			return fmt.Errorf("unknown config key: %s\nAvailable keys: api-key, default-model, default-length, disable-copy, temperature", key)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
}
