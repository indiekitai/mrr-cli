# mrr-cli

A terminal MRR (Monthly Recurring Revenue) tracker for indie hackers. Track your revenue from Stripe, Gumroad, Paddle, or manual entries — all from the command line.

![License](https://img.shields.io/badge/license-MIT-blue.svg)

## Features

- 📊 **Track MRR** from multiple sources (Stripe, Gumroad, Paddle, manual)
- 💰 **Currency formatting** in USD (configurable)
- 📈 **Growth rate calculation** vs previous month
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
```

The report shows:
- MRR (recurring revenue)
- One-time revenue
- Total revenue
- Growth rate vs previous month
- Breakdown by source

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

## License

MIT
