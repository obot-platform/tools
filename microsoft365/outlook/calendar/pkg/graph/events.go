package graph

import (
	"context"
	"fmt"
	"time"

	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/recurrence"
	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/util"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/groups"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
)

type CreateEventInfo struct {
	Attendees                               []string // slice of email addresses
	OptionalAttendees                       []string // slice of email addresses for optional attendees
	Subject, Location, Body, ID, Recurrence string
	Owner                                   OwnerType
	IsOnline                                bool
	Start, End                              time.Time
}

func GetEvent(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, eventID, calendarID string, owner OwnerType) (models.Eventable, error) {
	expand := []string{"attachments"}
	if calendarID != "" {
		switch owner {
		case OwnerTypeUser:
			requestParameters := &users.ItemCalendarsItemEventsEventItemRequestBuilderGetQueryParameters{
				Expand: expand,
			}
			configuration := &users.ItemCalendarsItemEventsEventItemRequestBuilderGetRequestConfiguration{
				QueryParameters: requestParameters,
			}
			resp, err := client.Me().Calendars().ByCalendarId(calendarID).Events().ByEventId(eventID).Get(ctx, configuration)
			if err != nil {
				return nil, fmt.Errorf("failed to get event: %w", err)
			}
			return resp, nil
		case OwnerTypeGroup:
			requestParameters := &groups.ItemEventsEventItemRequestBuilderGetQueryParameters{
				Expand: expand,
			}
			configuration := &groups.ItemEventsEventItemRequestBuilderGetRequestConfiguration{
				QueryParameters: requestParameters,
			}
			resp, err := client.Groups().ByGroupId(calendarID).Events().ByEventId(eventID).Get(ctx, configuration)
			if err != nil {
				return nil, fmt.Errorf("failed to get event: %w", err)
			}
			return resp, nil
		}
	}
	requestParameters := &users.ItemEventsEventItemRequestBuilderGetQueryParameters{
		Expand: expand,
	}
	configuration := &users.ItemEventsEventItemRequestBuilderGetRequestConfiguration{
		QueryParameters: requestParameters,
	}
	resp, err := client.Me().Events().ByEventId(eventID).Get(ctx, configuration)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return resp, nil
}

func CreateEvent(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, info CreateEventInfo) (models.Eventable, error) {
	requestBody := models.NewEvent()

	if info.Recurrence != "" {
		// Recurrence is pretty complicated in the graph API, so we use an internal tool call to generate it.
		r, err := recurrence.Generate(ctx, info.Recurrence)
		if err != nil {
			return nil, fmt.Errorf("failed to generate recurrence: %w", err)
		}

		graphRecurrence, err := r.ConvertForGraphAPI()
		if err != nil {
			return nil, fmt.Errorf("failed to convert recurrence for Graph API: %w", err)
		}

		requestBody.SetRecurrence(graphRecurrence)
	}

	var attendees []models.Attendeeable
	// Handle both required and optional attendees in a single loop
	for _, attendeeList := range []struct {
		emails []string
		type_  models.AttendeeType
	}{
		{info.Attendees, models.REQUIRED_ATTENDEETYPE},
		{info.OptionalAttendees, models.OPTIONAL_ATTENDEETYPE},
	} {
		for _, a := range attendeeList.emails {
			attendee := models.NewAttendee()
			email := models.NewEmailAddress()
			email.SetAddress(&a)
			attendee.SetEmailAddress(email)
			attendee.SetTypeEscaped(util.Ptr(attendeeList.type_))
			attendees = append(attendees, attendee)
		}
	}
	requestBody.SetAttendees(attendees)

	requestBody.SetSubject(&info.Subject)

	location := models.NewLocation()
	location.SetDisplayName(&info.Location)
	requestBody.SetLocation(location)

	body := models.NewItemBody()
	body.SetContent(&info.Body)
	body.SetContentType(util.Ptr(models.TEXT_BODYTYPE))
	requestBody.SetBody(body)

	requestBody.SetIsOnlineMeeting(&info.IsOnline)

	start := models.NewDateTimeTimeZone()
	start.SetDateTime(util.Ptr(info.Start.UTC().Format(time.RFC3339)))
	start.SetTimeZone(util.Ptr("UTC"))
	requestBody.SetStart(start)

	end := models.NewDateTimeTimeZone()
	end.SetDateTime(util.Ptr(info.End.UTC().Format(time.RFC3339)))
	end.SetTimeZone(util.Ptr("UTC"))
	requestBody.SetEnd(end)

	if info.ID != "" {
		switch info.Owner {
		case OwnerTypeUser:
			event, err := client.Me().Calendars().ByCalendarId(info.ID).Events().Post(ctx, requestBody, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create event: %w", err)
			}
			return event, nil
		case OwnerTypeGroup:
			event, err := client.Groups().ByGroupId(info.ID).Events().Post(ctx, requestBody, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create event: %w", err)
			}
			return event, nil
		default:
			return nil, fmt.Errorf("invalid owner type: %s (possible values are \"user\" and \"group\")", info.Owner)
		}
	}

	// Create the event in the user's default calendar.
	event, err := client.Me().Events().Post(ctx, requestBody, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}
	return event, nil
}

