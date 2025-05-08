package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/client"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/global"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/util"
)

func ListPermissions(ctx context.Context, driveID string, itemID string) error {
	if driveID == "me" {
		return fmt.Errorf("drive_id must be the actual drive ID, cannot be 'me'")
	}

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	permissions, err := graph.ListDriveItemPermissions(ctx, c, driveID, itemID)
	if err != nil {
		return fmt.Errorf("failed to list permissions: %w", err)
	}

	for _, permission := range permissions {
		if id := permission.GetId(); id != nil {
			fmt.Printf("Permission ID: %s\n", util.Deref(id))
		}

		if roles := permission.GetRoles(); roles != nil {
			fmt.Printf("  Roles: %v\n", roles)
		}

		if link := permission.GetLink(); link != nil {
			fmt.Printf("  Link:\n")
			if url := link.GetWebUrl(); url != nil {
				fmt.Printf("    Web URL: %s\n", util.Deref(url))
			}
			if app := link.GetApplication(); app != nil {
				fmt.Printf("    Application:\n")
				if id := app.GetId(); id != nil {
					fmt.Printf("      ID: %s\n", util.Deref(id))
				}
				if name := app.GetDisplayName(); name != nil {
					fmt.Printf("      Display Name: %s\n", util.Deref(name))
				}
			}
		}

		if grantedToV2 := permission.GetGrantedToV2(); grantedToV2 != nil {
			fmt.Printf("  Granted To:\n")
			if user := grantedToV2.GetUser(); user != nil {
				if id := user.GetId(); id != nil {
					fmt.Printf("    User ID: %s\n", util.Deref(id))
				}
				if name := user.GetDisplayName(); name != nil {
					fmt.Printf("    Display Name: %s\n", util.Deref(name))
				}
			}
			if siteUser := grantedToV2.GetSiteUser(); siteUser != nil {
				fmt.Printf("    Site User:\n")
				if id := siteUser.GetId(); id != nil {
					fmt.Printf("      ID: %s\n", util.Deref(id))
				}
				if name := siteUser.GetDisplayName(); name != nil {
					fmt.Printf("      Display Name: %s\n", util.Deref(name))
				}
			}
		}

		if inheritedFrom := permission.GetInheritedFrom(); inheritedFrom != nil {
			fmt.Printf("  Inherited From:\n")
			if driveId := inheritedFrom.GetDriveId(); driveId != nil {
				fmt.Printf("    Drive ID: %s\n", util.Deref(driveId))
			}
			if id := inheritedFrom.GetId(); id != nil {
				fmt.Printf("    Item ID: %s\n", util.Deref(id))
			}
			if path := inheritedFrom.GetPath(); path != nil {
				fmt.Printf("    Path: %s\n", util.Deref(path))
			}
		}

		fmt.Println()
	}
	return nil
}
