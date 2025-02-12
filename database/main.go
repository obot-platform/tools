package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"slices"

	"obot-platform/database/pkg/cmd"

	"github.com/gptscript-ai/go-gptscript"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

var workspaceID = os.Getenv("DATABASE_WORKSPACE_ID")

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: gptscript-go-tool <command>")
		os.Exit(1)
	}
	command := os.Args[1]

	g, err := gptscript.NewGPTScript()
	if err != nil {
		fmt.Printf("Error creating GPTScript: %v\n", err)
		os.Exit(1)
	}
	defer g.Close()

	var (
		ctx                    = context.Background()
		dbFileName             = "acorn.db"
		dbWorkspacePath        = "/databases/" + dbFileName
		revisionID      string = "-1"
		initialDBData   []byte
	)

	workspaceDB, err := g.ReadFileWithRevisionInWorkspace(ctx, dbWorkspacePath, gptscript.ReadFileInWorkspaceOptions{
		WorkspaceID: workspaceID,
	})

	var notFoundErr *gptscript.NotFoundInWorkspaceError
	if err != nil && !errors.As(err, &notFoundErr) {
		fmt.Printf("Error reading DB file: %v\n", err)
		os.Exit(1)
	}

	// Create a temporary file for the SQLite database
	dbFile, err := os.CreateTemp("", dbFileName)
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		os.Exit(1)
	}
	defer dbFile.Close()
	defer os.Remove(dbFile.Name())

	// Write the data to the temporary file
	if workspaceDB != nil && workspaceDB.Content != nil {
		initialDBData = workspaceDB.Content
		if err := os.WriteFile(dbFile.Name(), initialDBData, 0644); err != nil {
			fmt.Printf("Error writing to temp file: %v\n", err)
			os.Exit(1)
		}
		if workspaceDB.RevisionID != "" {
			revisionID = workspaceDB.RevisionID
		}
	}

	// Run the requested command
	var result string
	switch command {
	case "listDatabaseTables":
		result, err = cmd.ListDatabaseTables(ctx, dbFile)
	case "listDatabaseTableRows":
		result, err = cmd.ListDatabaseTableRows(ctx, dbFile, os.Getenv("TABLE"))
	case "runDatabaseSQL":
		result, err = cmd.RunDatabaseCommand(ctx, dbFile, os.Getenv("SQL"), "-header")
		if err == nil {
			err = saveWorkspaceDB(ctx, g, dbWorkspacePath, revisionID, dbFile, initialDBData)
		}
	case "databaseContext":
		result, err = cmd.DatabaseContext(ctx, dbFile)
	default:
		err = fmt.Errorf("unknown command: %s", command)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Print(result)
}

// saveWorkspaceDB saves the updated database file to the workspace if the content of the database has changed.
func saveWorkspaceDB(
	ctx context.Context,
	g *gptscript.GPTScript,
	dbWorkspacePath string,
	revisionID string,
	dbFile *os.File,
	initialDBData []byte,
) error {
	updatedDBData, err := os.ReadFile(dbFile.Name())
	if err != nil {
		return fmt.Errorf("Error reading updated DB file: %v", err)
	}

	if hash(initialDBData) == hash(updatedDBData) {
		return nil
	}

	if err := g.WriteFileInWorkspace(ctx, dbWorkspacePath, updatedDBData, gptscript.WriteFileInWorkspaceOptions{
		WorkspaceID:      workspaceID,
		CreateRevision:   &([]bool{true}[0]),
		LatestRevisionID: revisionID,
	}); err != nil {
		return fmt.Errorf("Error writing updated DB file to workspace: %v", err)
	}

	// Delete old revisions after successfully writing the new revision
	revisions, err := g.ListRevisionsForFileInWorkspace(ctx, dbWorkspacePath, gptscript.ListRevisionsForFileInWorkspaceOptions{
		WorkspaceID: workspaceID,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing revisions: %v\n", err)
		return nil
	}

	lastRevisionIndex := slices.IndexFunc(revisions, func(rev gptscript.FileInfo) bool {
		return rev.RevisionID == revisionID
	})

	if lastRevisionIndex < 0 {
		return nil
	}

	for _, rev := range revisions[:lastRevisionIndex+1] {
		if err := g.DeleteRevisionForFileInWorkspace(ctx, dbWorkspacePath, rev.RevisionID, gptscript.DeleteRevisionForFileInWorkspaceOptions{
			WorkspaceID: workspaceID,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting revision %s: %v\n", rev.RevisionID, err)
		}
	}

	return nil
}

// hash computes the SHA-256 hash of the given data and returns it as a hexadecimal string
func hash(data []byte) string {
	if data == nil {
		return ""
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
