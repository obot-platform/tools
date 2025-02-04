package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/gptscript-ai/tools/word/pkg/client"
	"github.com/gptscript-ai/tools/word/pkg/global"
	"github.com/gptscript-ai/tools/word/pkg/graph"
)

func GetDoc(ctx context.Context, docID string) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return err
	}

	var content string
	if strings.HasSuffix(docID, ".docx") || strings.Contains(docID, "/") {
		content, err = graph.GetDocByPath(ctx, c, docID)
	} else {
		content, err = graph.GetDoc(ctx, c, docID)
	}
	if err != nil {
		return fmt.Errorf("failed to get doc %q: %w", docID, err)
	}

	fmt.Println(content)

	return nil
}

func GetDocByPath(ctx context.Context, path string) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return err
	}

	content, err := graph.GetDocByPath(ctx, c, path)
	if err != nil {
		return fmt.Errorf("failed to get doc %q: %w", path, err)
	}

	fmt.Println(content)

	return nil
}
