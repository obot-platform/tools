package printer

import (
	"fmt"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func PrintDrive(drive models.Driveable, details bool) {
	fmt.Printf("Drive ID: %s\n", *drive.GetId())
	fmt.Printf("Name: %s\n", *drive.GetName())
	if driveType := drive.GetDriveType(); driveType != nil {
		fmt.Printf("Drive Type: %s\n", *driveType)
	}
	if webUrl := drive.GetWebUrl(); webUrl != nil {
		fmt.Printf("Web URL: %s\n", *webUrl)
	}
	if owner := drive.GetOwner(); owner != nil {
		user := owner.GetUser()
		if user != nil && user.GetDisplayName() != nil {
			fmt.Printf("Owner: %s\n", *user.GetDisplayName())
		}
	}
	if details {
		if quota := drive.GetQuota(); quota != nil {
			if total := quota.GetTotal(); total != nil {
				fmt.Printf("Total: %.2f GB\n", float64(*total)/1024/1024/1024)
			}
			if used := quota.GetUsed(); used != nil {
				fmt.Printf("Used: %.2f GB\n", float64(*used)/1024/1024/1024)
			}
			if remaining := quota.GetRemaining(); remaining != nil {
				fmt.Printf("Remaining: %.2f GB\n", float64(*remaining)/1024/1024/1024)
			}
			if state := quota.GetState(); state != nil {
				fmt.Printf("State: %s\n", *state)
			}
		}
	}
}
