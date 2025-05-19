package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// DatabaseContext generates a markdown-formatted string with instructions
// and the database's current schemas.
func DatabaseContext(ctx context.Context, dbFile *os.File) (string, error) {
	var builder strings.Builder

	// Add usage instructions
	builder.WriteString(`<sqlite_database_tools_instructions>
You have access to tools for interacting with a SQLite database.
The "Run Database SQL" tool lets you run SQL against the SQLite3 database.
Display all results from these tools and their schemas in markdown format.
If the user refers to creating or modifying tables, assume they mean a SQLite3 table and not writing a table in a markdown file.
</sqlite_database_tools_instructions>
`)

	// Add the schemas section
	schemas, err := getSchemas(ctx, dbFile)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve schemas: %w", err)
	}
	if schemas != "" {
		builder.WriteString("<current_database_schema>\n")
		builder.WriteString(schemas)
		builder.WriteString("\n</current_database_schema>\n")
	} else {
		builder.WriteString("# DATABASE HAS NO TABLES\n")
	}

	return builder.String(), nil
}

// getSchemas retrieves all schemas from the database using the sqlite3 CLI.
func getSchemas(ctx context.Context, dbFile *os.File) (string, error) {
	query := `SELECT sql FROM sqlite_master WHERE type IN ('table', 'index', 'view', 'trigger') AND name NOT LIKE 'sqlite_%' ORDER BY name;`

	// Execute the query using the RunDatabaseCommand function
	output, err := RunDatabaseCommand(ctx, dbFile, query)
	if err != nil {
		return "", fmt.Errorf("error querying schemas: %w", err)
	}

	// Return raw output as-is
	return strings.TrimSpace(output), nil
}
