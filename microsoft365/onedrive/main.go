package main

import (
	"context"
	"fmt"
	"os"

	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/commands"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: Onedrive <command>")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "listAllDrives":
		if err := commands.ListAllDrives(context.Background()); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "getDrive":
		if err := commands.GetDrive(context.Background(), os.Getenv("DRIVE_ID")); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "listDriveItems":
		if err := commands.ListDriveItems(context.Background(), os.Getenv("DRIVE_ID"), os.Getenv("FOLDER_ID")); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "listSharedItems":
		if err := commands.ListSharedItems(context.Background()); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "getDriveItem":
		if err := commands.GetDriveItem(context.Background(), os.Getenv("DRIVE_ID"), os.Getenv("ITEM_ID")); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "updateDriveItem":
		if err := commands.UpdateDriveItem(context.Background(), os.Getenv("DRIVE_ID"), os.Getenv("ITEM_ID"), os.Getenv("NEW_FOLDER_ID"), os.Getenv("NEW_NAME")); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "deleteDriveItem":
		if err := commands.DeleteDriveItem(context.Background(), os.Getenv("DRIVE_ID"), os.Getenv("ITEM_ID")); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "downloadDriveItem":
		if err := commands.DownloadDriveItem(context.Background(), os.Getenv("DRIVE_ID"), os.Getenv("ITEM_ID"), os.Getenv("WORKSPACE_FILE_NAME")); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "uploadDriveItem":
		if err := commands.UploadDriveItem(context.Background(), os.Getenv("DRIVE_ID"), os.Getenv("FOLDER_ID"), os.Getenv("WORKSPACE_FILE_NAME")); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "copyDriveItem":
		if err := commands.CopyDriveItem(context.Background(), os.Getenv("SOURCE_DRIVE_ID"), os.Getenv("SOURCE_ITEM_ID"), 
			os.Getenv("TARGET_DRIVE_ID"), os.Getenv("TARGET_FOLDER_ID"), os.Getenv("NEW_NAME")); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "createFolder":
		if err := commands.CreateFolder(context.Background(), os.Getenv("DRIVE_ID"), os.Getenv("FOLDER_ID"), os.Getenv("FOLDER_NAME")); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %q\n", command)
		os.Exit(1)
	}
}