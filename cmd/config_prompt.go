/*
Copyright © 2026 Raypaste
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/raypaste/raypaste-cli/internal/output"
	"github.com/raypaste/raypaste-cli/internal/prompts"
	"github.com/raypaste/raypaste-cli/pkg/types"
	"github.com/spf13/cobra"
)

// configPromptCmd represents the config prompt command
var configPromptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Manage custom prompt templates",
	Long: output.Bold("Manage custom prompt templates") + output.Cyan(" for raypaste.") + `

This command allows you to create, view, list, and remove custom prompt templates.
Custom prompts are stored in ~/.raypaste/prompts/.

` + output.Bold("Available subcommands:") + `
  ` + output.Green("add") + `    - Add a new custom prompt interactively or via flags
  ` + output.Green("list") + `   - List all available prompts (built-in and custom)
  ` + output.Green("show") + `   - Show details of a specific prompt
  ` + output.Green("remove") + ` - Remove a custom prompt

` + output.Bold("Examples:") + `
  raypaste config prompt add code-review
  raypaste config prompt list
  raypaste config prompt show metaprompt
  raypaste config prompt remove my-custom-prompt`,
}

// configPromptAddCmd represents the config prompt add command
var configPromptAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new custom prompt",
	Long:  `Add a new custom prompt template interactively or via command-line flags.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Validate prompt name (no spaces, special chars)
		if !isValidPromptName(name) {
			return fmt.Errorf("invalid prompt name: %s (must contain only letters, numbers, hyphens, and underscores)", name)
		}

		// Check if prompt already exists
		store, err := prompts.NewStore()
		if err != nil {
			return fmt.Errorf("failed to load prompts: %w", err)
		}

		existing, _ := store.Get(name)
		if existing != nil {
			return fmt.Errorf("prompt '%s' already exists. Use 'remove' first to overwrite", name)
		}

		// Get flags
		description, _ := cmd.Flags().GetString("description")
		shortDirective, _ := cmd.Flags().GetString("short")
		mediumDirective, _ := cmd.Flags().GetString("medium")
		longDirective, _ := cmd.Flags().GetString("long")
		system, _ := cmd.Flags().GetString("system")
		fromFile, _ := cmd.Flags().GetString("from-file")

		// If from-file is specified, read system prompt from file
		if fromFile != "" {
			data, err := os.ReadFile(fromFile)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", fromFile, err)
			}
			system = string(data)
		}

		// If not all required flags provided, enter interactive mode
		if description == "" || system == "" {
			fmt.Fprintln(os.Stderr, output.Blue("Creating custom prompt:")+" "+output.Bold(name))
			fmt.Fprintln(os.Stderr, output.Red("Press Ctrl+C to cancel at any time"))
			fmt.Fprintln(os.Stderr, "")
		}

		reader := bufio.NewReader(os.Stdin)

		// Get description
		if description == "" {
			fmt.Fprint(os.Stderr, output.Blue("Description")+output.White(" (what this prompt does - NOT included in the system prompt to LLM): "))
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			description = strings.TrimSpace(input)
			if description == "" {
				return fmt.Errorf("description is required")
			}
		}

		// Get system prompt
		if system == "" {
			fmt.Fprintln(os.Stderr, output.Blue("System Prompt")+output.White(" (the main prompt template):"))
			fmt.Fprintln(os.Stderr, output.Blue("Hint:")+output.White(" Use {{.LengthDirective}} for length guidance and {{.Context}} for project context"))
			fmt.Fprintln(os.Stderr, output.Blue("Input:")+output.White(" (Press Ctrl+D or type 'EOF' on a new line to finish):"))

			var lines []string
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					// Ctrl+D pressed
					break
				}
				line = strings.TrimRight(line, "\n")
				if line == "EOF" {
					break
				}
				lines = append(lines, line)
			}
			system = strings.Join(lines, "\n")
			if strings.TrimSpace(system) == "" {
				return fmt.Errorf("system prompt is required")
			}
		}

		// Build length directives
		lengthDirectives := make(map[string]string)

		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, output.Bold("Length Directives"))
		fmt.Fprintln(os.Stderr, output.White("Controls how much output the LLM generates for each length mode. Two options:"))
		fmt.Fprintln(os.Stderr, "  "+output.Cyan("Token count")+" — a plain integer sets the "+output.Bold("max_tokens")+" API parameter")
		fmt.Fprintln(os.Stderr, "    "+output.Yellow("e.g. 300, 800, 1500"))
		fmt.Fprintln(os.Stderr, "  "+output.Cyan("Text directive")+" — a string is injected into "+output.Bold("{{.LengthDirective}}")+" in your system prompt")
		fmt.Fprintln(os.Stderr, "    "+output.Yellow(`e.g. "Be concise, 2-3 sentences max"`))
		fmt.Fprintln(os.Stderr, output.White("Press Enter to skip a length (it will not be supported by this prompt)."))
		fmt.Fprintln(os.Stderr, "")

		// Get short directive
		if shortDirective == "" {
			fmt.Fprint(os.Stderr, output.Blue("  Short")+" (optional, press Enter to skip): ")
			input, _ := reader.ReadString('\n')
			shortDirective = strings.TrimSpace(input)
		}
		if shortDirective != "" {
			lengthDirectives["short"] = shortDirective
		}

		// Get medium directive
		if mediumDirective == "" {
			fmt.Fprint(os.Stderr, output.Blue("  Medium")+" (optional, press Enter to skip): ")
			input, _ := reader.ReadString('\n')
			mediumDirective = strings.TrimSpace(input)
		}
		if mediumDirective != "" {
			lengthDirectives["medium"] = mediumDirective
		}

		// Get long directive
		if longDirective == "" {
			fmt.Fprint(os.Stderr, output.Blue("  Long")+" (optional, press Enter to skip): ")
			input, _ := reader.ReadString('\n')
			longDirective = strings.TrimSpace(input)
		}
		if longDirective != "" {
			lengthDirectives["long"] = longDirective
		}

		// If no directives provided, default to supporting all lengths with empty strings
		// (this will use the default directives from llm.LengthParams)
		if len(lengthDirectives) == 0 {
			lengthDirectives["short"] = ""
			lengthDirectives["medium"] = ""
			lengthDirectives["long"] = ""
		}

		// Create the prompt
		prompt := &prompts.Prompt{
			Name:             name,
			Description:      description,
			System:           system,
			LengthDirectives: lengthDirectives,
		}

		// Save the prompt
		if err := store.SavePrompt(prompt); err != nil {
			return fmt.Errorf("failed to save prompt: %w", err)
		}

		fmt.Fprintf(os.Stderr, "%s %s %s %s\n", output.Green("✓"), output.Green("Custom prompt"), output.BoldHiGreen(name), output.Green("added successfully"))
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintf(os.Stderr, "%s %s %s\n", output.Blue("→"), output.White("Use it with:"), output.BoldWhite(fmt.Sprintf("raypaste '...' -p %s", name)))
		fmt.Fprintln(os.Stderr, output.BoldYellow("Important for instant completion mode:")+output.Yellow(" use single quotes and not double quotes to protect table/column names wrapped in backticks `"))
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintf(os.Stderr, "%s %s %s\n", output.Blue("→"), output.White("Use it in interactive mode by typing:"), output.BoldWhite("/prompt <name>"))
		fmt.Fprintln(os.Stderr, "")

		return nil
	},
}

// configPromptListCmd represents the config prompt list command
var configPromptListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available prompts",
	Long:  `List all available prompt templates, including built-in and custom prompts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := prompts.NewStore()
		if err != nil {
			return fmt.Errorf("failed to load prompts: %w", err)
		}

		allPrompts := store.List()

		fmt.Fprintln(os.Stderr, output.Bold("Available Prompts"))
		fmt.Fprintln(os.Stderr, strings.Repeat("=", 50))

		for _, name := range allPrompts {
			prompt, err := store.Get(name)
			if err != nil {
				continue
			}

			isBuiltIn := store.IsBuiltIn(name)
			status := output.Green("[built-in]")
			if !isBuiltIn {
				status = output.Cyan("[custom]")
			}

			// Show supported lengths
			var lengths []string
			for _, length := range []types.OutputLength{types.OutputLengthShort, types.OutputLengthMedium, types.OutputLengthLong} {
				if _, ok := prompt.LengthDirectives[string(length)]; ok {
					lengths = append(lengths, string(length))
				}
			}

			fmt.Fprintf(os.Stderr, "%s %s %s\n", output.Bold(name), status, output.Yellow("["+strings.Join(lengths, ", ")+"]"))
			fmt.Fprintf(os.Stderr, "  %s\n", prompt.Description)
		}

		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, output.Cyan("Tip:")+" Use 'raypaste config prompt show <name>' to view prompt details")

		return nil
	},
}

// configPromptShowCmd represents the config prompt show command
var configPromptShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show prompt details",
	Long:  `Display the details of a specific prompt template.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		store, err := prompts.NewStore()
		if err != nil {
			return fmt.Errorf("failed to load prompts: %w", err)
		}

		prompt, err := store.Get(name)
		if err != nil {
			return err
		}

		isBuiltIn := store.IsBuiltIn(name)
		status := output.Green("built-in")
		if !isBuiltIn {
			status = output.Cyan("custom")
		}

		fmt.Fprintf(os.Stderr, "%s: %s (%s)\n", output.Bold("Name"), output.Cyan(prompt.Name), status)
		fmt.Fprintf(os.Stderr, "%s: %s\n", output.Bold("Description"), prompt.Description)

		// Show supported lengths
		fmt.Fprintf(os.Stderr, "%s: ", output.Bold("Supported Lengths"))
		var lengths []string
		for _, length := range []types.OutputLength{types.OutputLengthShort, types.OutputLengthMedium, types.OutputLengthLong} {
			if directive, ok := prompt.LengthDirectives[string(length)]; ok {
				if directive != "" {
					lengths = append(lengths, string(length)+" (custom)")
				} else {
					lengths = append(lengths, string(length)+" (default)")
				}
			}
		}
		fmt.Fprintln(os.Stderr, strings.Join(lengths, ", "))

		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, output.Bold("System Prompt:"))
		fmt.Fprintln(os.Stderr, strings.Repeat("-", 50))
		fmt.Println(prompt.System)
		fmt.Fprintln(os.Stderr, strings.Repeat("-", 50))

		return nil
	},
}

// configPromptRemoveCmd represents the config prompt remove command
var configPromptRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a custom prompt",
	Long:  `Remove a custom prompt template. Built-in prompts cannot be removed.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		store, err := prompts.NewStore()
		if err != nil {
			return fmt.Errorf("failed to load prompts: %w", err)
		}

		// Check if it's a built-in prompt
		if store.IsBuiltIn(name) {
			return fmt.Errorf("cannot remove built-in prompt: %s", name)
		}

		// Confirm deletion unless --force is used
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Fprintf(os.Stderr, "%s %s %s %s [y/N]: ", output.Yellow("?"), output.Yellow("Are you sure you want to remove prompt (this action cannot be undone)"), output.BoldYellow(name), output.Yellow("?"))
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Fprintln(os.Stderr, output.Red("Cancelled - prompt not removed"))
				return nil
			}
		}

		if err := store.DeletePrompt(name); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "%s %s %s %s\n", output.Green("✓"), output.Green("Prompt"), output.BoldBlue(name), output.Green("removed successfully"))

		return nil
	},
}

// isValidPromptName checks if a prompt name contains only valid characters
func isValidPromptName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if !isValidPromptChar(r) {
			return false
		}
	}
	return true
}

// isValidPromptChar checks if a rune is a valid character for prompt names
func isValidPromptChar(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z':
		return true
	case r >= 'A' && r <= 'Z':
		return true
	case r >= '0' && r <= '9':
		return true
	case r == '-', r == '_':
		return true
	default:
		return false
	}
}

func init() {
	configCmd.AddCommand(configPromptCmd)

	// Add subcommands
	configPromptCmd.AddCommand(configPromptAddCmd)
	configPromptCmd.AddCommand(configPromptListCmd)
	configPromptCmd.AddCommand(configPromptShowCmd)
	configPromptCmd.AddCommand(configPromptRemoveCmd)

	// Flags for add command
	configPromptAddCmd.Flags().StringP("description", "d", "", "Description of the prompt")
	configPromptAddCmd.Flags().String("short", "", "Directive for short output length")
	configPromptAddCmd.Flags().String("medium", "", "Directive for medium output length")
	configPromptAddCmd.Flags().String("long", "", "Directive for long output length")
	configPromptAddCmd.Flags().StringP("system", "s", "", "System prompt template (can also use --from-file)")
	configPromptAddCmd.Flags().StringP("from-file", "f", "", "Read system prompt from file")

	// Flags for remove command
	configPromptRemoveCmd.Flags().BoolP("force", "f", false, "Force removal without confirmation")
}
