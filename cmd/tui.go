package cmd

import (
	"github.com/spf13/cobra"

	"github.com/indiekitai/mrr-cli/ui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Interactive terminal UI",
	Long: `Launch an interactive terminal UI with vim-style keybindings.

Keybindings:
  j/k     - Navigate up/down
  g/G     - Go to first/last entry
  a       - Add new entry
  e       - Edit selected entry
  d       - Delete selected entry
  r       - Refresh
  q/Esc   - Quit`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ui.Run()
	},
}
