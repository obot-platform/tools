package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/printers"
	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/util"
	"github.com/obot-platform/tools/microsoft365/outlook/common/id"
)

func GetEventDetails(ctx context.Context, eventID, calendarID string, owner graph.OwnerType) error {
	trueEventID, err := id.GetOutlookID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get Outlook ID: %w", err)
	}

	var trueCalendarID string
	if calendarID != "" {
		trueCalendarID, err = id.GetOutlookID(ctx, calendarID)
		if err != nil {
			return fmt.Errorf("failed to get Outlook ID: %w", err)
		}
	}

	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	event, err := graph.GetEvent(ctx, c, trueEventID, trueCalendarID, owner)
	if err != nil {
		return fmt.Errorf("failed to get event: %w", err)
	}

	event.SetId(util.Ptr(eventID))

	printers.PrintEvent(event, true)
	return nil
}
