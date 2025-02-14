package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

type Output struct {
	Columns []string         `json:"columns"`
	Rows    []map[string]any `json:"rows"`
}

// ListDatabaseTableRows lists all rows from the specified table using RunDatabaseCommand and returns the JSON output directly.
func ListDatabaseTableRows(ctx context.Context, dbFile *os.File, table string) (string, error) {
	if table == "" {
		return "", fmt.Errorf("table name cannot be empty")
	}

	// Get column names using PRAGMA
	columnsQuery := fmt.Sprintf("PRAGMA table_info(%q);", table)
	columnsOutput, err := RunDatabaseCommand(ctx, dbFile, columnsQuery, "-json")
	if err != nil {
		return "", fmt.Errorf("error getting columns for table %q: %w", table, err)
	}

	// Parse column information
	var columnInfo []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(columnsOutput), &columnInfo); err != nil {
		return "", fmt.Errorf("error parsing column information: %w", err)
	}

	columns := make([]string, len(columnInfo))
	for i, col := range columnInfo {
		columns[i] = col.Name
	}

	// Get all rows
	rowsQuery := fmt.Sprintf("SELECT * FROM %q;", table)
	rowsOutput, err := RunDatabaseCommand(ctx, dbFile, rowsQuery, "-json")
	if err != nil {
		return "", fmt.Errorf("error executing query for table %q: %w", table, err)
	}

	// Parse rows
	var rows []map[string]any
	if rowsOutput != "" {
		if err := json.Unmarshal([]byte(rowsOutput), &rows); err != nil {
			return "", fmt.Errorf("error parsing JSON output: %w", err)
		}
	}

	// Create and marshal output
	output := Output{
		Columns: columns,
		Rows:    rows,
	}

	result, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("error marshaling output: %w", err)
	}

	return string(result), nil
}
