package models

import (
	"fmt"
	"time"
)

// Entry represents a revenue entry
type Entry struct {
	ID        int64
	Amount    int64  // Amount in cents
	Source    string // stripe, gumroad, paddle, manual
	Type      string // recurring, one-time
	Note      string
	Date      time.Time
	CreatedAt time.Time
}

// ValidSources contains all valid source values
var ValidSources = []string{"stripe", "gumroad", "paddle", "manual"}

// ValidTypes contains all valid type values
var ValidTypes = []string{"recurring", "one-time"}

// IsValidSource checks if a source is valid
func IsValidSource(source string) bool {
	for _, s := range ValidSources {
		if s == source {
			return true
		}
	}
	return false
}

// IsValidType checks if a type is valid
func IsValidType(t string) bool {
	for _, vt := range ValidTypes {
		if vt == t {
			return true
		}
	}
	return false
}

// FormatAmount formats cents as currency string
func FormatAmount(cents int64, currency string) string {
	dollars := float64(cents) / 100.0
	switch currency {
	case "EUR":
		return fmt.Sprintf("€%.2f", dollars)
	case "GBP":
		return fmt.Sprintf("£%.2f", dollars)
	default:
		return fmt.Sprintf("$%.2f", dollars)
	}
}
