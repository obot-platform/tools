package main

import (
	"context"
	"fmt"
	"os"

	"github.com/obot-platform/tools/microsoft365/outlook/contact/pkg/commands"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: Contact <command>")
		os.Exit(1)
	}

	command := os.Args[1]

	var mainErr error
	switch command {
	case "listContacts":
		mainErr = commands.ListContacts(context.Background())
	case "createContact":
		mainErr = commands.CreateContact(context.Background(), os.Getenv("GIVEN_NAME"), os.Getenv("SURNAME"), os.Getenv("EMAILS"), os.Getenv("BUSINESS_PHONES"))
	case "updateContact":
		mainErr = commands.UpdateContact(context.Background(), os.Getenv("CONTACT_ID"), os.Getenv("GIVEN_NAME"), os.Getenv("SURNAME"), os.Getenv("EMAILS"), os.Getenv("BUSINESS_PHONES"))
	case "deleteContact":
		mainErr = commands.DeleteContact(context.Background(), os.Getenv("CONTACT_ID"))
	case "getContact":
		mainErr = commands.GetContact(context.Background(), os.Getenv("CONTACT_ID"))
	default:
		fmt.Printf("Unknown command: %q\n", command)
		os.Exit(1)
	}

	if mainErr != nil {
		fmt.Println(mainErr)
		os.Exit(1)
	}
}
