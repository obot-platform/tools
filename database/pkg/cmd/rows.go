package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Output struct {
	Columns []string         `json:"columns"`
	Rows    []map[string]any `json:"rows"`
}

// ListDatabaseTableRows lists all rows from the specified table using RunDatabaseCommand and returns a JSON object containing the results.
func ListDatabaseTableRows(ctx context.Context, dbFile *os.File, table string) (string, error) {
	if table == "" {
		return "", fmt.Errorf("table name cannot be empty")
	}

	// Build the query to fetch all rows from the table
	query := fmt.Sprintf("SELECT * FROM %q;", table)

	// Execute the query using RunDatabaseCommand
	rawOutput, err := RunDatabaseCommand(ctx, dbFile, fmt.Sprintf("-header %q", query))
	if err != nil {
		return "", fmt.Errorf("error executing query for table %q: %w", table, err)
	}

	// Split raw output into rows
	lines := strings.Split(strings.TrimSpace(rawOutput), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("no output from query for table %q", table)
	}

	// The first line contains column names
	columns := strings.Split(lines[0], "|")
	output := Output{
		Columns: columns,
		Rows:    []map[string]any{},
	}

	// Process the remaining lines as rows
	for _, line := range lines[1:] {
		values := strings.Split(line, "|")
		rowData := map[string]any{}
		for i, col := range columns {
			if i < len(values) {
				rowData[col] = values[i]
			} else {
				rowData[col] = nil
			}
		}
		output.Rows = append(output.Rows, rowData)
	}

	// Marshal the result to JSON
	content, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("error marshalling output to JSON: %w", err)
	}

	return string(content), nil
}
