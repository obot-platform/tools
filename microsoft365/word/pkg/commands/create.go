package commands

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/obot-platform/tools/microsoft365/word/pkg/client"
	"github.com/obot-platform/tools/microsoft365/word/pkg/convert"
	"github.com/obot-platform/tools/microsoft365/word/pkg/global"
	"github.com/obot-platform/tools/microsoft365/word/pkg/graph"
)

func WriteDoc(ctx context.Context, name string, content string, overwriteIfExists bool) error {
	c, err := client.NewClient(global.ReadWriteScopes)
	if err != nil {
		return err
	}

	// Ensure name has .docx extension
	name = strings.TrimSuffix(name, filepath.Ext(name)) + ".docx"

	// Check if file already exists
	if !overwriteIfExists {
		exists, err := graph.DocExists(ctx, c, name)
		if err != nil {
			return fmt.Errorf("failed to check if document exists: %w", err)
		}

		if exists {
			return fmt.Errorf("document with name %q already exists, aborting to prevent overwrite", name)
		}
	}

	slog.Info("Creating new Word Document in OneDrive", "name", name)

	contentBytes, err := convert.MarkdownToDocx(content)
	if err != nil {
		return fmt.Errorf("failed to convert markdown to docx: %w", err)
	}

	name, id, err := graph.CreateDoc(ctx, c, name, contentBytes)
	if err != nil {
		return err
	}

	fmt.Printf("Wrote content to document with name=%q and id=%q\n", name, id)

	return nil
}
