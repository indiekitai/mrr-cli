# mrr-cli

A terminal MRR (Monthly Recurring Revenue) tracker for indie hackers. Track your revenue from Stripe, Gumroad, Paddle, or manual entries — all from the command line.

![License](https://img.shields.io/badge/license-MIT-blue.svg)

## Features

- 📊 **Track MRR** from multiple sources (Stripe, Gumroad, Paddle, manual)
- 💰 **Currency formatting** in USD (configurable)
- 📈 **Growth rate calculation** vs previous month
- 💵 **ARR & Valuation** estimates with configurable multiplier
- 🔮 **Forecasting** - project future MRR and milestones
- 📤 **CSV Import/Export** for data portability
- 🤖 **Agent-friendly** JSON output for automation
- 🏷️ **Status badges** for README files
- 🎨 **Pretty colored output** with table formatting
- 🖥️ **Interactive TUI** with vim-style keybindings
- 🗄️ **SQLite storage** - single file at `~/.mrr-cli/data.db`
- ⚡ **Single binary** - no dependencies

## Installation

### From Source

```bash
git clone https://github.com/indiekitai/mrr-cli.git
cd mrr-cli
go build -o mrr .
sudo mv mrr /usr/local/bin/
```

### Go Install

```bash
go install github.com/indiekitai/mrr-cli@latest
```

## Usage

### Add Revenue Entry

```bash
# Basic add (defaults to manual source, recurring type)
mrr add 29.99

# With options
mrr add 99.00 --source stripe --type recurring
mrr add 49.99 --source gumroad --note "Lifetime license" --type one-time
mrr add 100 --date 2024-01-15
```

**Options:**
- `--source, -s`: Revenue source (`stripe`, `gumroad`, `paddle`, `manual`)
- `--type, -t`: Revenue type (`recurring`, `one-time`)
- `--note, -n`: Note for this entry
- `--date, -d`: Date in YYYY-MM-DD format (defaults to today)

### List Entries

```bash
# List all entries
mrr list

# Filter by month
mrr list --month 2024-01

# Filter by source or type
mrr list --source stripe
mrr list --type recurring

# JSON output for automation
mrr list --json
```

### Edit Entry

```bash
mrr edit 1 --amount 49.99
mrr edit 1 --source stripe
mrr edit 1 --note "Updated note"
mrr edit 1 --amount 99 --source gumroad
```

### Delete Entry

```bash
mrr delete 1        # With confirmation
mrr delete 1 -f     # Force (skip confirmation)
```

### Generate Report

```bash
# Current month report
mrr report

# Specific month
mrr report --month 2024-01

# Custom valuation multiplier (default 3x)
mrr report --multiplier 5

# JSON output for automation
mrr report --json

# Quiet mode - just the MRR number
mrr report --quiet
```

The report shows:
- **MRR** (Monthly Recurring Revenue)
- **ARR** (Annual Recurring Revenue = MRR × 12)
- **Growth rate** vs previous month
- **Valuation** estimate (ARR × multiplier)
- One-time revenue
- Breakdown by source

Example output:
```
Monthly Report: February 2026
───────────────────────────────────

MRR:        $1,234.00
ARR:        $14,808.00
Growth:     +15.2% vs last month
Valuation:  $44,424.00 (at 3x ARR)

By Source:
    stripe:   $800.00  (64.8%)
    gumroad:  $434.00  (35.2%)
```

### CSV Export

```bash
# Export all entries to stdout
mrr export

# Export specific month
mrr export --month 2024-01

# Export to file
mrr export --output entries.csv

# Export as JSON
mrr export --json
```

CSV format:
```csv
date,amount,source,type,note
2024-01-01,49.99,stripe,recurring,SaaS subscription
2024-01-15,19.00,gumroad,one-time,ebook sale
```

### CSV Import

```bash
mrr import entries.csv
```

Import from a CSV file with the same format as export.

### Forecast Future MRR

```bash
# Show projections and milestones
mrr forecast

# JSON output
mrr forecast --json
```

Shows projected MRR for 3, 6, and 12 months based on current growth rate, plus estimated time to reach revenue milestones ($1k, $5k, $10k, $50k, $100k MRR).

Example output:
```
MRR Forecast (based on 15.2% monthly growth)
────────────────────────────────────────────

Current:      $1,234.00
In 3 months:  $1,890.00
In 6 months:  $2,895.00
In 12 months: $6,786.00

Milestones:
  $5,000 MRR: ~8 months (Oct 2026)
  $10,000 MRR: ~13 months (Mar 2027)
  $50,000 MRR: ~25 months (Mar 2028)
```

### Generate Badge

```bash
# Output SVG to stdout
mrr badge

# Save to file
mrr badge --output mrr.svg
```

Generates a shields.io-style SVG badge showing current MRR. Perfect for README files!

### Interactive TUI

```bash
mrr tui
```

**Keybindings:**
| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `g` | Go to first entry |
| `G` | Go to last entry |
| `a` | Add new entry |
| `e` | Edit selected entry |
| `d` | Delete selected entry |
| `r` | Refresh |
| `q` / `Esc` | Quit |

## Agent-Friendly Output

All listing and report commands support `--json` / `-j` for machine-readable output:

```bash
# Get MRR as a single number
mrr report --quiet
# Output: 1234.00

# Full report as JSON
mrr report --json

# List entries as JSON
mrr list --json

# Export as JSON
mrr export --json
```

This makes it easy to integrate with scripts, automation tools, or AI agents.

## Data Storage

All data is stored locally in SQLite at `~/.mrr-cli/data.db`.

### Schema

```sql
CREATE TABLE entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    amount INTEGER NOT NULL,        -- Amount in cents
    source TEXT NOT NULL,           -- stripe, gumroad, paddle, manual
    type TEXT NOT NULL,             -- recurring, one-time
    note TEXT,
    date DATE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Backup

```bash
cp ~/.mrr-cli/data.db ~/backup/mrr-backup.db
```

## Examples

### Track Stripe Subscription Revenue

```bash
# January subscriptions
mrr add 299 --source stripe --date 2024-01-01 --note "Pro plan x3"
mrr add 99 --source stripe --date 2024-01-15 --note "Basic plan x1"

# February
mrr add 398 --source stripe --date 2024-02-01 --note "Pro plan x4"

# Check growth
mrr report --month 2024-02
```

### Track Gumroad Sales

```bash
# One-time product sales
mrr add 49 --source gumroad --type one-time --note "eBook sale"
mrr add 149 --source gumroad --type one-time --note "Course bundle"
```

### Export and Backup

```bash
# Monthly export routine
mrr export --month $(date +%Y-%m) --output ~/backup/mrr-$(date +%Y-%m).csv
```

### Automation Script

```bash
#!/bin/bash
# Get current MRR for monitoring
MRR=$(mrr report --quiet)
echo "Current MRR: $MRR"

# Check if milestone reached
if (( $(echo "$MRR > 1000" | bc -l) )); then
    echo "🎉 Milestone reached: $1000 MRR!"
fi
```

## Agent/API Reference

For AI agents and automation scripts, mrr-cli provides structured output formats.

### JSON Output

All data commands support `--json` flag:

```bash
mrr list --json
mrr report --json
mrr export --json
mrr forecast --json
```

### JSON Schemas

**`mrr list --json`**
```json
[
  {
    "id": 1,
    "date": "2026-02-20",
    "amount": 99.99,
    "source": "stripe",
    "type": "recurring",
    "note": "Pro subscription"
  }
]
```

**`mrr report --json`**
```json
{
  "month": "2026-02",
  "mrr": 1234.00,
  "arr": 14808.00,
  "growth_rate": 15.2,
  "valuation": 44424.00,
  "multiplier": 3,
  "one_time": 200.00,
  "by_source": {
    "stripe": 800.00,
    "gumroad": 434.00
  }
}
```

**`mrr forecast --json`**
```json
{
  "current_mrr": 1234.00,
  "growth_rate": 15.2,
  "projections": {
    "3_months": 1890.00,
    "6_months": 2895.00,
    "12_months": 6786.00
  },
  "milestones": [
    {"target": 5000, "months": 8, "date": "2026-10"},
    {"target": 10000, "months": 13, "date": "2027-03"}
  ]
}
```

### Quiet Mode

Get just the number for scripting:

```bash
mrr report --quiet
# Output: 1234.00

# Use in scripts
MRR=$(mrr report --quiet)
if [ "$MRR" -gt 1000 ]; then echo "Milestone!"; fi
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (invalid input, database error, etc.) |

### Example: AI Agent Integration

```bash
# Get structured data for analysis
DATA=$(mrr report --json)
echo "$DATA" | jq '.mrr, .growth_rate'

# Check health
mrr report --quiet | xargs -I {} echo "MRR: ${}"
```

## License

MIT
