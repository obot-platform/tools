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

	var mainErr error
	switch command {
	case "listAllDrives":
		mainErr = commands.ListAllDrives(context.Background())
	case "getDrive":
		mainErr = commands.GetDrive(context.Background(), os.Getenv("DRIVE_ID"))
	case "listDriveItems":
		mainErr = commands.ListDriveItems(context.Background(), os.Getenv("DRIVE_ID"), os.Getenv("FOLDER_ID"))
	case "listSharedItems":
		mainErr = commands.ListSharedItems(context.Background())
	case "getDriveItem":
		mainErr = commands.GetDriveItem(context.Background(), os.Getenv("DRIVE_ID"), os.Getenv("ITEM_ID"))
	case "updateDriveItem":
		mainErr = commands.UpdateDriveItem(context.Background(), os.Getenv("DRIVE_ID"), os.Getenv("ITEM_ID"), os.Getenv("NEW_FOLDER_ID"), os.Getenv("NEW_NAME"))
	case "deleteDriveItem":
		mainErr = commands.DeleteDriveItem(context.Background(), os.Getenv("DRIVE_ID"), os.Getenv("ITEM_ID"))
	case "downloadDriveItem":
		mainErr = commands.DownloadDriveItem(context.Background(), os.Getenv("DRIVE_ID"), os.Getenv("ITEM_ID"), os.Getenv("WORKSPACE_FILE_NAME"))
	case "uploadDriveItem":
		mainErr = commands.UploadDriveItem(context.Background(), os.Getenv("DRIVE_ID"), os.Getenv("FOLDER_ID"), os.Getenv("WORKSPACE_FILE_NAME"))
	case "copyDriveItem":
		mainErr = commands.CopyDriveItem(context.Background(), os.Getenv("SOURCE_DRIVE_ID"), os.Getenv("SOURCE_ITEM_ID"), os.Getenv("TARGET_DRIVE_ID"), os.Getenv("TARGET_FOLDER_ID"), os.Getenv("NEW_NAME"))
	case "createFolder":
		mainErr = commands.CreateFolder(context.Background(), os.Getenv("DRIVE_ID"), os.Getenv("FOLDER_ID"), os.Getenv("FOLDER_NAME"))
	case "addPermission":
		mainErr = commands.AddPermission(context.Background(), os.Getenv("DRIVE_ID"), os.Getenv("ITEM_ID"), os.Getenv("EMAILS"), os.Getenv("ROLE"), os.Getenv("MESSAGE"), os.Getenv("PASSWORD"), os.Getenv("EXPIRATION_DATE_TIME"))
	case "deletePermission":
		mainErr = commands.DeletePermission(context.Background(), os.Getenv("DRIVE_ID"), os.Getenv("ITEM_ID"), os.Getenv("PERMISSION_ID"))
	case "listPermissions":
		mainErr = commands.ListPermissions(context.Background(), os.Getenv("DRIVE_ID"), os.Getenv("ITEM_ID"))
	default:
		fmt.Printf("Unknown command: %q\n", command)
		os.Exit(1)
	}

	if mainErr != nil {
		fmt.Println(mainErr)
		os.Exit(1)
	}
}
