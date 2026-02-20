package cmd

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/indiekitai/mrr-cli/db"
	"github.com/indiekitai/mrr-cli/models"
)

var (
	servePort   int
	servePublic bool
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve a public MRR dashboard",
	Long: `Start a web server with a beautiful MRR dashboard.

The dashboard shows:
- Current MRR with big number
- Trend chart (last 6 months)
- Goal progress bar (if set)
- Last updated timestamp

Examples:
  mrr serve                   # Serve on port 8080
  mrr serve --port 3000       # Custom port
  mrr serve --public          # Public mode (hides sensitive details)`,
	RunE: runServe,
}

func init() {
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "Port to listen on")
	serveCmd.Flags().BoolVar(&servePublic, "public", false, "Public mode (hides entry details)")
}

type dashboardData struct {
	CurrentMRR    float64            `json:"current_mrr"`
	ARR           float64            `json:"arr"`
	GrowthRate    *float64           `json:"growth_rate,omitempty"`
	MonthlyTrend  []monthlyDataPoint `json:"monthly_trend"`
	Goal          *goalData          `json:"goal,omitempty"`
	IsPublic      bool               `json:"is_public"`
	LastUpdated   string             `json:"last_updated"`
}

type monthlyDataPoint struct {
	Month string  `json:"month"`
	MRR   float64 `json:"mrr"`
}

type goalData struct {
	Amount   float64 `json:"amount"`
	Deadline string  `json:"deadline,omitempty"`
	Progress float64 `json:"progress"`
}

func runServe(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()

	fmt.Println()
	fmt.Printf("  %s MRR Dashboard\n", cyan("📊"))
	fmt.Println("  " + strings.Repeat("─", 35))
	fmt.Printf("  %s http://localhost:%d\n", green("→"), servePort)
	if servePublic {
		fmt.Println("  Mode: Public (hiding entry details)")
	}
	fmt.Println()
	fmt.Println("  Press Ctrl+C to stop")
	fmt.Println()

	http.HandleFunc("/", handleDashboard)
	http.HandleFunc("/api/data", handleAPIData)

	return http.ListenAndServe(fmt.Sprintf(":%d", servePort), nil)
}

func getDashboardData() (*dashboardData, error) {
	currentMonth := time.Now().Format("2006-01")

	// Get current MRR
	report, err := db.GetMonthlyReport(currentMonth)
	if err != nil {
		return nil, err
	}

	currentMRR := float64(report.RecurringRevenue) / 100.0

	data := &dashboardData{
		CurrentMRR:  currentMRR,
		ARR:         currentMRR * 12,
		IsPublic:    servePublic,
		LastUpdated: time.Now().Format(time.RFC3339),
	}

	// Growth rate
	prevMRR, err := db.GetPreviousMonthMRR(currentMonth)
	if err == nil && prevMRR > 0 {
		growthRate := float64(report.RecurringRevenue-prevMRR) / float64(prevMRR) * 100
		data.GrowthRate = &growthRate
	}

	// Get last 6 months of data
	data.MonthlyTrend = []monthlyDataPoint{}
	for i := 5; i >= 0; i-- {
		monthDate := time.Now().AddDate(0, -i, 0)
		monthStr := monthDate.Format("2006-01")

		monthReport, err := db.GetMonthlyReport(monthStr)
		if err != nil {
			continue
		}

		data.MonthlyTrend = append(data.MonthlyTrend, monthlyDataPoint{
			Month: monthStr,
			MRR:   float64(monthReport.RecurringRevenue) / 100.0,
		})
	}

	// Goal
	goal, err := GetGoal()
	if err == nil && goal != nil {
		progress := currentMRR / (float64(goal.Amount) / 100.0) * 100
		if progress > 100 {
			progress = 100
		}

		data.Goal = &goalData{
			Amount:   float64(goal.Amount) / 100.0,
			Deadline: goal.Deadline,
			Progress: progress,
		}
	}

	return data, nil
}

func handleAPIData(w http.ResponseWriter, r *http.Request) {
	data, err := getDashboardData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(data)
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	data, err := getDashboardData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, generateHTML(data))
}

