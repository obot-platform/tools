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
	builder.WriteString(`# START INSTRUCTIONS: Run Database Command tool

You have access to tools for interacting with a SQLite database.
The "Run Database Command" tool is a wrapper around the sqlite3 CLI, and the "sqlite3_args" argument will be passed to it.
Do not include a database file argument in the "sqlite3_args" argument. The database file is automatically passed to the sqlite3 CLI.
**Ensure that all SQL statements are properly encapsulated in double quotes** to be recognized as complete inputs by the SQLite interface.
For example, use "\"CREATE TABLE example (id INTEGER);\"", not CREATE TABLE example (id INTEGER);.
This means you should wrap the entire SQL command string in double quotes, ensuring it is treated as a single argument.
Display all results from these tools and their schemas in markdown format.
If the user refers to creating or modifying tables, assume they mean a SQLite3 table and not writing a table in a markdown file.

# END INSTRUCTIONS: Run Database Command tool
`)

	// Add the schemas section
	schemas, err := getSchemas(ctx, dbFile)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve schemas: %w", err)
	}
	if schemas != "" {
		builder.WriteString("# START CURRENT DATABASE SCHEMAS\n")
		builder.WriteString(schemas)
		builder.WriteString("\n# END CURRENT DATABASE SCHEMAS\n")
	} else {
		builder.WriteString("# DATABASE HAS NO TABLES\n")
	}

	return builder.String(), nil
}

// getSchemas retrieves all schemas from the database using the sqlite3 CLI.
func getSchemas(ctx context.Context, dbFile *os.File) (string, error) {
	query := `SELECT sql FROM sqlite_master WHERE type IN ('table', 'index', 'view', 'trigger') AND name NOT LIKE 'sqlite_%' ORDER BY name;`

	// Execute the query using the RunDatabaseCommand function
	output, err := RunDatabaseCommand(ctx, dbFile, fmt.Sprintf("%q", query))
	if err != nil {
		return "", fmt.Errorf("error querying schemas: %w", err)
	}

	// Return raw output as-is
	return strings.TrimSpace(output), nil
}
