package cmd

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/indiekitai/mrr-cli/db"
	"github.com/indiekitai/mrr-cli/models"
)

// Config represents the config file structure
type Config struct {
	Goal *GoalConfig `json:"goal,omitempty"`
}

// GoalConfig represents a MRR goal
type GoalConfig struct {
	Amount   int64  `json:"amount"`   // Amount in cents
	Deadline string `json:"deadline"` // YYYY-MM format
	SetAt    string `json:"set_at"`   // YYYY-MM-DD format
}

var goalDeadline string

var goalCmd = &cobra.Command{
	Use:   "goal",
	Short: "Manage MRR goals",
	Long: `Set, view, and track MRR goals.

Examples:
  mrr goal set 10000                  # Set $10,000 MRR goal
  mrr goal set 10000 --by 2026-06     # Set goal with deadline
  mrr goal status                     # Show progress towards goal
  mrr goal clear                      # Remove the goal`,
}

var goalSetCmd = &cobra.Command{
	Use:   "set <amount>",
	Short: "Set a MRR goal",
	Long: `Set a MRR goal in dollars.

Examples:
  mrr goal set 10000                  # Set $10,000 MRR goal
  mrr goal set 10000 --by 2026-06     # Set goal with deadline`,
	Args: cobra.ExactArgs(1),
	RunE: runGoalSet,
}

var goalStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show progress towards goal",
	RunE:  runGoalStatus,
}

var goalClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Remove the MRR goal",
	RunE:  runGoalClear,
}

func init() {
	goalSetCmd.Flags().StringVar(&goalDeadline, "by", "", "Target deadline (YYYY-MM)")

	goalCmd.AddCommand(goalSetCmd)
	goalCmd.AddCommand(goalStatusCmd)
	goalCmd.AddCommand(goalClearCmd)
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".mrr-cli", "config.json"), nil
}

func loadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	config := &Config{}

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return config, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return config, nil
}

