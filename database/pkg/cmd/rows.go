package cmd

import (
	"context"
	"fmt"
)

const tableRowsSQL = `
SELECT json_build_object(
	'columns', (
		SELECT array_agg(column_name ORDER BY ordinal_position)
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = '%s'
	),
	'rows', COALESCE((
		SELECT json_agg(row_to_json(t))
		FROM %s t
	), '[]')
)::text;
`

// ListDatabaseTableRows returns table contents with columns
func ListDatabaseTableRows(ctx context.Context, dsn string, table string) (string, error) {
	if table == "" {
		return "", fmt.Errorf("table name cannot be empty")
	}

	query := fmt.Sprintf(tableRowsSQL, table, table)
	return RunDatabaseCommand(ctx, dsn, query, "-At")
}
