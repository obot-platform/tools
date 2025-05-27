package main

import (
	"context"
	"fmt"
	"os"

	"github.com/obot-platform/tools/microsoft365/word/pkg/commands"
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
	case "writeDoc":
		err = commands.WriteDoc(ctx, os.Getenv("DOC_NAME"), os.Getenv("DOC_CONTENT"), os.Getenv("OVERWRITE_IF_EXISTS") == "true")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
