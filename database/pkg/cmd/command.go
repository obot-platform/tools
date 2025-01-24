package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/google/shlex"
)

// RunDatabaseCommand runs a sqlite3 command against the database and returns the output from the sqlite3 CLI.
func RunDatabaseCommand(ctx context.Context, dbFile *os.File, sqlite3Args string) (string, error) {
	// Remove the "sqlite3" prefix and trim whitespace
	sqlite3Args = strings.TrimPrefix(strings.TrimSpace(sqlite3Args), "sqlite3")

	// Split the arguments using shlex
	args, err := shlex.Split(sqlite3Args)
	if err != nil {
		return "", fmt.Errorf("error parsing sqlite3 args: %w", err)
	}

	// Append the database file name as the first argument
	args = append([]string{dbFile.Name()}, args...)

	// Build the sqlite3 command
	cmd := exec.CommandContext(ctx, "sqlite3", args...)

	// Redirect command output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command and capture errors
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error executing sqlite3: %w, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}
