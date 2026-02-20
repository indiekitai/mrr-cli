package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/indiekitai/mrr-cli/db"
	"github.com/indiekitai/mrr-cli/models"
)

var (
	listMonth  string
	listSource string
	listType   string
	listJSON   bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List revenue entries",
	Long: `List all revenue entries with optional filters.

Examples:
  mrr list
  mrr list --month 2024-01
  mrr list --source stripe
  mrr list --type recurring
  mrr list --json`,
	RunE: runList,
}

func init() {
	listCmd.Flags().StringVarP(&listMonth, "month", "m", "", "Filter by month (YYYY-MM)")
	listCmd.Flags().StringVarP(&listSource, "source", "s", "", "Filter by source")
	listCmd.Flags().StringVarP(&listType, "type", "t", "", "Filter by type")
	listCmd.Flags().BoolVarP(&listJSON, "json", "j", false, "Output as JSON")
}

type listEntry struct {
	ID        int64   `json:"id"`
	Date      string  `json:"date"`
	Amount    float64 `json:"amount"`
	Source    string  `json:"source"`
	Type      string  `json:"type"`
	Note      string  `json:"note,omitempty"`
	CreatedAt string  `json:"created_at"`
}

type listOutput struct {
	Entries []listEntry `json:"entries"`
	Total   float64     `json:"total"`
	Count   int         `json:"count"`
}

func runList(cmd *cobra.Command, args []string) error {
	entries, err := db.ListEntries(listMonth, listSource, listType)
	if err != nil {
		return err
	}

	if listJSON {
		return listAsJSON(entries)
	}

	if len(entries) == 0 {
		yellow := color.New(color.FgYellow).SprintFunc()
		fmt.Printf("%s No entries found.\n", yellow("⚠"))
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Date", "Amount", "Source", "Type", "Note"})
	table.SetBorder(false)
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
	)

	var total int64
	for _, e := range entries {
		total += e.Amount
		
		note := e.Note
		if len(note) > 30 {
			note = note[:27] + "..."
		}

		typeColor := tablewriter.FgGreenColor
		if e.Type == "one-time" {
			typeColor = tablewriter.FgYellowColor
		}

		table.Rich([]string{
			fmt.Sprintf("%d", e.ID),
			e.Date.Format("2006-01-02"),
			models.FormatAmount(e.Amount, "USD"),
			e.Source,
			e.Type,
			note,
		}, []tablewriter.Colors{
			{},
			{},
			{tablewriter.FgGreenColor},
			{tablewriter.FgMagentaColor},
			{typeColor},
			{},
		})
	}

	table.Render()

	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("\n%s entries, total: %s\n", cyan(fmt.Sprintf("%d", len(entries))), green(models.FormatAmount(total, "USD")))

	return nil
}

func listAsJSON(entries []models.Entry) error {
	var output listOutput
	var total int64

	for _, e := range entries {
		total += e.Amount
		output.Entries = append(output.Entries, listEntry{
			ID:        e.ID,
			Date:      e.Date.Format("2006-01-02"),
			Amount:    float64(e.Amount) / 100.0,
			Source:    e.Source,
			Type:      e.Type,
			Note:      e.Note,
			CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	output.Total = float64(total) / 100.0
	output.Count = len(entries)

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
