package cmd

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// RunDatabaseCommand executes a command against the Postgres database
func RunDatabaseCommand(ctx context.Context, dsn string, sql string, opts ...string) (string, error) {
	if sql == "" {
		return "", fmt.Errorf("SQL cannot be empty")
	}

	args := append([]string{dsn}, opts...)

	unquoted, err := strconv.Unquote(sql)
	if err != nil {
		unquoted = sql
	}
	args = append(args, "-c", unquoted)

	cmd := exec.CommandContext(ctx, "psql", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("psql error: %w\nstderr: %s", err, stderr.String())
	}

	if stderr.Len() > 0 {
		return stdout.String(), fmt.Errorf("psql stderr: %s", stderr.String())
	}

	return stdout.String(), nil
}

// EnsureTenantSchema creates a schema and role for a tenant with proper isolation
func EnsureTenantSchema(ctx context.Context, adminDSN, workspaceID string) (string, error) {
	schemaName := workspaceSchemaName(workspaceID)
	userName := schemaName
	password := generatePassword(workspaceID)
	dbName := "obot_db"

	// Create shared database if it doesn't exist
	checkDBSQL := fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname = '%s'", dbName)
	dbExistsCheck, err := RunDatabaseCommand(ctx, adminDSN, checkDBSQL, "-At")
	if err != nil {
		return "", fmt.Errorf("error checking for shared database: %w", err)
	}
	if strings.TrimSpace(dbExistsCheck) != "1" {
		createDBSQL := fmt.Sprintf("CREATE DATABASE %s", dbName)
		if _, err := RunDatabaseCommand(ctx, adminDSN, createDBSQL); err != nil {
			return "", fmt.Errorf("error creating shared database: %w", err)
		}

		// Create tenant role
		createRoleSQL := fmt.Sprintf(`CREATE ROLE %s WITH LOGIN PASSWORD '%s'`, userName, password)
		if _, err := RunDatabaseCommand(ctx, adminDSN, createRoleSQL); err != nil && !strings.Contains(err.Error(), "already exists") {
			return "", fmt.Errorf("error creating role: %w", err)
		}

		// Connect to shared database
		dbDSN, err := dsnWithDatabase(adminDSN, dbName)
		if err != nil {
			return "", fmt.Errorf("error constructing DSN for shared database: %w", err)
		}

		// Create schema for tenant (owned by admin user)
		createSchemaSQL := fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, schemaName)
		if _, err := RunDatabaseCommand(ctx, dbDSN, createSchemaSQL); err != nil {
			return "", fmt.Errorf("error creating schema: %w", err)
		}

		// Revoke PUBLIC access on public schema
		revokePublicSchemaSQL := `REVOKE ALL ON SCHEMA public FROM PUBLIC`
		if _, err := RunDatabaseCommand(ctx, dbDSN, revokePublicSchemaSQL); err != nil {
			return "", fmt.Errorf("error revoking public schema privileges: %w", err)
		}

		// Set up tenant schema permissions
		statements := []string{
			fmt.Sprintf(`REVOKE ALL ON SCHEMA %s FROM PUBLIC`, schemaName),
			fmt.Sprintf(`GRANT USAGE, CREATE ON SCHEMA %s TO %s`, schemaName, userName),
			fmt.Sprintf(`GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA %s TO %s`, schemaName, userName),
			fmt.Sprintf(`GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA %s TO %s`, schemaName, userName),
			fmt.Sprintf(`GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA %s TO %s`, schemaName, userName),
			fmt.Sprintf(`ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT ALL ON TABLES TO %s`, schemaName, userName),
			fmt.Sprintf(`ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT ALL ON SEQUENCES TO %s`, schemaName, userName),
			fmt.Sprintf(`ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT ALL ON FUNCTIONS TO %s`, schemaName, userName),
			fmt.Sprintf(`ALTER ROLE %s SET search_path = %s`, userName, schemaName),
		}

		for _, stmt := range statements {
			if _, err := RunDatabaseCommand(ctx, dbDSN, stmt); err != nil {
				return "", fmt.Errorf("error executing statement '%s': %w", stmt, err)
			}
		}
	}

	userDSN := fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=require", userName, password, extractHost(adminDSN), dbName)
	return userDSN, nil
}

// generatePassword creates a hashed password using the workspaceID
func generatePassword(workspaceID string) string {
	hash := sha256.Sum256([]byte(workspaceID))
	return hex.EncodeToString(hash[:])
}

// extractHost extracts the host part from the DSN
func extractHost(dsn string) string {
	re := regexp.MustCompile(`^(postgresql://[^:]+:[^@]+@)([^/]+)(/[^?]*)(\?.+)?$`)
	matches := re.FindStringSubmatch(dsn)
	if len(matches) >= 3 {
		return matches[2]
	}
	return ""
}

// workspaceSchemaName converts a workspace ID into a valid PostgreSQL schema/role identifier
func workspaceSchemaName(workspaceID string) string {
	hash := sha256.Sum256([]byte(workspaceID))
	return "schema_" + hex.EncodeToString(hash[:16])
}

// dsnWithDatabase switches the database in the DSN string
func dsnWithDatabase(adminDSN, dbName string) (string, error) {
	if strings.HasPrefix(adminDSN, "postgresql://") {
		re := regexp.MustCompile(`^(postgresql://[^/]+/)([^?]*)(\?.+)?$`)
		matches := re.FindStringSubmatch(adminDSN)
		if len(matches) >= 3 {
			if matches[3] != "" {
				return matches[1] + dbName + matches[3], nil
			}
			return matches[1] + dbName, nil
		}
	}
	return "", fmt.Errorf("invalid DSN format")
}