func generateHTML(data *dashboardData) string {
	// Format numbers
	mrrFormatted := models.FormatAmount(int64(data.CurrentMRR*100), "USD")
	arrFormatted := models.FormatAmount(int64(data.ARR*100), "USD")

	// Growth badge
	var growthBadge string
	if data.GrowthRate != nil {
		if *data.GrowthRate > 0 {
			growthBadge = fmt.Sprintf(`<span class="badge green">↑ %.1f%%</span>`, *data.GrowthRate)
		} else if *data.GrowthRate < 0 {
			growthBadge = fmt.Sprintf(`<span class="badge red">↓ %.1f%%</span>`, *data.GrowthRate)
		} else {
			growthBadge = `<span class="badge gray">→ 0%</span>`
		}
	}

	// Goal section
	var goalHTML string
	if data.Goal != nil {
		goalFormatted := models.FormatAmount(int64(data.Goal.Amount*100), "USD")
		deadlineStr := ""
		if data.Goal.Deadline != "" {
			deadline, _ := time.Parse("2006-01", data.Goal.Deadline)
			deadlineStr = fmt.Sprintf(" by %s", deadline.Format("January 2006"))
		}
		goalHTML = fmt.Sprintf(`
		<div class="goal-section">
			<div class="goal-header">🎯 Goal: %s MRR%s</div>
			<div class="progress-bar">
				<div class="progress-fill" style="width: %.1f%%"></div>
			</div>
			<div class="goal-stats">%.1f%% complete</div>
		</div>`, html.EscapeString(goalFormatted), html.EscapeString(deadlineStr), data.Goal.Progress, data.Goal.Progress)
	}

	// Chart data
	chartLabels := []string{}
	chartData := []string{}
	maxMRR := 0.0
	for _, point := range data.MonthlyTrend {
		t, _ := time.Parse("2006-01", point.Month)
		chartLabels = append(chartLabels, fmt.Sprintf(`"%s"`, t.Format("Jan")))
		chartData = append(chartData, fmt.Sprintf("%.2f", point.MRR))
		if point.MRR > maxMRR {
			maxMRR = point.MRR
		}
	}

	// Chart bars (simple CSS bars)
	var chartBars string
	for _, point := range data.MonthlyTrend {
		t, _ := time.Parse("2006-01", point.Month)
		height := 0.0
		if maxMRR > 0 {
			height = point.MRR / maxMRR * 100
		}
		label := t.Format("Jan")
		chartBars += fmt.Sprintf(`
			<div class="chart-bar-wrapper">
				<div class="chart-bar" style="height: %.1f%%" title="%s"></div>
				<div class="chart-label">%s</div>
			</div>`, height, models.FormatAmount(int64(point.MRR*100), "USD"), label)
	}

	// Parse last updated time
	lastUpdated, _ := time.Parse(time.RFC3339, data.LastUpdated)
	lastUpdatedStr := lastUpdated.Format("Jan 2, 2006 at 3:04 PM")

	// Build recent entries section (only for non-public mode)
	var recentEntriesHTML string
	if !data.IsPublic {
		entries, err := db.ListEntries("", "", "")
		if err == nil && len(entries) > 0 {
			// Sort by date descending and take first 5
			sort.Slice(entries, func(i, j int) bool {
				return entries[i].Date.After(entries[j].Date)
			})
			
			limit := 5
			if len(entries) < limit {
				limit = len(entries)
			}
			
			entriesRows := ""
			for _, e := range entries[:limit] {
				entriesRows += fmt.Sprintf(`
				<tr>
					<td>%s</td>
					<td>%s</td>
					<td class="amount">%s</td>
					<td>%s</td>
				</tr>`, 
					e.Date.Format("Jan 2"),
					html.EscapeString(e.Source),
					models.FormatAmount(e.Amount, "USD"),
					html.EscapeString(e.Type),
				)
			}
			
			recentEntriesHTML = fmt.Sprintf(`
			<div class="section">
				<h3>Recent Entries</h3>
				<table class="entries-table">
					<thead>
						<tr><th>Date</th><th>Source</th><th>Amount</th><th>Type</th></tr>
					</thead>
					<tbody>%s</tbody>
				</table>
			</div>`, entriesRows)
		}
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>MRR Dashboard</title>
	<style>
		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
			background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
			min-height: 100vh;
			padding: 40px 20px;
			color: #1a1a2e;
		}
		.container {
			max-width: 600px;
			margin: 0 auto;
		}
		.card {
			background: white;
			border-radius: 20px;
			padding: 40px;
			box-shadow: 0 20px 60px rgba(0,0,0,0.3);
		}
		.header {
			text-align: center;
			margin-bottom: 30px;
		}
		.mrr-label {
			font-size: 14px;
			text-transform: uppercase;
			letter-spacing: 2px;
			color: #666;
			margin-bottom: 8px;
		}
		.mrr-value {
			font-size: 64px;
			font-weight: 700;
			background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
			-webkit-background-clip: text;
			-webkit-text-fill-color: transparent;
			background-clip: text;
			line-height: 1.1;
		}
		.badge {
			display: inline-block;
			padding: 4px 12px;
			border-radius: 20px;
			font-size: 14px;
			font-weight: 600;
			margin-left: 8px;
		}
		.badge.green {
			background: #d4edda;
			color: #155724;
		}
		.badge.red {
			background: #f8d7da;
			color: #721c24;
		}
		.badge.gray {
			background: #e2e8f0;
			color: #4a5568;
		}
		.arr {
			text-align: center;
			color: #666;
			font-size: 18px;
			margin-top: 10px;
		}
		.section {
			margin-top: 30px;
		}
		h3 {
			font-size: 14px;
			text-transform: uppercase;
			letter-spacing: 1px;
			color: #666;
			margin-bottom: 15px;
		}
		.chart {
			display: flex;
			align-items: flex-end;
			justify-content: space-around;
			height: 120px;
			padding: 10px 0;
			border-bottom: 2px solid #e2e8f0;
		}
		.chart-bar-wrapper {
			display: flex;
			flex-direction: column;
			align-items: center;
			flex: 1;
		}
		.chart-bar {
			width: 30px;
			background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
			border-radius: 6px 6px 0 0;
			min-height: 4px;
			transition: height 0.3s ease;
		}
		.chart-bar:hover {
			opacity: 0.8;
		}
		.chart-label {
			font-size: 12px;
			color: #666;
			margin-top: 8px;
		}
		.goal-section {
			margin-top: 30px;
			padding: 20px;
			background: #f8f9fa;
			border-radius: 12px;
		}
		.goal-header {
			font-size: 16px;
			font-weight: 600;
			margin-bottom: 15px;
		}
		.progress-bar {
			height: 12px;
			background: #e2e8f0;
			border-radius: 6px;
			overflow: hidden;
		}
		.progress-fill {
			height: 100%%;
			background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
			border-radius: 6px;
			transition: width 0.5s ease;
		}
		.goal-stats {
			font-size: 14px;
			color: #666;
			margin-top: 10px;
		}
		.entries-table {
			width: 100%%;
			border-collapse: collapse;
			font-size: 14px;
		}
		.entries-table th {
			text-align: left;
			padding: 10px 8px;
			border-bottom: 2px solid #e2e8f0;
			color: #666;
			font-weight: 500;
		}
		.entries-table td {
			padding: 10px 8px;
			border-bottom: 1px solid #f0f0f0;
		}
		.entries-table .amount {
			font-weight: 600;
			color: #155724;
		}
		.footer {
			text-align: center;
			margin-top: 30px;
			padding-top: 20px;
			border-top: 1px solid #e2e8f0;
			color: #999;
			font-size: 12px;
		}
		.footer a {
			color: #667eea;
			text-decoration: none;
		}
		@media (max-width: 480px) {
			.mrr-value {
				font-size: 48px;
			}
			.card {
				padding: 25px;
			}
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="card">
			<div class="header">
				<div class="mrr-label">Monthly Recurring Revenue</div>
				<div class="mrr-value">%s</div>
				%s
				<div class="arr">%s ARR</div>
			</div>

			<div class="section">
				<h3>Last 6 Months</h3>
				<div class="chart">
					%s
				</div>
			</div>

			%s

			%s

			<div class="footer">
				Last updated: %s<br>
				Powered by <a href="https://github.com/indiekitai/mrr-cli" target="_blank">mrr-cli</a>
			</div>
		</div>
	</div>
</body>
</html>`,
		html.EscapeString(mrrFormatted),
		growthBadge,
		html.EscapeString(arrFormatted),
		chartBars,
		goalHTML,
		recentEntriesHTML,
		html.EscapeString(lastUpdatedStr),
	)
}
