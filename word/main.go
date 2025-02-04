package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/gptscript-ai/tools/word/pkg/commands"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: gptscript-go-tool <command>")
		os.Exit(1)
	}

	command := os.Args[1]

	var (
		err error
		ctx = context.Background()
	)
	switch command {
	case "listDocs":
		err = commands.ListDocs(ctx)
	case "getDoc":
		err = commands.GetDoc(ctx, os.Getenv("DOC_ID"))
	case "getDocByPath":
		err = commands.GetDocByPath(ctx, os.Getenv("DOC_PATH"))
	case "createDoc":
		slog.Info("Creating doc", "token", os.Getenv("GPTSCRIPT_MICROSOFT_WORD_TOKEN"))
		err = commands.CreateDoc(ctx, os.Getenv("DOC_DRIVE_DIR"), os.Getenv("DOC_TITLE"), os.Getenv("DOC_CONTENT"))
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
