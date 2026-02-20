package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/indiekitai/mrr-cli/db"
	"github.com/indiekitai/mrr-cli/models"
)

var (
	editAmount string
	editSource string
	editNote   string
)

var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit an existing entry",
	Long: `Edit an existing revenue entry by ID.

Examples:
  mrr edit 1 --amount 49.99
  mrr edit 1 --source stripe
  mrr edit 1 --note "Updated note"
  mrr edit 1 --amount 99 --source gumroad`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

func init() {
	editCmd.Flags().StringVarP(&editAmount, "amount", "a", "", "New amount")
	editCmd.Flags().StringVarP(&editSource, "source", "s", "", "New source")
	editCmd.Flags().StringVarP(&editNote, "note", "n", "", "New note")
}

func runEdit(cmd *cobra.Command, args []string) error {
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid ID: %s", args[0])
	}

	var amount *int64
	var source, note *string

	if editAmount != "" {
		amountStr := strings.TrimPrefix(editAmount, "$")
		amountFloat, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return fmt.Errorf("invalid amount: %s", editAmount)
		}
		amountCents := int64(amountFloat * 100)
		amount = &amountCents
	}

	if editSource != "" {
		if !models.IsValidSource(editSource) {
			return fmt.Errorf("invalid source: %s (valid: %v)", editSource, models.ValidSources)
		}
		source = &editSource
	}

	if cmd.Flags().Changed("note") {
		note = &editNote
	}

	if amount == nil && source == nil && note == nil {
		return fmt.Errorf("no fields to update (use --amount, --source, or --note)")
	}

	if err := db.UpdateEntry(id, amount, source, note); err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	fmt.Printf("%s Updated entry #%s\n", green("✓"), cyan(fmt.Sprintf("%d", id)))

	return nil
}
