package cmd

import (
	"context"
	"fmt"
)

const listTablesSQL = `SELECT COALESCE(
	json_agg(json_build_object('name', table_name)), '[]'
)::text
FROM (
	SELECT table_name
	FROM information_schema.tables
	WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
	ORDER BY table_name
) AS ordered_tables;`

// ListDatabaseTables returns a JSON string containing the list of tables
func ListDatabaseTables(ctx context.Context, dsn string) (string, error) {
	output, err := RunDatabaseCommand(ctx, dsn, listTablesSQL, "-At")
	if err != nil {
		return "", fmt.Errorf("error listing tables: %w", err)
	}

	if output == "" {
		return `{"tables":[]}`, nil
	}

	return fmt.Sprintf(`{"tables":%s}`, output), nil
}
