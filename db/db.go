package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/indiekitai/mrr-cli/models"
)

var db *sql.DB

// Init initializes the database connection and creates tables
func Init() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	dataDir := filepath.Join(homeDir, ".mrr-cli")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "data.db")
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Create tables
	schema := `
	CREATE TABLE IF NOT EXISTS entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		amount INTEGER NOT NULL,
		source TEXT NOT NULL DEFAULT 'manual',
		type TEXT NOT NULL DEFAULT 'recurring',
		note TEXT,
		date DATE NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_entries_date ON entries(date);
	CREATE INDEX IF NOT EXISTS idx_entries_source ON entries(source);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// Close closes the database connection
func Close() {
	if db != nil {
		db.Close()
	}
}

// AddEntry adds a new revenue entry
func AddEntry(amount int64, source, entryType, note string, date time.Time) (int64, error) {
	result, err := db.Exec(
		"INSERT INTO entries (amount, source, type, note, date) VALUES (?, ?, ?, ?, ?)",
		amount, source, entryType, note, date.Format("2006-01-02"),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to add entry: %w", err)
	}
	return result.LastInsertId()
}

// GetEntry retrieves a single entry by ID
func GetEntry(id int64) (*models.Entry, error) {
	row := db.QueryRow(
		"SELECT id, amount, source, type, note, date, created_at FROM entries WHERE id = ?",
		id,
	)

	var entry models.Entry
	var dateStr string
	var createdAtStr string
	var note sql.NullString

	err := row.Scan(&entry.ID, &entry.Amount, &entry.Source, &entry.Type, &note, &dateStr, &createdAtStr)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("entry not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	entry.Date = parseDate(dateStr)
	entry.CreatedAt = parseDateTime(createdAtStr)
	if note.Valid {
		entry.Note = note.String
	}

	return &entry, nil
}

// ListEntries lists entries with optional filters
func ListEntries(month string, source string, entryType string) ([]models.Entry, error) {
	query := "SELECT id, amount, source, type, note, date, created_at FROM entries WHERE 1=1"
	args := []interface{}{}

	if month != "" {
		query += " AND strftime('%Y-%m', date) = ?"
		args = append(args, month)
	}
	if source != "" {
		query += " AND source = ?"
		args = append(args, source)
	}
	if entryType != "" {
		query += " AND type = ?"
		args = append(args, entryType)
	}

	query += " ORDER BY date DESC, id DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list entries: %w", err)
	}
	defer rows.Close()

	var entries []models.Entry
	for rows.Next() {
		var entry models.Entry
		var dateStr string
		var createdAtStr string
		var note sql.NullString

		err := rows.Scan(&entry.ID, &entry.Amount, &entry.Source, &entry.Type, &note, &dateStr, &createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		entry.Date = parseDate(dateStr)
		entry.CreatedAt = parseDateTime(createdAtStr)
		if note.Valid {
			entry.Note = note.String
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// UpdateEntry updates an existing entry
func UpdateEntry(id int64, amount *int64, source, note *string) error {
	// First check if entry exists
	_, err := GetEntry(id)
	if err != nil {
		return err
	}

	updates := []string{}
	args := []interface{}{}

	if amount != nil {
		updates = append(updates, "amount = ?")
		args = append(args, *amount)
	}
	if source != nil {
		updates = append(updates, "source = ?")
		args = append(args, *source)
	}
	if note != nil {
		updates = append(updates, "note = ?")
		args = append(args, *note)
	}

	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := "UPDATE entries SET "
	for i, u := range updates {
		if i > 0 {
			query += ", "
		}
		query += u
	}
	query += " WHERE id = ?"
	args = append(args, id)

	_, err = db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}

	return nil
}

// DeleteEntry deletes an entry by ID
func DeleteEntry(id int64) error {
	result, err := db.Exec("DELETE FROM entries WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("entry not found: %d", id)
	}

	return nil
}

// MonthlyReport contains aggregated data for a month
type MonthlyReport struct {
	Month           string
	TotalRevenue    int64
	RecurringRevenue int64
	OneTimeRevenue  int64
	BySource        map[string]int64
	EntryCount      int
}

// GetMonthlyReport generates a report for a specific month
func GetMonthlyReport(month string) (*MonthlyReport, error) {
	report := &MonthlyReport{
		Month:    month,
		BySource: make(map[string]int64),
	}

	// Get entries for the month
	entries, err := ListEntries(month, "", "")
	if err != nil {
		return nil, err
	}

	report.EntryCount = len(entries)

	for _, e := range entries {
		report.TotalRevenue += e.Amount
		if e.Type == "recurring" {
			report.RecurringRevenue += e.Amount
		} else {
			report.OneTimeRevenue += e.Amount
		}
		report.BySource[e.Source] += e.Amount
	}

	return report, nil
}

// GetPreviousMonthMRR gets the MRR for the previous month
func GetPreviousMonthMRR(currentMonth string) (int64, error) {
	t, err := time.Parse("2006-01", currentMonth)
	if err != nil {
		return 0, fmt.Errorf("invalid month format: %w", err)
	}

	prevMonth := t.AddDate(0, -1, 0).Format("2006-01")
	
	var total int64
	err = db.QueryRow(
		"SELECT COALESCE(SUM(amount), 0) FROM entries WHERE strftime('%Y-%m', date) = ? AND type = 'recurring'",
		prevMonth,
	).Scan(&total)
	
	return total, err
}

// GetAllEntries returns all entries (for TUI)
func GetAllEntries() ([]models.Entry, error) {
	return ListEntries("", "", "")
}

// parseDate parses various date formats from SQLite
func parseDate(s string) time.Time {
	// Try different formats
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		time.RFC3339,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// parseDateTime parses various datetime formats from SQLite
func parseDateTime(s string) time.Time {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		time.RFC3339,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
