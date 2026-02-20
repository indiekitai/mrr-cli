package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/indiekitai/mrr-cli/db"
)

var rootCmd = &cobra.Command{
	Use:   "mrr",
	Short: "MRR Tracker - Track your Monthly Recurring Revenue",
	Long: `MRR Tracker is a terminal-based tool for indie hackers to track
their Monthly Recurring Revenue (MRR) from various sources.

Store data locally in SQLite, view pretty reports, and use a
VisiData-style TUI for interactive management.

Examples:
  mrr add 29.99 --source stripe
  mrr list --month 2024-01
  mrr report
  mrr forecast
  mrr export --json
  mrr tui`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return db.Init()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		db.Close()
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(forecastCmd)
	rootCmd.AddCommand(badgeCmd)
	rootCmd.AddCommand(goalCmd)
	rootCmd.AddCommand(serveCmd)
}
