package main

import (
	"context"
	"fmt"
	"obot-platform/database/pkg/cmd"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: gptscript-go-tool <command>")
		os.Exit(1)
	}
	command := os.Args[1]
	ctx := context.Background()

	workspaceID := os.Getenv("DATABASE_WORKSPACE_ID")
	if workspaceID == "" {
		// TODO(njhale): Figure out why DATABASE_WORKSPACE_ID is not set here for the UI tools.
		workspaceID = os.Getenv("GPTSCRIPT_WORKSPACE_ID")
	}

	// Get admin DSN from environment variable
	adminDSN := os.Getenv("POSTGRES_DSN")

	// Setup database and user with admin credentials
	dsn, err := cmd.EnsureTenantSchema(ctx, adminDSN, workspaceID)
	if err != nil {
		fmt.Printf("Error setting up database: %v\n", err)
		os.Exit(1)
	}

	// Run the requested command using the user credentials
	var result string
	switch command {
	case "listDatabaseTables":
		result, err = cmd.ListDatabaseTables(ctx, dsn)

	case "listDatabaseTableRows":
		table := os.Getenv("TABLE")
		if table == "" {
			err = fmt.Errorf("TABLE environment variable is required")
			break
		}
		result, err = cmd.ListDatabaseTableRows(ctx, dsn, table)

	case "runDatabaseSQL":
		sql := os.Getenv("SQL")
		if sql == "" {
			err = fmt.Errorf("SQL environment variable is required")
			break
		}
		result, err = cmd.RunDatabaseCommand(ctx, dsn, sql)

	case "databaseContext":
		result, err = cmd.DatabaseContext(ctx, dsn)

	default:
		err = fmt.Errorf("unknown command: %s", command)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Print(result)
}
