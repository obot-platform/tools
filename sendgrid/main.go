package main

import (
	"context"
	"fmt"
	"os"

	"github.com/obot-platform/tools/sendgrid/cmd"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: gptscript-go-tool <command>")
		os.Exit(1)
	}
	command := os.Args[1]

	// Run the requested command
	var (
		result string
		err    error
		ctx    = context.Background()
	)
	switch command {
	case "sendEmail":
		result, err = cmd.Send(ctx,
			os.Getenv("SENDGRID_API_KEY"),
			os.Getenv("SENDGRID_EMAIL_ADDRESS"),
			os.Getenv("FROM_NAME"),
			os.Getenv("TO"),
			os.Getenv("SUBJECT"),
			os.Getenv("TEXT_BODY"),
			os.Getenv("HTML_BODY"),
		)

	default:
		err = fmt.Errorf("unknown command: %s", command)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Print(result)
}
