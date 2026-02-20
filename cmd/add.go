package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/indiekitai/mrr-cli/db"
	"github.com/indiekitai/mrr-cli/models"
)

var (
	addSource string
	addType   string
	addNote   string
	addDate   string
)

var addCmd = &cobra.Command{
	Use:   "add <amount>",
	Short: "Add a new revenue entry",
	Long: `Add a new revenue entry. Amount should be in dollars (e.g., 29.99).

Examples:
  mrr add 29.99
  mrr add 99.00 --source stripe --type recurring
  mrr add 49.99 --source gumroad --note "Lifetime license"
  mrr add 100 --date 2024-01-15`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

func init() {
	addCmd.Flags().StringVarP(&addSource, "source", "s", "manual", "Revenue source (stripe, gumroad, paddle, manual)")
	addCmd.Flags().StringVarP(&addType, "type", "t", "recurring", "Revenue type (recurring, one-time)")
	addCmd.Flags().StringVarP(&addNote, "note", "n", "", "Note for this entry")
	addCmd.Flags().StringVarP(&addDate, "date", "d", "", "Date (YYYY-MM-DD, defaults to today)")
}

func runAdd(cmd *cobra.Command, args []string) error {
	// Parse amount
	amountStr := strings.TrimPrefix(args[0], "$")
	amountFloat, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %s", args[0])
	}
	amountCents := int64(amountFloat * 100)

	// Validate source
	if !models.IsValidSource(addSource) {
		return fmt.Errorf("invalid source: %s (valid: %v)", addSource, models.ValidSources)
	}

	// Validate type
	if !models.IsValidType(addType) {
		return fmt.Errorf("invalid type: %s (valid: %v)", addType, models.ValidTypes)
	}

	// Parse date
	var date time.Time
	if addDate != "" {
		date, err = time.Parse("2006-01-02", addDate)
		if err != nil {
			return fmt.Errorf("invalid date format: %s (use YYYY-MM-DD)", addDate)
		}
	} else {
		date = time.Now()
	}

	// Add to database
	id, err := db.AddEntry(amountCents, addSource, addType, addNote, date)
	if err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	fmt.Printf("%s Added entry #%s: %s from %s (%s)\n",
		green("✓"),
		cyan(fmt.Sprintf("%d", id)),
		models.FormatAmount(amountCents, "USD"),
		addSource,
		addType,
	)

	return nil
}
