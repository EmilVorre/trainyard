package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/Emilvorre/trainyard/internal/scaffold"
	"github.com/Emilvorre/trainyard/internal/setup"
	"github.com/Emilvorre/trainyard/internal/validate"
)

var rootCmd = &cobra.Command{
	Use: "yard",
	Short: "Trainyard is a tool for managing ephemeral Kubernetes preview environments",
	Long: `yard is the CLI for Trainyard.

Run on your server to set up a preview environment host, or in a repo to scaffold the GitHub Actions config for PR previews.`,
}

func main() {
	rootCmd.AddCommand(setup.Command())
	rootCmd.AddCommand(scaffold.Command())
	rootCmd.AddCommand(validate.Command())
	// Future Features: rootCmd.AddCommand(status.Command())
	// Future Features: rootCmd.AddCommand(logs.Command())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}