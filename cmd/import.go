package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/indiekitai/mrr-cli/db"
	"github.com/indiekitai/mrr-cli/models"
)

var importCmd = &cobra.Command{
	Use:   "import <file.csv>",
	Short: "Import entries from CSV",
	Long: `Import revenue entries from a CSV file.

CSV format:
  date,amount,source,type,note
  2026-02-01,49.99,stripe,recurring,SaaS subscription
  2026-02-15,19.00,gumroad,one-time,ebook sale

Examples:
  mrr import entries.csv`,
	Args: cobra.ExactArgs(1),
	RunE: runImport,
}

func runImport(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return fmt.Errorf("CSV file is empty or has only headers")
	}

	// Validate header
	header := records[0]
	expectedHeader := []string{"date", "amount", "source", "type", "note"}
	if len(header) < 4 {
		return fmt.Errorf("invalid CSV header, expected: %v", expectedHeader)
	}

	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	imported := 0
	skipped := 0

	for i, record := range records[1:] {
		lineNum := i + 2 // Account for header and 0-indexing

		if len(record) < 4 {
			fmt.Printf("%s Line %d: insufficient fields, skipping\n", yellow("⚠"), lineNum)
			skipped++
			continue
		}

		// Parse date
		dateStr := strings.TrimSpace(record[0])
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			fmt.Printf("%s Line %d: invalid date '%s', skipping\n", red("✗"), lineNum, dateStr)
			skipped++
			continue
		}

		// Parse amount
		amountStr := strings.TrimSpace(record[1])
		amountStr = strings.TrimPrefix(amountStr, "$")
		amountFloat, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			fmt.Printf("%s Line %d: invalid amount '%s', skipping\n", red("✗"), lineNum, record[1])
			skipped++
			continue
		}
		amountCents := int64(amountFloat * 100)

		// Validate source
		source := strings.TrimSpace(strings.ToLower(record[2]))
		if !models.IsValidSource(source) {
			fmt.Printf("%s Line %d: invalid source '%s', skipping\n", red("✗"), lineNum, source)
			skipped++
			continue
		}

		// Validate type
		entryType := strings.TrimSpace(strings.ToLower(record[3]))
		if !models.IsValidType(entryType) {
			fmt.Printf("%s Line %d: invalid type '%s', skipping\n", red("✗"), lineNum, entryType)
			skipped++
			continue
		}

		// Note (optional)
		note := ""
		if len(record) > 4 {
			note = strings.TrimSpace(record[4])
		}

		// Add to database
		_, err = db.AddEntry(amountCents, source, entryType, note, date)
		if err != nil {
			fmt.Printf("%s Line %d: failed to add entry: %v\n", red("✗"), lineNum, err)
			skipped++
			continue
		}

		imported++
	}

	fmt.Printf("\n%s Imported %d entries", green("✓"), imported)
	if skipped > 0 {
		fmt.Printf(", %s %d entries", yellow("skipped"), skipped)
	}
	fmt.Println()

	return nil
}
