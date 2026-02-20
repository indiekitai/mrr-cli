package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/indiekitai/mrr-cli/db"
	"github.com/indiekitai/mrr-cli/models"
)

var (
	exportMonth  string
	exportOutput string
	exportJSON   bool
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export entries to CSV",
	Long: `Export revenue entries to a CSV file.

Examples:
  mrr export                          # Export all entries to stdout
  mrr export --month 2026-02          # Export specific month
  mrr export --output entries.csv     # Export to file
  mrr export --json                   # Export as JSON`,
	RunE: runExport,
}

func init() {
	exportCmd.Flags().StringVarP(&exportMonth, "month", "m", "", "Month to export (YYYY-MM)")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file path")
	exportCmd.Flags().BoolVarP(&exportJSON, "json", "j", false, "Output as JSON")
}

type exportEntry struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
	Source string  `json:"source"`
	Type   string  `json:"type"`
	Note   string  `json:"note"`
}

func runExport(cmd *cobra.Command, args []string) error {
	entries, err := db.ListEntries(exportMonth, "", "")
	if err != nil {
		return err
	}

	if exportJSON {
		return exportAsJSON(entries)
	}

	return exportAsCSV(entries)
}

func exportAsJSON(entries []models.Entry) error {
	var exportEntries []exportEntry
	for _, e := range entries {
		exportEntries = append(exportEntries, exportEntry{
			Date:   e.Date.Format("2006-01-02"),
			Amount: float64(e.Amount) / 100.0,
			Source: e.Source,
			Type:   e.Type,
			Note:   e.Note,
		})
	}

	var output *os.File
	var err error
	if exportOutput != "" {
		output, err = os.Create(exportOutput)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer output.Close()
	} else {
		output = os.Stdout
	}

	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(exportEntries)
}

func exportAsCSV(entries []models.Entry) error {
	var output *os.File
	var err error
	if exportOutput != "" {
		output, err = os.Create(exportOutput)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer output.Close()
	} else {
		output = os.Stdout
	}

	writer := csv.NewWriter(output)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"date", "amount", "source", "type", "note"}); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write entries
	for _, e := range entries {
		record := []string{
			e.Date.Format("2006-01-02"),
			fmt.Sprintf("%.2f", float64(e.Amount)/100.0),
			e.Source,
			e.Type,
			e.Note,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}
