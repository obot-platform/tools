package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gptscript-ai/tools/outlook/mail/pkg/commands"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/graph"
	"github.com/obot-platform/obot/apiclient"
)

var (
	url         = os.Getenv("OBOT_SERVER_URL")
	token       = os.Getenv("OBOT_TOKEN")
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
	case "listMessages":
		if err := commands.ListMessages(
			context.Background(),
			os.Getenv("FOLDER_ID"),
			os.Getenv("START"),
			os.Getenv("END"),
			os.Getenv("LIMIT"),
		); err != nil {
			fmt.Printf("failed to list mail: %v\n", err)
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
	case "getMessageDetails":
		if err := commands.GetMessageDetails(context.Background(), os.Getenv("MESSAGE_ID")); err != nil {
			fmt.Printf("failed to get message details: %v\n", err)
			os.Exit(1)
		}
	case "searchMessages":
		if err := commands.SearchMessages(
			context.Background(),
			os.Getenv("SUBJECT"),
			os.Getenv("FROM_ADDRESS"),
			os.Getenv("FROM_NAME"),
			os.Getenv("FOLDER_ID"),
			os.Getenv("START"),
			os.Getenv("END"),
			os.Getenv("LIMIT"),
		); err != nil {
			fmt.Printf("failed to search messages: %v\n", err)
			os.Exit(1)
		}
	case "createDraft":
		if err := commands.CreateDraft(context.Background(), getDraftInfoFromEnv()); err != nil {
			fmt.Printf("failed to create draft: %v\n", err)
			os.Exit(1)
		}
	case "createGroupThreadMessage":
		if err := commands.CreateGroupThreadMessage(context.Background(), os.Getenv("GROUP_ID"), os.Getenv("REPLY_TO_THREAD_ID"), getDraftInfoFromEnv()); err != nil {
			fmt.Printf("failed to create group thread message: %v\n", err)
			os.Exit(1)
		}
	case "sendDraft":
		if err := commands.SendDraft(context.Background(), os.Getenv("DRAFT_ID")); err != nil {
			fmt.Printf("failed to send draft: %v\n", err)
			os.Exit(1)
		}
	case "deleteMessage":
		if err := commands.DeleteMessage(context.Background(), os.Getenv("MESSAGE_ID")); err != nil {
			fmt.Printf("failed to delete message: %v\n", err)
			os.Exit(1)
		}
	case "deleteGroupThread":
		if err := commands.DeleteGroupThread(context.Background(), os.Getenv("GROUP_ID"), os.Getenv("THREAD_ID")); err != nil {
			fmt.Printf("failed to delete group thread: %v\n", err)
			os.Exit(1)
		}
	case "moveMessage":
		if err := commands.MoveMessage(context.Background(), os.Getenv("MESSAGE_ID"), os.Getenv("DESTINATION_FOLDER_ID")); err != nil {
			fmt.Printf("failed to move message: %v\n", err)
			os.Exit(1)
		}
	case "getMyEmailAddress":
		if err := commands.GetMyEmailAddress(context.Background()); err != nil {
			fmt.Printf("failed to get my email address: %v\n", err)
			os.Exit(1)
		}
	case "listAttachments":
		if err := commands.ListAttachments(context.Background(), os.Getenv("MESSAGE_ID")); err != nil {
			fmt.Printf("failed to list attachments: %v\n", err)
			os.Exit(1)
		}
	case "downloadAttachment":
		client := apiclient.NewClientFromEnv()

		if err := commands.DownloadAttachment(context.Background(), os.Getenv("MESSAGE_ID"), os.Getenv("ATTACHMENT_ID"), client, &commands.DownloadAttachmentOpts{
			ThreadID:    threadID,
			ProjectID:   projectID,
			AssistantID: assistantID,
		}); err != nil {
			fmt.Printf("failed to download attachment: %v\n", err)
			os.Exit(1)
		}
	case "getAttachment":
		if err := commands.GetAttachment(context.Background(), os.Getenv("MESSAGE_ID"), os.Getenv("ATTACHMENT_ID")); err != nil {
			fmt.Printf("failed to get attachment: %v\n", err)
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
		Subject:          os.Getenv("SUBJECT"),
		Body:             os.Getenv("BODY"),
		Recipients:       smartSplit(os.Getenv("RECIPIENTS"), ","),
		CC:               smartSplit(os.Getenv("CC"), ","),
		BCC:              smartSplit(os.Getenv("BCC"), ","),
		Attachments:      attachments,
		ReplyAll:         os.Getenv("REPLY_ALL") == "true",
		ReplyToMessageID: os.Getenv("REPLY_MESSAGE_ID"),
	}

	// We need to unset BODY, because if it's still set when we try to write files to the workspace,
	// it will cause problems, since the workspace tools have an argument with the same name.
	_ = os.Unsetenv("BODY")

	return info
}
