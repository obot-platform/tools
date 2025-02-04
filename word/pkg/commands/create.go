package commands

import (
	"context"
	"log/slog"

	"github.com/gptscript-ai/tools/word/pkg/client"
	"github.com/gptscript-ai/tools/word/pkg/global"
	"github.com/gptscript-ai/tools/word/pkg/graph"
)

func CreateDoc(ctx context.Context, dir, name, content string) error {
	c, err := client.NewClient(global.ReadWriteScopes)
	if err != nil {
		return err
	}

	slog.Info("Creating new Word Document in OneDrive", "dir", dir, "name", name)

	_, _, err = graph.CreateDoc(ctx, c, dir, name, content)
	if err != nil {
		return err
	}

	return nil
}
