/*
Copyright © 2026 Raypaste
*/
package cmd

import (
	"fmt"

	"github.com/raypaste/raypaste-cli/internal/output"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  output.Bold("Print version information") + output.Cyan(" for raypaste-cli"),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s %s\n", output.Bold("raypaste-cli"), output.Cyan(Version))
		fmt.Printf("Git commit: %s\n", GitCommit)
		fmt.Printf("Build date: %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
