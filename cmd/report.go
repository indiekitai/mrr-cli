package cmd

import (
	"encoding/json"
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

var (
	reportMonth      string
	reportMultiplier float64
	reportJSON       bool
	reportQuiet      bool
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate monthly report",
	Long: `Generate a monthly revenue report with MRR, ARR, growth rate, and valuation.

Examples:
  mrr report
  mrr report --month 2024-01
  mrr report --multiplier 5        # Use 5x ARR for valuation
  mrr report --json                # Output as JSON
  mrr report --quiet               # Output only MRR number`,
	RunE: runReport,
}

func init() {
	reportCmd.Flags().StringVarP(&reportMonth, "month", "m", "", "Month to report (YYYY-MM, defaults to current)")
	reportCmd.Flags().Float64Var(&reportMultiplier, "multiplier", 3.0, "ARR multiplier for valuation (default 3x)")
	reportCmd.Flags().BoolVarP(&reportJSON, "json", "j", false, "Output as JSON")
	reportCmd.Flags().BoolVarP(&reportQuiet, "quiet", "q", false, "Output only MRR number")
}

type reportData struct {
	Month            string             `json:"month"`
	MRR              float64            `json:"mrr"`
	ARR              float64            `json:"arr"`
	OneTimeRevenue   float64            `json:"one_time_revenue"`
	TotalRevenue     float64            `json:"total_revenue"`
	GrowthRate       *float64           `json:"growth_rate,omitempty"`
	PrevMRR          *float64           `json:"prev_mrr,omitempty"`
	Valuation        float64            `json:"valuation"`
	Multiplier       float64            `json:"multiplier"`
	BySource         map[string]float64 `json:"by_source"`
	BySourcePercent  map[string]float64 `json:"by_source_percent"`
	EntryCount       int                `json:"entry_count"`
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

	mrr := float64(report.RecurringRevenue) / 100.0
	arr := mrr * 12
	valuation := arr * reportMultiplier

	// Quiet mode - just output MRR number
	if reportQuiet {
		fmt.Printf("%.2f\n", mrr)
		return nil
	}

	// Build report data
	data := reportData{
		Month:           month,
		MRR:             mrr,
		ARR:             arr,
		OneTimeRevenue:  float64(report.OneTimeRevenue) / 100.0,
		TotalRevenue:    float64(report.TotalRevenue) / 100.0,
		Valuation:       valuation,
		Multiplier:      reportMultiplier,
		BySource:        make(map[string]float64),
		BySourcePercent: make(map[string]float64),
		EntryCount:      report.EntryCount,
	}

	for source, amount := range report.BySource {
		data.BySource[source] = float64(amount) / 100.0
		if report.TotalRevenue > 0 {
			data.BySourcePercent[source] = float64(amount) / float64(report.TotalRevenue) * 100
		}
	}

	// Growth rate calculation
	prevMRR, err := db.GetPreviousMonthMRR(month)
	if err == nil && prevMRR > 0 {
		growthRate := float64(report.RecurringRevenue-prevMRR) / float64(prevMRR) * 100
		data.GrowthRate = &growthRate
		prevMRRFloat := float64(prevMRR) / 100.0
		data.PrevMRR = &prevMRRFloat
	}

	if reportJSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(data)
	}

	return printReport(data, report)
}

func printReport(data reportData, report *db.MonthlyReport) error {
	// Colors
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	// Parse month for display
	t, _ := time.Parse("2006-01", data.Month)
	monthDisplay := t.Format("January 2006")

	// Header
	fmt.Println()
	fmt.Printf("  %s\n", cyan(fmt.Sprintf("Monthly Report: %s", monthDisplay)))
	fmt.Println("  " + strings.Repeat("─", 35))
	fmt.Println()

	if report.EntryCount == 0 {
		fmt.Printf("  %s No entries for this month.\n\n", yellow("⚠"))
		return nil
	}

	// Main metrics
	fmt.Printf("  %s        %s\n", bold("MRR:"), green(models.FormatAmount(int64(data.MRR*100), "USD")))
	fmt.Printf("  %s        %s\n", bold("ARR:"), green(models.FormatAmount(int64(data.ARR*100), "USD")))

	// Growth rate
	if data.GrowthRate != nil {
		growthStr := fmt.Sprintf("%+.1f%% vs last month", *data.GrowthRate)
		if *data.GrowthRate > 0 {
			fmt.Printf("  %s     %s\n", bold("Growth:"), green(growthStr))
		} else if *data.GrowthRate < 0 {
			fmt.Printf("  %s     %s\n", bold("Growth:"), red(growthStr))
		} else {
			fmt.Printf("  %s     %s\n", bold("Growth:"), yellow(growthStr))
		}
	}

	// Valuation
	fmt.Printf("  %s  %s (at %.0fx ARR)\n", bold("Valuation:"), cyan(models.FormatAmount(int64(data.Valuation*100), "USD")), data.Multiplier)
	fmt.Println()

	// One-time revenue if exists
	if data.OneTimeRevenue > 0 {
		fmt.Printf("  %s %s\n", bold("One-time:"), yellow(models.FormatAmount(int64(data.OneTimeRevenue*100), "USD")))
		fmt.Println()
	}

	// Breakdown by source
	if len(data.BySource) > 0 {
		fmt.Printf("  %s\n", bold("By Source:"))

		table := tablewriter.NewWriter(os.Stdout)
		table.SetBorder(false)
		table.SetColumnSeparator("")
		table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT})

		// Sort sources for consistent output
		sources := make([]string, 0, len(data.BySource))
		for source := range data.BySource {
			sources = append(sources, source)
		}
		sort.Slice(sources, func(i, j int) bool {
			return data.BySource[sources[i]] > data.BySource[sources[j]]
		})

		for _, source := range sources {
			amount := data.BySource[source]
			pct := data.BySourcePercent[source]

			table.Rich([]string{
				"    " + source + ":",
				models.FormatAmount(int64(amount*100), "USD"),
				fmt.Sprintf("(%.1f%%)", pct),
			}, []tablewriter.Colors{
				{tablewriter.FgMagentaColor},
				{tablewriter.FgGreenColor},
				{},
			})
		}

		table.Render()
	}

	fmt.Println()

	return nil
}
