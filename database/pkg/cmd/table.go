package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

type tables struct {
	Tables []Table `json:"tables"`
}

type Table struct {
	Name string `json:"name,omitempty"`
}

// ListDatabaseTables returns a JSON string containing the list of tables in the database.
func ListDatabaseTables(ctx context.Context, dbFile *os.File) (string, error) {
	// Query to fetch table names
	query := "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%';"

	// Execute the query using RunDatabaseCommand with JSON output
	output, err := RunDatabaseCommand(ctx, dbFile, query, "-json")
	if err != nil {
		return "", fmt.Errorf("error executing query to list tables: %w", err)
	}

	var dbTables tables
	if output != "" {
		if err := json.Unmarshal([]byte(output), &(dbTables.Tables)); err != nil {
			return "", fmt.Errorf("error parsing table names: %w", err)
		}
	}

	// Marshal final result
	data, err := json.Marshal(dbTables)
	if err != nil {
		return "", fmt.Errorf("error marshaling tables to JSON: %w", err)
	}

	return string(data), nil
}
