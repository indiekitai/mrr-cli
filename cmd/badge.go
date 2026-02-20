package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/indiekitai/mrr-cli/db"
	"github.com/indiekitai/mrr-cli/models"
)

var badgeOutput string

var badgeCmd = &cobra.Command{
	Use:   "badge",
	Short: "Generate SVG badge showing current MRR",
	Long: `Generate an SVG badge displaying current MRR.
Similar to shields.io badges, suitable for README files.

Examples:
  mrr badge                      # Output to stdout
  mrr badge --output mrr.svg     # Save to file`,
	RunE: runBadge,
}

func init() {
	badgeCmd.Flags().StringVarP(&badgeOutput, "output", "o", "", "Output file path")
}

func runBadge(cmd *cobra.Command, args []string) error {
	currentMonth := time.Now().Format("2006-01")

	report, err := db.GetMonthlyReport(currentMonth)
	if err != nil {
		return err
	}

	mrrStr := models.FormatAmount(report.RecurringRevenue, "USD")
	svg := generateBadgeSVG("MRR", mrrStr)

	if badgeOutput != "" {
		err := os.WriteFile(badgeOutput, []byte(svg), 0644)
		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Printf("%s Badge saved to %s\n", green("✓"), badgeOutput)
		return nil
	}

	fmt.Print(svg)
	return nil
}

func generateBadgeSVG(label, value string) string {
	// Calculate widths based on text length
	labelWidth := len(label)*7 + 10
	valueWidth := len(value)*7 + 10
	totalWidth := labelWidth + valueWidth

	// Shield.io style SVG badge
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20" role="img" aria-label="%s: %s">
  <title>%s: %s</title>
  <linearGradient id="s" x2="0" y2="100%%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <clipPath id="r">
    <rect width="%d" height="20" rx="3" fill="#fff"/>
  </clipPath>
  <g clip-path="url(#r)">
    <rect width="%d" height="20" fill="#555"/>
    <rect x="%d" width="%d" height="20" fill="#4c1"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110">
    <text aria-hidden="true" x="%d" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="%d">%s</text>
    <text x="%d" y="140" transform="scale(.1)" fill="#fff" textLength="%d">%s</text>
    <text aria-hidden="true" x="%d" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="%d">%s</text>
    <text x="%d" y="140" transform="scale(.1)" fill="#fff" textLength="%d">%s</text>
  </g>
</svg>`,
		totalWidth, label, value,
		label, value,
		totalWidth,
		labelWidth,
		labelWidth, valueWidth,
		(labelWidth*10)/2, (labelWidth-10)*10, label,
		(labelWidth*10)/2, (labelWidth-10)*10, label,
		labelWidth*10+(valueWidth*10)/2, (valueWidth-10)*10, value,
		labelWidth*10+(valueWidth*10)/2, (valueWidth-10)*10, value,
	)
}
