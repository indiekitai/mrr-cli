package cmd

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/indiekitai/mrr-cli/db"
	"github.com/indiekitai/mrr-cli/models"
)

var forecastJSON bool

var forecastCmd = &cobra.Command{
	Use:   "forecast",
	Short: "Project future MRR and milestones",
	Long: `Show projected MRR for next 3, 6, 12 months based on current growth rate.
Also shows estimated time to reach revenue milestones.

Examples:
  mrr forecast
  mrr forecast --json`,
	RunE: runForecast,
}

func init() {
	forecastCmd.Flags().BoolVarP(&forecastJSON, "json", "j", false, "Output as JSON")
}

type forecastData struct {
	CurrentMRR   float64            `json:"current_mrr"`
	GrowthRate   float64            `json:"growth_rate"`
	Projections  map[string]float64 `json:"projections"`
	Milestones   []milestoneData    `json:"milestones"`
	BasedOnMonth string             `json:"based_on_month"`
}

type milestoneData struct {
	Target        float64 `json:"target"`
	MonthsAway    int     `json:"months_away"`
	EstimatedDate string  `json:"estimated_date"`
}

func runForecast(cmd *cobra.Command, args []string) error {
	currentMonth := time.Now().Format("2006-01")

	// Get current MRR
	report, err := db.GetMonthlyReport(currentMonth)
	if err != nil {
		return err
	}

	currentMRR := float64(report.RecurringRevenue) / 100.0

	// Get previous month MRR for growth rate
	prevMRR, err := db.GetPreviousMonthMRR(currentMonth)
	if err != nil {
		prevMRR = 0
	}

	var growthRate float64
	if prevMRR > 0 {
		growthRate = float64(report.RecurringRevenue-prevMRR) / float64(prevMRR) * 100
	} else {
		growthRate = 0
	}

	// Calculate projections
	monthlyGrowth := 1 + (growthRate / 100)
	
	// Cap projections at reasonable values (1 billion)
	maxProjection := 1000000000.0
	capProjection := func(val float64) float64 {
		if val > maxProjection || math.IsInf(val, 1) || math.IsNaN(val) {
			return maxProjection
		}
		return val
	}
	
	projections := map[string]float64{
		"3_months":  capProjection(currentMRR * math.Pow(monthlyGrowth, 3)),
		"6_months":  capProjection(currentMRR * math.Pow(monthlyGrowth, 6)),
		"12_months": capProjection(currentMRR * math.Pow(monthlyGrowth, 12)),
	}

	// Calculate milestones
	milestones := []float64{1000, 5000, 10000, 50000, 100000}
	var milestoneResults []milestoneData

	for _, target := range milestones {
		if currentMRR >= target {
			continue // Skip already passed milestones
		}

		if growthRate <= 0 {
			continue // Can't reach milestones with no growth
		}

		// months = log(target/current) / log(1+growthRate)
		monthsAway := int(math.Ceil(math.Log(target/currentMRR) / math.Log(monthlyGrowth)))
		if monthsAway < 0 || monthsAway > 120 {
			continue // Skip unrealistic projections (>10 years)
		}

		estDate := time.Now().AddDate(0, monthsAway, 0)
		milestoneResults = append(milestoneResults, milestoneData{
			Target:        target,
			MonthsAway:    monthsAway,
			EstimatedDate: estDate.Format("Jan 2006"),
		})
	}

	data := forecastData{
		CurrentMRR:   currentMRR,
		GrowthRate:   growthRate,
		Projections:  projections,
		Milestones:   milestoneResults,
		BasedOnMonth: currentMonth,
	}

	if forecastJSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(data)
	}

	return printForecast(data)
}

func printForecast(data forecastData) error {
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	growthStr := fmt.Sprintf("%.1f%%", data.GrowthRate)
	if data.GrowthRate > 0 {
		growthStr = green(growthStr)
	} else if data.GrowthRate < 0 {
		growthStr = color.New(color.FgRed).Sprintf("%.1f%%", data.GrowthRate)
	} else {
		growthStr = yellow(growthStr)
	}

	fmt.Println()
	fmt.Printf("  %s\n", cyan(fmt.Sprintf("MRR Forecast (based on %s monthly growth)", growthStr)))
	fmt.Println("  " + strings.Repeat("─", 44))
	fmt.Println()

	fmt.Printf("  %s   %s\n", bold("Current:"), green(models.FormatAmount(int64(data.CurrentMRR*100), "USD")))

	if data.GrowthRate == 0 {
		fmt.Printf("\n  %s No growth data available for projections.\n", yellow("⚠"))
		fmt.Println("  Add entries across multiple months to see forecasts.")
		fmt.Println()
		return nil
	}

	fmt.Printf("  %s %s\n", bold("In 3 months:"), formatProjection(data.Projections["3_months"]))
	fmt.Printf("  %s %s\n", bold("In 6 months:"), formatProjection(data.Projections["6_months"]))
	fmt.Printf("  %s%s\n", bold("In 12 months:"), formatProjection(data.Projections["12_months"]))
	fmt.Println()

	if len(data.Milestones) > 0 {
		fmt.Printf("  %s\n", bold("Milestones:"))
		for _, m := range data.Milestones {
			fmt.Printf("    %s MRR: ~%d months (%s)\n",
				green(models.FormatAmount(int64(m.Target*100), "USD")),
				m.MonthsAway,
				yellow(m.EstimatedDate),
			)
		}
	} else if data.CurrentMRR >= 100000 {
		fmt.Printf("  %s All milestones reached! 🎉\n", green("✓"))
	} else {
		fmt.Printf("  %s Add data across multiple months to see milestone projections.\n", yellow("⚠"))
	}

	fmt.Println()
	return nil
}

func formatProjection(amount float64) string {
	return models.FormatAmount(int64(amount*100), "USD")
}
