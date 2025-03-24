package cmd

import (
	"context"
	"fmt"
	"strings"
)

const getSchemasSQL = `
WITH table_columns AS (
  SELECT
    table_name,
    ordinal_position,
    column_name,
    data_type,
    is_nullable,
    column_default
  FROM information_schema.columns
  WHERE table_schema = 'public'
),
constraints AS (
  SELECT
    conname,
    contype,
    conrelid::regclass::text AS table_name,
    pg_get_constraintdef(oid, true) AS definition
  FROM pg_constraint
  WHERE connamespace = 'public'::regnamespace
),
indexes AS (
  SELECT
    tablename,
    indexdef
  FROM pg_indexes
  WHERE schemaname = 'public'
)
SELECT format(
  E'\nCREATE TABLE %I (\n%s%s\n);\n\n%s\n',
  tc.table_name,
  tc.table_name,
  string_agg(
    format('  %I %s%s%s',
      tc.column_name,
      tc.data_type,
      CASE WHEN tc.column_default IS NOT NULL THEN ' DEFAULT ' || tc.column_default ELSE '' END,
      CASE WHEN tc.is_nullable = 'NO' THEN ' NOT NULL' ELSE '' END
    ),
    E',\n'
    ORDER BY tc.ordinal_position
  ),
  CASE
    WHEN ct.constraint_defs IS NOT NULL THEN E',\n' || ct.constraint_defs
    ELSE ''
  END,
  COALESCE(idx.index_defs, '')
)
FROM table_columns tc
LEFT JOIN (
  SELECT
    table_name,
    string_agg(
      format('  CONSTRAINT %I %s', conname, definition),
      E',\n'
    ) AS constraint_defs
  FROM constraints
  GROUP BY table_name
) ct ON tc.table_name = ct.table_name
LEFT JOIN (
  SELECT
    tablename,
    string_agg(indexdef, E'\n') AS index_defs
  FROM indexes
  GROUP BY tablename
) idx ON tc.table_name = idx.tablename
GROUP BY tc.table_name, ct.constraint_defs, idx.index_defs
ORDER BY tc.table_name;
`

// DatabaseContext returns markdown with database schema information
func DatabaseContext(ctx context.Context, dsn string) (string, error) {
	var builder strings.Builder

	builder.WriteString(`# PostgreSQL Database Tool

You have access to tools for interacting with a PostgreSQL database.
The "Run Database SQL" tool lets you run SQL against the PostgreSQL database.
Display all results from these tools and their schemas in markdown format.
If the user refers to creating or modifying tables, assume they mean a PostgreSQL table and not writing a table in a markdown file.

`)

	schemas, err := RunDatabaseCommand(ctx, dsn, getSchemasSQL, "-At")
	if err != nil {
		return "", fmt.Errorf("error getting schemas: %w", err)
	}

	if schemas == "" {
		builder.WriteString("\n# No tables found in database\n")
	} else {
		builder.WriteString("\n# Database Schema\n\n")
		builder.WriteString(schemas)
	}

	return builder.String(), nil
}
