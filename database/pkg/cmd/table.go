package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type tables struct {
	Tables []Table `json:"tables"`
}

type Table struct {
	Name string `json:"name,omitempty"`
}

// ListDatabaseTables returns a JSON object containing the list of tables in the database.
func ListDatabaseTables(ctx context.Context, dbFile *os.File) (string, error) {
	tables, err := listTables(ctx, dbFile)
	if err != nil {
		return "", fmt.Errorf("failed to list tables: %w", err)
	}

	content, err := json.Marshal(tables)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tables to JSON: %w", err)
	}

	return string(content), nil
}

// listTables retrieves the list of tables in the database using RunDatabaseCommand.
func listTables(ctx context.Context, dbFile *os.File) (tables, error) {
	// Query to fetch table names
	query := "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%';"

	// Execute the query using RunDatabaseCommand
	rawOutput, err := RunDatabaseCommand(ctx, dbFile, fmt.Sprintf("%q", query))
	if err != nil {
		return tables{}, fmt.Errorf("error executing query to list tables: %w", err)
	}

	// Process the output
	lines := strings.Split(strings.TrimSpace(rawOutput), "\n")
	if len(lines) == 0 {
		return tables{}, nil // No tables found
	}

	var result tables
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			result.Tables = append(result.Tables, Table{Name: line})
		}
	}

	return result, nil
}
