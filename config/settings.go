package config

const (
	SchemaVersion = 4
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
