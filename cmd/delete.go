package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/indiekitai/mrr-cli/db"
	"github.com/indiekitai/mrr-cli/models"
)

var forceDelete bool

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an entry",
	Long: `Delete a revenue entry by ID.

Examples:
  mrr delete 1
  mrr delete 1 -f  # Skip confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

func init() {
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Skip confirmation")
}

func runDelete(cmd *cobra.Command, args []string) error {
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid ID: %s", args[0])
	}

	// Get entry first to show what will be deleted
	entry, err := db.GetEntry(id)
	if err != nil {
		return err
	}

	if !forceDelete {
		yellow := color.New(color.FgYellow).SprintFunc()
		fmt.Printf("%s Delete entry #%d: %s from %s on %s? [y/N] ",
			yellow("⚠"),
			entry.ID,
			models.FormatAmount(entry.Amount, "USD"),
			entry.Source,
			entry.Date.Format("2006-01-02"),
		)

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	if err := db.DeleteEntry(id); err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Deleted entry #%d\n", green("✓"), id)

	return nil
}