func DeleteEvent(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, eventID, calendarID string, owner OwnerType) error {
	if calendarID != "" {
		switch owner {
		case OwnerTypeUser:
			if err := client.Me().Calendars().ByCalendarId(calendarID).Events().ByEventId(eventID).Delete(ctx, nil); err != nil {
				return fmt.Errorf("failed to delete event: %w", err)
			}
			return nil
		case OwnerTypeGroup:
			if err := client.Groups().ByGroupId(calendarID).Events().ByEventId(eventID).Delete(ctx, nil); err != nil {
				return fmt.Errorf("failed to delete event: %w", err)
			}
			return nil
		}
	}

	if err := client.Me().Events().ByEventId(eventID).Delete(ctx, nil); err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	return nil
}

func AcceptEvent(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, eventID, calendarID string, owner OwnerType) error {
	requestBody := users.NewItemEventsItemAcceptPostRequestBody()
	requestBody.SetSendResponse(util.Ptr(true))

	if calendarID != "" {
		switch owner {
		case OwnerTypeUser:
			if err := client.Me().Calendars().ByCalendarId(calendarID).Events().ByEventId(eventID).Accept().Post(ctx, requestBody, nil); err != nil {
				return fmt.Errorf("failed to accept event: %w", err)
			}
			return nil
		case OwnerTypeGroup:
			if err := client.Groups().ByGroupId(calendarID).Events().ByEventId(eventID).Accept().Post(ctx, requestBody, nil); err != nil {
				return fmt.Errorf("failed to accept event: %w", err)
			}
			return nil
		}
	}

	if err := client.Me().Events().ByEventId(eventID).Accept().Post(ctx, requestBody, nil); err != nil {
		return fmt.Errorf("failed to accept event: %w", err)
	}
	return nil
}

func TentativelyAcceptEvent(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, eventID, calendarID string, owner OwnerType) error {
	requestBody := users.NewItemEventsItemTentativelyAcceptPostRequestBody()
	requestBody.SetSendResponse(util.Ptr(true))

	if calendarID != "" {
		switch owner {
		case OwnerTypeUser:
			if err := client.Me().Calendars().ByCalendarId(calendarID).Events().ByEventId(eventID).TentativelyAccept().Post(ctx, requestBody, nil); err != nil {
				return fmt.Errorf("failed to tentatively accept event: %w", err)
			}
			return nil
		case OwnerTypeGroup:
			if err := client.Groups().ByGroupId(calendarID).Events().ByEventId(eventID).TentativelyAccept().Post(ctx, requestBody, nil); err != nil {
				return fmt.Errorf("failed to tentatively accept event: %w", err)
			}
			return nil
		}
	}

	if err := client.Me().Events().ByEventId(eventID).TentativelyAccept().Post(ctx, requestBody, nil); err != nil {
		return fmt.Errorf("failed to tentatively accept event: %w", err)
	}
	return nil
}

