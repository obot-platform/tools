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
	if owner := drive.GetOwner(); owner != nil {
		user := owner.GetUser()
		if user != nil && user.GetDisplayName() != nil {
			fmt.Printf("Owner: %s\n", *user.GetDisplayName())
		}
	}
	if details {
		if quota := drive.GetQuota(); quota != nil {
			if total := quota.GetTotal(); total != nil {
				fmt.Printf("Total: %d GB\n", *total/1024/1024/1024)
			}
			if used := quota.GetUsed(); used != nil {
				fmt.Printf("Used: %d GB\n", *used/1024/1024/1024)
			}
			if remaining := quota.GetRemaining(); remaining != nil {
				fmt.Printf("Remaining: %d GB\n", *remaining/1024/1024/1024)
			}
			if state := quota.GetState(); state != nil {
				fmt.Printf("State: %s\n", *state)
			}
		}
	}
}
