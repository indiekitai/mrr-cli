package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/indiekitai/mrr-cli/db"
	"github.com/indiekitai/mrr-cli/models"
)

var reportMonth string

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate monthly report",
	Long: `Generate a monthly revenue report with MRR, growth rate, and breakdown.

Examples:
  mrr report
  mrr report --month 2024-01`,
	RunE: runReport,
}

func init() {
	reportCmd.Flags().StringVarP(&reportMonth, "month", "m", "", "Month to report (YYYY-MM, defaults to current)")
}

func runReport(cmd *cobra.Command, args []string) error {
	month := reportMonth
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	report, err := db.GetMonthlyReport(month)
	if err != nil {
		return err
	}

	// Colors
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	// Header
	fmt.Println()
	fmt.Printf("  %s\n", cyan(fmt.Sprintf("📊 Revenue Report: %s", month)))
	fmt.Println("  " + strings.Repeat("─", 40))
	fmt.Println()

	if report.EntryCount == 0 {
		fmt.Printf("  %s No entries for this month.\n\n", yellow("⚠"))
		return nil
	}

	// Main metrics
	fmt.Printf("  %s  %s\n", bold("MRR (Recurring):"), green(models.FormatAmount(report.RecurringRevenue, "USD")))
	fmt.Printf("  %s %s\n", bold("One-time Revenue:"), yellow(models.FormatAmount(report.OneTimeRevenue, "USD")))
	fmt.Printf("  %s  %s\n", bold("Total Revenue:  "), cyan(models.FormatAmount(report.TotalRevenue, "USD")))
	fmt.Println()

	// Growth rate
	prevMRR, err := db.GetPreviousMonthMRR(month)
	if err == nil && prevMRR > 0 {
		growthRate := float64(report.RecurringRevenue-prevMRR) / float64(prevMRR) * 100
		growthStr := fmt.Sprintf("%+.1f%%", growthRate)
		if growthRate > 0 {
			fmt.Printf("  %s       %s (vs prev month MRR: %s)\n", bold("Growth Rate:"), green(growthStr), models.FormatAmount(prevMRR, "USD"))
		} else if growthRate < 0 {
			fmt.Printf("  %s       %s (vs prev month MRR: %s)\n", bold("Growth Rate:"), red(growthStr), models.FormatAmount(prevMRR, "USD"))
		} else {
			fmt.Printf("  %s       %s (vs prev month MRR: %s)\n", bold("Growth Rate:"), yellow(growthStr), models.FormatAmount(prevMRR, "USD"))
		}
		fmt.Println()
	}

	// Breakdown by source
	if len(report.BySource) > 0 {
		fmt.Printf("  %s\n", bold("Breakdown by Source:"))
		fmt.Println()

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Source", "Amount", "% of Total"})
		table.SetBorder(false)
		table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT})
		table.SetHeaderColor(
			tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
			tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
			tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		)

		// Sort sources for consistent output
		sources := make([]string, 0, len(report.BySource))
		for source := range report.BySource {
			sources = append(sources, source)
		}
		sort.Slice(sources, func(i, j int) bool {
			return report.BySource[sources[i]] > report.BySource[sources[j]]
		})

		for _, source := range sources {
			amount := report.BySource[source]
			pct := float64(amount) / float64(report.TotalRevenue) * 100

			table.Rich([]string{
				source,
				models.FormatAmount(amount, "USD"),
				fmt.Sprintf("%.1f%%", pct),
			}, []tablewriter.Colors{
				{tablewriter.FgMagentaColor},
				{tablewriter.FgGreenColor},
				{},
			})
		}

		table.Render()
	}

	fmt.Printf("\n  %s %d entries\n\n", bold("Total Entries:"), report.EntryCount)

	return nil
}