func saveConfig(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func runGoalSet(cmd *cobra.Command, args []string) error {
	// Parse amount
	var amount float64
	if _, err := fmt.Sscanf(args[0], "%f", &amount); err != nil {
		return fmt.Errorf("invalid amount: %s", args[0])
	}

	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	amountCents := int64(amount * 100)

	// Validate deadline if provided
	if goalDeadline != "" {
		if _, err := time.Parse("2006-01", goalDeadline); err != nil {
			return fmt.Errorf("invalid deadline format, use YYYY-MM: %s", goalDeadline)
		}
	}

	config, err := loadConfig()
	if err != nil {
		return err
	}

	config.Goal = &GoalConfig{
		Amount:   amountCents,
		Deadline: goalDeadline,
		SetAt:    time.Now().Format("2006-01-02"),
	}

	if err := saveConfig(config); err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("\n  %s Goal set: %s MRR", green("✓"), models.FormatAmount(amountCents, "USD"))
	if goalDeadline != "" {
		deadline, _ := time.Parse("2006-01", goalDeadline)
		fmt.Printf(" by %s", deadline.Format("January 2006"))
	}
	fmt.Println("\n")

	return nil
}

func runGoalStatus(cmd *cobra.Command, args []string) error {
	config, err := loadConfig()
	if err != nil {
		return err
	}

	if config.Goal == nil {
		yellow := color.New(color.FgYellow).SprintFunc()
		fmt.Printf("\n  %s No goal set. Use 'mrr goal set <amount>' to set one.\n\n", yellow("⚠"))
		return nil
	}

	// Get current MRR
	currentMonth := time.Now().Format("2006-01")
	report, err := db.GetMonthlyReport(currentMonth)
	if err != nil {
		return err
	}

	currentMRR := report.RecurringRevenue
	goalAmount := config.Goal.Amount
	progress := float64(currentMRR) / float64(goalAmount) * 100
	if progress > 100 {
		progress = 100
	}

	// Colors
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	// Header
	fmt.Println()
	goalStr := fmt.Sprintf("🎯 Goal: %s MRR", models.FormatAmount(goalAmount, "USD"))
	if config.Goal.Deadline != "" {
		deadline, _ := time.Parse("2006-01", config.Goal.Deadline)
		goalStr += fmt.Sprintf(" by %s", deadline.Format("January 2006"))
	}
	fmt.Printf("  %s\n", cyan(goalStr))
	fmt.Println("  " + strings.Repeat("━", 40))
	fmt.Println()

	// Current progress
	fmt.Printf("  %s  %s (%.1f%%)\n", bold("Current:"), green(models.FormatAmount(currentMRR, "USD")), progress)

	// Progress bar (32 chars)
	barWidth := 32
	filled := int(progress / 100 * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	fmt.Printf("  %s\n", bar)
	fmt.Println()

	// Stats
	remaining := goalAmount - currentMRR
	if remaining < 0 {
		remaining = 0
	}
	fmt.Printf("  %s %s / %s\n", bold("Progress:"), models.FormatAmount(currentMRR, "USD"), models.FormatAmount(goalAmount, "USD"))
	fmt.Printf("  %s %s\n", bold("Remaining:"), models.FormatAmount(remaining, "USD"))

	// Time left if deadline set
	if config.Goal.Deadline != "" {
		deadline, _ := time.Parse("2006-01", config.Goal.Deadline)
		// Set to end of month
		deadline = deadline.AddDate(0, 1, -1)
		now := time.Now()
		daysLeft := int(deadline.Sub(now).Hours() / 24)
		monthsLeft := int(math.Ceil(float64(daysLeft) / 30.0))

		if monthsLeft > 0 {
			fmt.Printf("  %s %d months\n", bold("Time left:"), monthsLeft)
		} else if daysLeft > 0 {
			fmt.Printf("  %s %d days\n", bold("Time left:"), daysLeft)
		} else {
			fmt.Printf("  %s %s\n", bold("Time left:"), yellow("deadline passed"))
		}
	}

	// Growth projection
	prevMRR, err := db.GetPreviousMonthMRR(currentMonth)
	if err == nil && prevMRR > 0 && currentMRR > 0 {
		growthRate := float64(currentMRR-prevMRR) / float64(prevMRR)
		monthlyGrowthPercent := growthRate * 100

		fmt.Println()
		fmt.Printf("  %s At current growth rate (%.1f%%/mo):\n", "📈", monthlyGrowthPercent)

		if growthRate > 0 && currentMRR < goalAmount {
			// months = log(goal/current) / log(1+growthRate)
			monthsToGoal := math.Log(float64(goalAmount)/float64(currentMRR)) / math.Log(1+growthRate)

			if monthsToGoal > 0 && monthsToGoal < 120 {
				projectedDate := time.Now().AddDate(0, int(math.Ceil(monthsToGoal)), 0)
				fmt.Printf("     Projected to reach goal in: %.1f months", monthsToGoal)

				// Check if within deadline
				if config.Goal.Deadline != "" {
					deadline, _ := time.Parse("2006-01", config.Goal.Deadline)
					if projectedDate.Before(deadline) || projectedDate.Format("2006-01") == config.Goal.Deadline {
						fmt.Printf(" %s\n", green("✅"))
					} else {
						fmt.Printf(" %s\n", yellow("⚠️"))
					}
				} else {
					fmt.Println()
				}

				fmt.Printf("     Expected date: %s\n", yellow(projectedDate.Format("January 2006")))
			}
		} else if currentMRR >= goalAmount {
			fmt.Printf("     %s Goal reached! 🎉\n", green("✓"))
		} else if growthRate <= 0 {
			fmt.Printf("     %s Negative growth - goal may not be reachable at current pace\n", yellow("⚠"))
		}
	}

	fmt.Println()
	return nil
}

func runGoalClear(cmd *cobra.Command, args []string) error {
	config, err := loadConfig()
	if err != nil {
		return err
	}

	if config.Goal == nil {
		yellow := color.New(color.FgYellow).SprintFunc()
		fmt.Printf("\n  %s No goal was set.\n\n", yellow("⚠"))
		return nil
	}

	config.Goal = nil

	if err := saveConfig(config); err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("\n  %s Goal cleared.\n\n", green("✓"))

	return nil
}

// GetGoal returns the current goal (exported for use by serve command)
func GetGoal() (*GoalConfig, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}
	return config.Goal, nil
}
