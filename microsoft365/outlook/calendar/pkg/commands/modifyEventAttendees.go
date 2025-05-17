package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/outlook/common/id"
)

func ModifyEventAttendees(ctx context.Context, eventID, calendarID string, owner graph.OwnerType, addRequiredAttendeesList, addOptionalAttendeesList, removeAttendeesList []string) error {
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

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := graph.ModifyEventAttendee(ctx, c, trueEventID, trueCalendarID, owner, addRequiredAttendeesList, addOptionalAttendeesList, removeAttendeesList); err != nil {
		return fmt.Errorf("failed to invite user to event: %w", err)
	}

	fmt.Println("Successfully invited user to event")
	return nil
}
