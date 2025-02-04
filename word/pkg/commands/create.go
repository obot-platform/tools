package commands

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/gptscript-ai/tools/word/pkg/client"
	"github.com/gptscript-ai/tools/word/pkg/convert"
	"github.com/gptscript-ai/tools/word/pkg/global"
	"github.com/gptscript-ai/tools/word/pkg/graph"
)

func CreateDoc(ctx context.Context, dir string, name string, content string) error {
	c, err := client.NewClient(global.ReadWriteScopes)
	if err != nil {
		return err
	}

	slog.Info("Creating new Word Document in OneDrive", "dir", dir, "name", name)

	contentBytes, err := convert.MarkdownToDocx(content)
	if err != nil {
		return fmt.Errorf("failed to convert markdown to docx: %w", err)
	}

	name = strings.TrimSuffix(name, filepath.Ext(name)) + ".docx"
	_, _, err = graph.CreateDoc(ctx, c, dir, name, contentBytes)
	if err != nil {
		return err
	}

	return nil
}
