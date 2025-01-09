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
		// Retrieve environment variables for sending email
		to := os.Getenv("TO")
		subject := os.Getenv("SUBJECT")
		textBody := os.Getenv("TEXT_BODY")
		htmlBody := os.Getenv("HTML_BODY")

		// Validate required fields
		if to == "" || subject == "" || (textBody == "" && htmlBody == "") {
			fmt.Println("Missing required environment variables: TO, SUBJECT, and at least one of TEXT_BODY or HTML_BODY")
			os.Exit(1)
		}

		// Call the Send function
		result, err = cmd.Send(ctx, to, subject, textBody, htmlBody)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Print(result)
}
