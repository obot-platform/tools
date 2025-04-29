package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/commands"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/graph"
	"github.com/obot-platform/obot/apiclient"
)

var (
	threadID    = os.Getenv("OBOT_THREAD_ID")
	projectID   = "p1" + strings.TrimPrefix(os.Getenv("OBOT_PROJECT_ID"), "t1")
	assistantID = os.Getenv("OBOT_AGENT_ID")
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: mail <command>")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "listMailFolders":
		if err := commands.ListMailFolders(context.Background()); err != nil {
			fmt.Printf("failed to list mail folders: %v\n", err)
			os.Exit(1)
		}
	case "listEmails":
		if err := commands.ListEmails(
			context.Background(),
			os.Getenv("FOLDER_ID"),
			os.Getenv("START"),
			os.Getenv("END"),
			os.Getenv("LIMIT"),
		); err != nil {
			fmt.Printf("failed to list emails: %v\n", err)
			os.Exit(1)
		}
	case "listGroupThreads":
		if err := commands.ListGroupThreads(
			context.Background(),
			os.Getenv("GROUP_ID"),
			os.Getenv("START"),
			os.Getenv("END"),
			os.Getenv("LIMIT"),
		); err != nil {
			fmt.Printf("failed to list group threads: %v\n", err)
			os.Exit(1)
		}
	case "listGroups":
		if err := commands.ListGroups(context.Background()); err != nil {
			fmt.Printf("failed to list groups: %v\n", err)
			os.Exit(1)
		}
	case "getEmailDetails":
		if err := commands.GetEmailDetails(context.Background(), os.Getenv("EMAIL_ID"), os.Getenv("GROUP_ID"), os.Getenv("THREAD_ID")); err != nil {
			fmt.Printf("failed to get email details: %v\n", err)
			os.Exit(1)
		}
	case "searchEmails":
		if err := commands.SearchEmails(
			context.Background(),
			os.Getenv("SUBJECT"),
			os.Getenv("FROM_ADDRESS"),
			os.Getenv("FROM_NAME"),
			os.Getenv("FOLDER_ID"),
			os.Getenv("START"),
			os.Getenv("END"),
			os.Getenv("LIMIT"),
		); err != nil {
			fmt.Printf("failed to search emails: %v\n", err)
			os.Exit(1)
		}
	case "createDraft":
		if err := commands.CreateDraft(context.Background(), getDraftInfoFromEnv()); err != nil {
			fmt.Printf("failed to create draft: %v\n", err)
			os.Exit(1)
		}
	case "createGroupThreadEmail":
		if err := commands.CreateGroupThreadEmail(context.Background(), os.Getenv("GROUP_ID"), os.Getenv("REPLY_TO_THREAD_ID"), getDraftInfoFromEnv()); err != nil {
			fmt.Printf("failed to create group thread email: %v\n", err)
			os.Exit(1)
		}
	case "sendDraft":
		if err := commands.SendDraft(context.Background(), os.Getenv("DRAFT_ID")); err != nil {
			fmt.Printf("failed to send draft: %v\n", err)
			os.Exit(1)
		}
	case "deleteEmail":
		if err := commands.DeleteEmail(context.Background(), os.Getenv("EMAIL_ID")); err != nil {
			fmt.Printf("failed to delete email: %v\n", err)
			os.Exit(1)
		}
	case "deleteGroupThread":
		if err := commands.DeleteGroupThread(context.Background(), os.Getenv("GROUP_ID"), os.Getenv("THREAD_ID")); err != nil {
			fmt.Printf("failed to delete group thread: %v\n", err)
			os.Exit(1)
		}
	case "moveEmail":
		if err := commands.MoveEmail(context.Background(), os.Getenv("EMAIL_ID"), os.Getenv("DESTINATION_FOLDER_ID")); err != nil {
			fmt.Printf("failed to move email: %v\n", err)
			os.Exit(1)
		}
	case "getMyEmailAddress":
		if err := commands.GetMyEmailAddress(context.Background()); err != nil {
			fmt.Printf("failed to get my email address: %v\n", err)
			os.Exit(1)
		}
	case "listAttachments":
		if err := commands.ListAttachments(context.Background(), os.Getenv("EMAIL_ID")); err != nil {
			fmt.Printf("failed to list attachments: %v\n", err)
			os.Exit(1)
		}
	case "downloadAttachment":
		client := apiclient.NewClientFromEnv()

		if err := commands.DownloadAttachment(context.Background(), os.Getenv("ATTACHMENT_ID"), client, &commands.DownloadAttachmentOpts{
			EmailID:       os.Getenv("EMAIL_ID"),
			ThreadID:      threadID,
			ProjectID:     projectID,
			AssistantID:   assistantID,
			GroupID:       os.Getenv("GROUP_ID"),
			GroupThreadID: os.Getenv("THREAD_ID"),
		}); err != nil {
			fmt.Printf("failed to download attachment: %v\n", err)
			os.Exit(1)
		}
	case "readAttachment":
		if err := commands.ReadAttachment(context.Background(), os.Getenv("EMAIL_ID"), os.Getenv("ATTACHMENT_ID")); err != nil {
			fmt.Printf("failed to read attachment: %v\n", err)
			os.Exit(1)
		}
	case "listGroupThreadEmailAttachments":
		if err := commands.ListGroupThreadEmailAttachments(
			context.Background(),
			os.Getenv("GROUP_ID"),
			os.Getenv("THREAD_ID"),
			os.Getenv("EMAIL_ID"),
		); err != nil {
			fmt.Printf("failed to list group thread email attachments: %v\n", err)
			os.Exit(1)
		}
	case "getGroupThreadEmailAttachment":
		if err := commands.GetGroupThreadEmailAttachment(
			context.Background(),
			os.Getenv("GROUP_ID"),
			os.Getenv("THREAD_ID"),
			os.Getenv("EMAIL_ID"),
			os.Getenv("ATTACHMENT_ID"),
		); err != nil {
			fmt.Printf("failed to get group thread email attachment: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func smartSplit(s, sep string) []string {
	if s == "" {
		return []string{} // Return an empty slice if the input is empty
	}
	return strings.Split(s, sep)
}

func getDraftInfoFromEnv() graph.DraftInfo {
	var attachments []string
	if os.Getenv("ATTACHMENTS") != "" {
		attachments = smartSplit(os.Getenv("ATTACHMENTS"), ",")
	}

	info := graph.DraftInfo{
		Subject:        os.Getenv("SUBJECT"),
		Body:           os.Getenv("BODY"),
		Recipients:     smartSplit(os.Getenv("RECIPIENTS"), ","),
		CC:             smartSplit(os.Getenv("CC"), ","),
		BCC:            smartSplit(os.Getenv("BCC"), ","),
		Attachments:    attachments,
		ReplyAll:       os.Getenv("REPLY_ALL") == "true",
		ReplyToEmailID: os.Getenv("REPLY_EMAIL_ID"),
	}

	// We need to unset BODY, because if it's still set when we try to write files to the workspace,
	// it will cause problems, since the workspace tools have an argument with the same name.
	_ = os.Unsetenv("BODY")

	return info
}
