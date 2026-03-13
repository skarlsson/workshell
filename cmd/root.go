package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ws",
	Short: "Workspace manager for AI agent workflows",
	Long:  "ws-manager orchestrates kitty sessions, zellij layouts, and git branches to manage multiple AI-assisted development workspaces.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
