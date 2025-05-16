package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/printers"
	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/util"
	"github.com/obot-platform/tools/microsoft365/outlook/common/id"
)

func ListEvents(ctx context.Context, calendarIDstring string, start, end time.Time, limit string) error {
	var limitInt int32
	if limit == "" {
		limitInt = 50 // default limit
	} else {
		var err error
		limitInt64, err := strconv.ParseInt(limit, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid limit value provided: (%s) %w. must be a positive integer", limit, err)
		}
		limitInt = int32(limitInt64)
	}

	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	calendars, err := graph.ListCalendars(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to list calendars: %w", err)
	}

	calendarIDs := util.Map(calendars, func(cal graph.CalendarInfo) string {
		return cal.ID
	})
	translatedCalendarIDs, err := id.SetOutlookIDs(ctx, calendarIDs)
	if err != nil {
		return fmt.Errorf("failed to set calendar IDs: %w", err)
	}

	// Parse requested calendar IDs if provided
	var requestedCalendarIDs []string
	if calendarIDstring != "" {
		requestedCalendarIDs = strings.Split(calendarIDstring, ",")
		for i := range requestedCalendarIDs {
			requestedCalendarIDs[i] = strings.TrimSpace(requestedCalendarIDs[i])
		}
		// Get the map of IDs and convert to a slice of the mapped values
		idMap, err := id.GetOutlookIDs(ctx, requestedCalendarIDs)
		if err != nil {
			return fmt.Errorf("failed to get Outlook IDs: %w", err)
		}

		// Create a map for O(1) lookups directly from the idMap
		requestedIDsMap := make(map[string]struct{}, len(idMap))
		for _, mappedID := range idMap {
			requestedIDsMap[mappedID] = struct{}{}
		}

		// Filter calendars in a single pass
		filteredCalendars := make([]graph.CalendarInfo, 0, len(calendars))
		for _, cal := range calendars {
			if _, exists := requestedIDsMap[cal.ID]; exists {
				filteredCalendars = append(filteredCalendars, cal)
			}
		}
		calendars = filteredCalendars
	}

	calendarEvents := map[graph.CalendarInfo][]models.Eventable{}
	for _, cal := range calendars {
		if cal.ID == "" {
			continue
		}

		events, err := graph.ListCalendarView(ctx, c, cal.ID, cal.Owner, &start, &end, limitInt)
		if err != nil {
			return fmt.Errorf("failed to list events for calendar %s: %w", util.Deref(cal.Calendar.GetName()), err)
		}

		if len(events) == 0 {
			continue
		}

		// Update the ID to the translated ID
		cal.ID = translatedCalendarIDs[cal.ID]

		eventIDs := util.Map(events, func(event models.Eventable) string {
			return util.Deref(event.GetId())
		})
		translatedEventIDs, err := id.SetOutlookIDs(ctx, eventIDs)
		if err != nil {
			return fmt.Errorf("failed to set event IDs: %w", err)
		}

		for i := range events {
			events[i].SetId(util.Ptr(translatedEventIDs[util.Deref(events[i].GetId())]))
		}

		if len(events) > 0 {
			calendarEvents[cal] = events
		}
	}

	if len(calendarEvents) == 0 {
		fmt.Println("No events found")
		return nil
	}

	gptscriptClient, err := gptscript.NewGPTScript()
	if err != nil {
		return fmt.Errorf("failed to create GPTScript client: %w", err)
	}

	var elements []gptscript.DatasetElement
	for cal, events := range calendarEvents {
		for _, event := range events {
			name := util.Deref(event.GetId()) + "_" + util.Deref(event.GetSubject())
			elements = append(elements, gptscript.DatasetElement{
				DatasetElementMeta: gptscript.DatasetElementMeta{
					Name:        name,
					Description: util.Deref(event.GetBodyPreview()),
				},
				Contents: printers.EventToString(ctx, c, cal, event),
			})
		}
	}

	datasetID, err := gptscriptClient.CreateDatasetWithElements(ctx, elements, gptscript.DatasetOptions{
		Name:        "event_list",
		Description: "List of Outlook Calendar events",
	})
	if err != nil {
		return fmt.Errorf("failed to create dataset with elements: %w", err)
	}

	fmt.Printf("Created dataset with ID %s with %d events\n", datasetID, len(elements))
	return nil
}
