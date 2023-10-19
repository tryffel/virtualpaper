package config

import (
	"fmt"
	"time"
)

const (
	SchemaVersion = 18
)

const (
	// MaxRows is absolute maximum for returning rows / results from database
	MaxRows = 500

	// MaxRulesToProcess is per-user absolute maximum number of rules to run for each document.
	MaxRulesToProcess = 50
)

// MaxRevords returns minimum of MaxRows and n, where n might be supplied from user input.
// If n == 0, return MaxRows.
func MaxRecords(n int) int {
	if n == 0 {
		return MaxRows
	}
	if n > MaxRows {
		return MaxRows
	}
	return n
}

var startedAt time.Time

func init() {
	startedAt = time.Now()
}

func Uptime() time.Duration {
	return time.Now().Sub(startedAt)
}

func UptimeString() string {
	duration := Uptime()

	if duration > time.Hour*24*3 {
		return fmt.Sprintf("%.2f days", duration.Hours()/24)
	}

	if duration > time.Hour {
		return fmt.Sprintf("%.2f hours", duration.Hours())
	}

	return fmt.Sprintf("%0.f minutes", duration.Minutes())
}
