package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/outlook/common/id"
)

func DeleteEvent(ctx context.Context, eventID, calendarID string, owner graph.OwnerType, deleteSeries bool) error {
	trueEventID, err := id.GetOutlookID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get outlook ID: %w", err)
	}

	var trueCalendarID string
	if calendarID != "" {
		trueCalendarID, err = id.GetOutlookID(ctx, calendarID)
		if err != nil {
			return fmt.Errorf("failed to get outlook ID: %w", err)
		}
	}

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if deleteSeries {
		if err := graph.DeleteEventSeries(ctx, c, trueEventID, trueCalendarID, owner); err != nil {
			return fmt.Errorf("failed to delete event series: %w", err)
		}
		fmt.Println("Event series deleted successfully")
	} else {
		if err := graph.DeleteEvent(ctx, c, trueEventID, trueCalendarID, owner); err != nil {
			return fmt.Errorf("failed to delete event: %w", err)
		} else {
			fmt.Println("Event deleted successfully")
		}
	}
	return nil
}
