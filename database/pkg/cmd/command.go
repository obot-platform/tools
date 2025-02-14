package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// RunDatabaseCommand runs a sqlite3 command against the database and returns the output from the sqlite3 CLI.
func RunDatabaseCommand(ctx context.Context, dbFile *os.File, sql string, opts ...string) (string, error) {
	// Remove the "sqlite3" prefix and trim whitespace
	args := append(opts, dbFile.Name())
	if arg := strings.TrimSpace(sql); arg != "" {
		// Use strconv.Unquote to safely handle quotes and escape sequences
		unquoted, err := strconv.Unquote(arg)
		if err != nil {
			// If unquoting fails (e.g. string wasn't quoted), use original
			unquoted = arg
		}
		args = append(args, unquoted)
	}

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