func DeclineEvent(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, eventID, calendarID string, owner OwnerType) error {
	requestBody := users.NewItemEventsItemDeclinePostRequestBody()
	requestBody.SetSendResponse(util.Ptr(true))

	if calendarID != "" {
		switch owner {
		case OwnerTypeUser:
			if err := client.Me().Calendars().ByCalendarId(calendarID).Events().ByEventId(eventID).Decline().Post(ctx, requestBody, nil); err != nil {
				return fmt.Errorf("failed to decline event: %w", err)
			}
			return nil
		case OwnerTypeGroup:
			if err := client.Groups().ByGroupId(calendarID).Events().ByEventId(eventID).Decline().Post(ctx, requestBody, nil); err != nil {
				return fmt.Errorf("failed to decline event: %w", err)
			}
			return nil
		}
	}

	if err := client.Me().Events().ByEventId(eventID).Decline().Post(ctx, requestBody, nil); err != nil {
		return fmt.Errorf("failed to decline event: %w", err)
	}
	return nil
}

func ModifyEventAttendee(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, eventID, calendarID string, owner OwnerType, addRequiredAttendees, addOptionalAttendees, removeAttendees []string) error {
	// Get existing event
	event, err := GetEvent(ctx, client, eventID, calendarID, owner)
	if err != nil {
		return fmt.Errorf("failed to get event: %w", err)
	}

	// Get current attendees and create a map for easier lookup
	currentAttendees := event.GetAttendees()
	attendeeMap := make(map[string]models.Attendeeable)
	for _, a := range currentAttendees {
		if email := a.GetEmailAddress(); email != nil {
			if addr := email.GetAddress(); addr != nil {
				attendeeMap[*addr] = a
			}
		}
	}

	// Remove specified attendees
	for _, email := range removeAttendees {
		delete(attendeeMap, email)
	}

	// Add new attendees
	for _, attendeeList := range []struct {
		emails []string
		type_  models.AttendeeType
	}{
		{addRequiredAttendees, models.REQUIRED_ATTENDEETYPE},
		{addOptionalAttendees, models.OPTIONAL_ATTENDEETYPE},
	} {
		for _, a := range attendeeList.emails {
			if _, exists := attendeeMap[a]; !exists {
				attendee := models.NewAttendee()
				email := models.NewEmailAddress()
				email.SetAddress(&a)
				attendee.SetEmailAddress(email)
				attendee.SetTypeEscaped(util.Ptr(attendeeList.type_))
				attendeeMap[a] = attendee
			}
		}
	}

	// Convert map back to slice
	var updatedAttendees []models.Attendeeable
	for _, attendee := range attendeeMap {
		updatedAttendees = append(updatedAttendees, attendee)
	}

	// Create update body
	updateBody := models.NewEvent()
	updateBody.SetAttendees(updatedAttendees)

	// Update the event
	if calendarID != "" {
		switch owner {
		case OwnerTypeUser:
			if _, err := client.Me().Calendars().ByCalendarId(calendarID).Events().ByEventId(eventID).Patch(ctx, updateBody, nil); err != nil {
				return fmt.Errorf("failed to update event attendees: %w", err)
			}
		case OwnerTypeGroup:
			if _, err := client.Groups().ByGroupId(calendarID).Events().ByEventId(eventID).Patch(ctx, updateBody, nil); err != nil {
				return fmt.Errorf("failed to update event attendees: %w", err)
			}
		}
	} else {
		if _, err := client.Me().Events().ByEventId(eventID).Patch(ctx, updateBody, nil); err != nil {
			return fmt.Errorf("failed to update event attendees: %w", err)
		}
	}

	return nil
}
