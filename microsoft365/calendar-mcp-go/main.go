// Copyright 2025 The Go MCP SDK Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/google/jsonschema-go/jsonschema"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/groups"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var httpAddr = flag.String("http", ":3000", "HTTP address to listen on for streamable HTTP server")

// StaticTokenCredential implements azcore.TokenCredential
type StaticTokenCredential struct {
	token string
}

func (s StaticTokenCredential) GetToken(_ context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: s.token}, nil
}

// CalendarMCPServer wraps the Microsoft Graph client for Calendar operations
type CalendarMCPServer struct {
	client *msgraphsdkgo.GraphServiceClient
}

// NewCalendarMCPServer creates a new Calendar MCP server with the given token
func NewCalendarMCPServer(token string) (*CalendarMCPServer, error) {
	credential := StaticTokenCredential{token: token}
	client, err := msgraphsdkgo.NewGraphServiceClientWithCredentials(credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Graph client: %w", err)
	}

	return &CalendarMCPServer{client: client}, nil
}

// Argument structures with proper JSON schema tags based on tool.gpt
type ListCalendarsArgs struct{}

type ListEventsTodayArgs struct{}

type ListEventsArgs struct {
	Start       string  `json:"start" jsonschema:"(Required) The start date and time of the time frame, in RFC 3339 format."`
	End         string  `json:"end" jsonschema:"(Required) The end date and time of the time frame, in RFC 3339 format."`
	CalendarIDs *string `json:"calendar_ids,omitempty" jsonschema:"(Optional) A comma-separated list of the unique IDs of the calendars to list events from. If unset, lists events from all calendars."`
	Limit       *string `json:"limit,omitempty" jsonschema:"(Optional) The maximum number of events to return for each calendar. If unset, returns up to 50 events for each calendar."`
}

type GetEventDetailsArgs struct {
	EventID    string  `json:"event_id" jsonschema:"The unique ID of the event."`
	CalendarID *string `json:"calendar_id,omitempty" jsonschema:"The unique ID of the calendar or group the event belongs to. If unset, uses the default calendar."`
	OwnerType  *string `json:"owner_type,omitempty" jsonschema:"The type of the owner of the calendar or group. Possible values are 'user' or 'group'. Required if calendar_id is set."`
}

type GetEventAttachmentsArgs struct {
	EventID    string  `json:"event_id" jsonschema:"The unique ID of the event."`
	CalendarID *string `json:"calendar_id,omitempty" jsonschema:"The unique ID of the calendar or group the event belongs to. If unset, uses the default calendar."`
	OwnerType  *string `json:"owner_type,omitempty" jsonschema:"The type of the owner of the calendar or group. Possible values are 'user' or 'group'. Required if calendar_id is set."`
}

type CreateEventArgs struct {
	Subject           string  `json:"subject" jsonschema:"(Required) The title of the event."`
	Location          string  `json:"location" jsonschema:"(Required) The location of the event."`
	Body              string  `json:"body" jsonschema:"(Required) The details of the event."`
	Attendees         string  `json:"attendees" jsonschema:"(Required) A comma-separated list of the email addresses of people required to attend the event. Example: 'john@example.com,jane@example.com'"`
	OptionalAttendees *string `json:"optional_attendees,omitempty" jsonschema:"(Optional) A comma-separated list of the email addresses of people optionally invited to the event. Example: 'john@example.com,jane@example.com'"`
	IsOnline          bool    `json:"is_online" jsonschema:"(Required) (boolean) Whether the event is online (true) or in person (false)."`
	Start             string  `json:"start" jsonschema:"(Required) The start time of the event, in RFC 3339 format."`
	End               string  `json:"end" jsonschema:"(Required) The end time of the event, in RFC 3339 format. When scheduling a recurring event, this should be the end time of the first event in the series."`
	Recurrence        *string `json:"recurrence,omitempty" jsonschema:"(Optional) If the meeting should recur, describe in plain English how often it should occur (daily, weekly, monthly, yearly) and during which date range (first and last occurrence) or how many total times the event should occur. ALWAYS include the date of the first occurrence, and optionally the date of the last occurrence."`
	CalendarID        *string `json:"calendar_id,omitempty" jsonschema:"The unique ID of the calendar or group to add the event to. If unset, adds the event to the default calendar."`
	OwnerType         *string `json:"owner_type,omitempty" jsonschema:"(Required if calendar_id is set) The type of the owner of the calendar or group. Possible values are 'user' or 'group'."`
}

type ModifyEventAttendeesArgs struct {
	EventID              string  `json:"event_id" jsonschema:"(Required) The unique ID of the event."`
	CalendarID           *string `json:"calendar_id,omitempty" jsonschema:"(Optional) The unique ID of the calendar or group the event belongs to. If unset, uses the default calendar."`
	OwnerType            *string `json:"owner_type,omitempty" jsonschema:"(Optional) The type of the owner of the calendar or group. Possible values are 'user' or 'group'. Required if calendar_id is set."`
	AddRequiredAttendees *string `json:"add_required_attendees,omitempty" jsonschema:"(Optional) A comma-separated list of the email addresses of additional people required to attend the event."`
	AddOptionalAttendees *string `json:"add_optional_attendees,omitempty" jsonschema:"(Optional) A comma-separated list of the email addresses of additional people optionally invited to the event."`
	RemoveAttendees      *string `json:"remove_attendees,omitempty" jsonschema:"(Optional) A comma-separated list of the email addresses of people to remove from the event."`
}

type DeleteEventArgs struct {
	EventID      string  `json:"event_id" jsonschema:"(Required) The unique ID of the event."`
	DeleteSeries *bool   `json:"delete_series,omitempty" jsonschema:"(Optional) Whether to delete the entire series of recurring events. If true, all events in the series will be deleted. If false, only the specific event will be deleted. Default is false."`
	CalendarID   *string `json:"calendar_id,omitempty" jsonschema:"(Optional) The unique ID of the calendar or group the event belongs to. If unset, uses the default calendar."`
	OwnerType    *string `json:"owner_type,omitempty" jsonschema:"(Optional) The type of the owner of the calendar or group. Possible values are 'user' or 'group'. Required if calendar_id is set."`
}

type SearchEventsArgs struct {
	Query string `json:"query" jsonschema:"(Required) The search query."`
	Start string `json:"start" jsonschema:"(Required) The start date and time of the time frame to search within, in RFC 3339 format."`
	End   string `json:"end" jsonschema:"(Required) The end date and time of the time frame to search within, in RFC 3339 format."`
}

type RespondToEventArgs struct {
	EventID    string  `json:"event_id" jsonschema:"The unique ID of the event."`
	CalendarID *string `json:"calendar_id,omitempty" jsonschema:"The unique ID of the calendar or group the event belongs to. If unset, uses the default calendar."`
	OwnerType  *string `json:"owner_type,omitempty" jsonschema:"The type of the owner of the calendar or group. Possible values are 'user' or 'group'. Required if calendar_id is set."`
	Response   string  `json:"response" jsonschema:"The response to the invitation. Possible values are 'accept', 'tentative', or 'decline'."`
}

// Helper types
type OwnerType string

const (
	OwnerTypeUser  OwnerType = "user"
	OwnerTypeGroup OwnerType = "group"
)

// Event detail types for GetEventDetails response
type DetailedEventInfo struct {
	ID               string                       `json:"id"`
	Subject          string                       `json:"subject"`
	Start            string                       `json:"start"`
	End              string                       `json:"end"`
	Location         string                       `json:"location,omitempty"`
	IsOnline         bool                         `json:"is_online"`
	Body             string                       `json:"body,omitempty"`
	BodyPreview      string                       `json:"body_preview,omitempty"`
	Attendees        []DetailedAttendeeInfo       `json:"attendees,omitempty"`
	Organizer        *DetailedOrganizerInfo       `json:"organizer,omitempty"`
	IsAllDay         bool                         `json:"is_all_day"`
	ShowAs           string                       `json:"show_as,omitempty"`
	Sensitivity      string                       `json:"sensitivity,omitempty"`
	Importance       string                       `json:"importance,omitempty"`
	Categories       []string                     `json:"categories,omitempty"`
	Recurrence       map[string]interface{}       `json:"recurrence,omitempty"`
	Attachments      []DetailedAttachmentInfo     `json:"attachments,omitempty"`
	OnlineMeetingUrl string                       `json:"online_meeting_url,omitempty"`
}

type DetailedAttendeeInfo struct {
	Name   string `json:"name,omitempty"`
	Email  string `json:"email"`
	Type   string `json:"type"`
	Status string `json:"status,omitempty"`
}

type DetailedOrganizerInfo struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

type DetailedAttachmentInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ContentType string `json:"content_type,omitempty"`
	Size        int32  `json:"size,omitempty"`
}

// Helper functions
func deref[T any](v *T) (r T) {
	if v != nil {
		return *v
	}
	return
}

func parseStartEnd(start, end string, optional bool) (time.Time, time.Time, error) {
	var (
		startTime time.Time
		endTime   time.Time
		err       error
	)

	if start != "" {
		startTime, err = time.Parse(time.RFC3339, start)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("failed to parse start time: %w", err)
		}
	} else if !optional {
		return time.Time{}, time.Time{}, fmt.Errorf("start time is required")
	}

	if end != "" {
		endTime, err = time.Parse(time.RFC3339, end)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("failed to parse end time: %w", err)
		}
	} else if !optional {
		return time.Time{}, time.Time{}, fmt.Errorf("end time is required")
	}

	return startTime, endTime, nil
}

func readEmailsFromString(emails string) []string {
	if emails == "" {
		return nil
	}
	raw := strings.Split(emails, ",")
	var result []string
	for _, e := range raw {
		email := strings.TrimSpace(e)
		if email != "" {
			result = append(result, email)
		}
	}
	return result
}

func validateEmails(label string, emails []string) error {
	for _, email := range emails {
		if _, err := mail.ParseAddress(email); err != nil {
			return fmt.Errorf("invalid email address for %s: %s", label, email)
		}
	}
	return nil
}

func ptr[T any](v T) *T {
	return &v
}

// ListEvents lists events in a given time frame
func (c *CalendarMCPServer) ListEvents(ctx context.Context, req *mcp.CallToolRequest, args ListEventsArgs) (*mcp.CallToolResult, any, error) {
	startTime, endTime, err := parseStartEnd(args.Start, args.End, false)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to parse start/end times: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	limit := int32(50)
	if args.Limit != nil {
		if parsedLimit, err := strconv.ParseInt(*args.Limit, 10, 32); err == nil {
			limit = int32(parsedLimit)
		}
	}

	resp, err := c.client.Me().CalendarView().Get(ctx, &users.ItemCalendarViewRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemCalendarViewRequestBuilderGetQueryParameters{
			StartDateTime: ptr(startTime.Format(time.RFC3339)),
			EndDateTime:   ptr(endTime.Format(time.RFC3339)),
			Top:           ptr(limit),
			Orderby:       []string{"start/dateTime"},
		},
	})
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to list events: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	type EventInfo struct {
		ID       string   `json:"id"`
		Subject  string   `json:"subject"`
		Start    string   `json:"start"`
		End      string   `json:"end"`
		Location string   `json:"location,omitempty"`
		IsOnline bool     `json:"is_online"`
		Body     string   `json:"body,omitempty"`
		Attendees []string `json:"attendees,omitempty"`
	}

	var events []EventInfo
	for _, event := range resp.GetValue() {
		eventInfo := EventInfo{
			ID:       deref(event.GetId()),
			Subject:  deref(event.GetSubject()),
			IsOnline: deref(event.GetIsOnlineMeeting()),
		}

		if startTime := event.GetStart(); startTime != nil {
			eventInfo.Start = deref(startTime.GetDateTime())
		}
		if endTime := event.GetEnd(); endTime != nil {
			eventInfo.End = deref(endTime.GetDateTime())
		}
		if location := event.GetLocation(); location != nil {
			eventInfo.Location = deref(location.GetDisplayName())
		}
		if body := event.GetBody(); body != nil {
			eventInfo.Body = deref(body.GetContent())
		}

		// Extract attendees
		if attendees := event.GetAttendees(); len(attendees) > 0 {
			for _, attendee := range attendees {
				if email := attendee.GetEmailAddress(); email != nil {
					if addr := email.GetAddress(); addr != nil {
						eventInfo.Attendees = append(eventInfo.Attendees, *addr)
					}
				}
			}
		}

		events = append(events, eventInfo)
	}

	result, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to marshal event data: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(result),
			},
		},
	}, nil, nil
}

// CreateEvent creates a new calendar event
func (c *CalendarMCPServer) CreateEvent(ctx context.Context, req *mcp.CallToolRequest, args CreateEventArgs) (*mcp.CallToolResult, any, error) {
	// Parse and validate times
	startTime, endTime, err := parseStartEnd(args.Start, args.End, false)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to parse start/end times: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	// Parse attendees
	requiredAttendees := readEmailsFromString(args.Attendees)
	optionalAttendees := readEmailsFromString(deref(args.OptionalAttendees))

	// Validate emails
	if err := validateEmails("required attendees", requiredAttendees); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Email validation failed: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}
	if err := validateEmails("optional attendees", optionalAttendees); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Email validation failed: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	// Build the event request
	requestBody := models.NewEvent()
	requestBody.SetSubject(&args.Subject)
	requestBody.SetIsOnlineMeeting(&args.IsOnline)

	// Set location
	location := models.NewLocation()
	location.SetDisplayName(&args.Location)
	requestBody.SetLocation(location)

	// Set body
	body := models.NewItemBody()
	body.SetContent(&args.Body)
	body.SetContentType(ptr(models.TEXT_BODYTYPE))
	requestBody.SetBody(body)

	// Set times
	start := models.NewDateTimeTimeZone()
	start.SetDateTime(ptr(startTime.UTC().Format(time.RFC3339)))
	start.SetTimeZone(ptr("UTC"))
	requestBody.SetStart(start)

	end := models.NewDateTimeTimeZone()
	end.SetDateTime(ptr(endTime.UTC().Format(time.RFC3339)))
	end.SetTimeZone(ptr("UTC"))
	requestBody.SetEnd(end)

	// Set attendees
	var allAttendees []models.Attendeeable
	for _, attendeeList := range []struct {
		emails []string
		type_  models.AttendeeType
	}{
		{requiredAttendees, models.REQUIRED_ATTENDEETYPE},
		{optionalAttendees, models.OPTIONAL_ATTENDEETYPE},
	} {
		for _, email := range attendeeList.emails {
			attendee := models.NewAttendee()
			emailAddr := models.NewEmailAddress()
			emailAddr.SetAddress(&email)
			attendee.SetEmailAddress(emailAddr)
			attendee.SetTypeEscaped(ptr(attendeeList.type_))
			allAttendees = append(allAttendees, attendee)
		}
	}
	requestBody.SetAttendees(allAttendees)

	// Create the event
	event, err := c.client.Me().Events().Post(ctx, requestBody, nil)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to create event: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Event created successfully. ID: %s", deref(event.GetId())),
			},
		},
	}, nil, nil
}

// DeleteEvent deletes a calendar event
func (c *CalendarMCPServer) DeleteEvent(ctx context.Context, req *mcp.CallToolRequest, args DeleteEventArgs) (*mcp.CallToolResult, any, error) {
	// Handle series deletion if requested
	if args.DeleteSeries != nil && *args.DeleteSeries {
		// First get the event to check if it's part of a series
		event, err := c.client.Me().Events().ByEventId(args.EventID).Get(ctx, nil)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to get event to check series: %v", err),
					},
				},
				IsError: true,
			}, nil, err
		}

		seriesMasterID := event.GetSeriesMasterId()
		if seriesMasterID != nil {
			// Delete the series master event
			err := c.client.Me().Events().ByEventId(*seriesMasterID).Delete(ctx, nil)
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to delete event series: %v", err),
						},
					},
					IsError: true,
				}, nil, err
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Event series deleted successfully"),
					},
				},
			}, nil, nil
		}
	}

	// Delete single event
	err := c.client.Me().Events().ByEventId(args.EventID).Delete(ctx, nil)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to delete event %s: %v", args.EventID, err),
				},
			},
			IsError: true,
		}, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Event %s deleted successfully", args.EventID),
			},
		},
	}, nil, nil
}

// GetEventDetails gets detailed information about a specific event
func (c *CalendarMCPServer) GetEventDetails(ctx context.Context, req *mcp.CallToolRequest, args GetEventDetailsArgs) (*mcp.CallToolResult, any, error) {
	var event models.Eventable
	var err error

	if args.CalendarID != nil && args.OwnerType != nil {
		if *args.OwnerType == "group" {
			event, err = c.client.Groups().ByGroupId(*args.CalendarID).Events().ByEventId(args.EventID).Get(ctx, &groups.ItemEventsEventItemRequestBuilderGetRequestConfiguration{
				QueryParameters: &groups.ItemEventsEventItemRequestBuilderGetQueryParameters{
					Expand: []string{"attachments"},
				},
			})
		} else {
			event, err = c.client.Me().Calendars().ByCalendarId(*args.CalendarID).Events().ByEventId(args.EventID).Get(ctx, &users.ItemCalendarsItemEventsEventItemRequestBuilderGetRequestConfiguration{
				QueryParameters: &users.ItemCalendarsItemEventsEventItemRequestBuilderGetQueryParameters{
					Expand: []string{"attachments"},
				},
			})
		}
	} else {
		event, err = c.client.Me().Events().ByEventId(args.EventID).Get(ctx, &users.ItemEventsEventItemRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.ItemEventsEventItemRequestBuilderGetQueryParameters{
				Expand: []string{"attachments"},
			},
		})
	}

	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to get event details: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	eventInfo := DetailedEventInfo{
		ID:          deref(event.GetId()),
		Subject:     deref(event.GetSubject()),
		IsOnline:    deref(event.GetIsOnlineMeeting()),
		IsAllDay:    deref(event.GetIsAllDay()),
		BodyPreview: deref(event.GetBodyPreview()),
	}

	// Extract times
	if startTime := event.GetStart(); startTime != nil {
		eventInfo.Start = deref(startTime.GetDateTime())
	}
	if endTime := event.GetEnd(); endTime != nil {
		eventInfo.End = deref(endTime.GetDateTime())
	}
	if location := event.GetLocation(); location != nil {
		eventInfo.Location = deref(location.GetDisplayName())
	}
	if body := event.GetBody(); body != nil {
		eventInfo.Body = deref(body.GetContent())
	}
	if onlineMeeting := event.GetOnlineMeeting(); onlineMeeting != nil {
		eventInfo.OnlineMeetingUrl = deref(onlineMeeting.GetJoinUrl())
	}

	// Extract showAs, sensitivity, importance
	if showAs := event.GetShowAs(); showAs != nil {
		eventInfo.ShowAs = string(*showAs)
	}
	if sensitivity := event.GetSensitivity(); sensitivity != nil {
		eventInfo.Sensitivity = string(*sensitivity)
	}
	if importance := event.GetImportance(); importance != nil {
		eventInfo.Importance = string(*importance)
	}

	// Extract categories
	if categories := event.GetCategories(); len(categories) > 0 {
		eventInfo.Categories = categories
	}

	// Extract organizer
	if organizer := event.GetOrganizer(); organizer != nil {
		if emailAddr := organizer.GetEmailAddress(); emailAddr != nil {
			eventInfo.Organizer = &DetailedOrganizerInfo{
				Name:  deref(emailAddr.GetName()),
				Email: deref(emailAddr.GetAddress()),
			}
		}
	}

	// Extract attendees
	if attendees := event.GetAttendees(); len(attendees) > 0 {
		for _, attendee := range attendees {
			if email := attendee.GetEmailAddress(); email != nil {
				attendeeInfo := DetailedAttendeeInfo{
					Name:  deref(email.GetName()),
					Email: deref(email.GetAddress()),
				}
				if attendeeType := attendee.GetTypeEscaped(); attendeeType != nil {
					attendeeInfo.Type = string(*attendeeType)
				}
				if status := attendee.GetStatus(); status != nil {
					if response := status.GetResponse(); response != nil {
						attendeeInfo.Status = string(*response)
					}
				}
				eventInfo.Attendees = append(eventInfo.Attendees, attendeeInfo)
			}
		}
	}

	// Extract attachments
	if attachments := event.GetAttachments(); len(attachments) > 0 {
		for _, attachment := range attachments {
			attachmentInfo := DetailedAttachmentInfo{
				ID:   deref(attachment.GetId()),
				Name: deref(attachment.GetName()),
			}
			if contentType := attachment.GetContentType(); contentType != nil {
				attachmentInfo.ContentType = *contentType
			}
			if size := attachment.GetSize(); size != nil {
				attachmentInfo.Size = *size
			}
			eventInfo.Attachments = append(eventInfo.Attachments, attachmentInfo)
		}
	}

	// Extract recurrence pattern if present (simplified)
	if recurrence := event.GetRecurrence(); recurrence != nil {
		recurrenceMap := make(map[string]interface{})
		if pattern := recurrence.GetPattern(); pattern != nil {
			patternMap := make(map[string]interface{})
			if recurrenceType := pattern.GetTypeEscaped(); recurrenceType != nil {
				patternMap["type"] = string(*recurrenceType)
			}
			if interval := pattern.GetInterval(); interval != nil {
				patternMap["interval"] = *interval
			}
			if daysOfWeek := pattern.GetDaysOfWeek(); len(daysOfWeek) > 0 {
				days := make([]string, len(daysOfWeek))
				for i, day := range daysOfWeek {
					days[i] = string(day)
				}
				patternMap["daysOfWeek"] = days
			}
			recurrenceMap["pattern"] = patternMap
		}
		// Note: Range information not available in this API version
		if len(recurrenceMap) > 0 {
			eventInfo.Recurrence = recurrenceMap
		}
	}

	result, err := json.MarshalIndent(eventInfo, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to marshal event details: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(result),
			},
		},
	}, nil, nil
}

// GetEventAttachments downloads attachments from an event
func (c *CalendarMCPServer) GetEventAttachments(ctx context.Context, req *mcp.CallToolRequest, args GetEventAttachmentsArgs) (*mcp.CallToolResult, any, error) {
	var event models.Eventable
	var err error

	if args.CalendarID != nil && args.OwnerType != nil {
		if *args.OwnerType == "group" {
			event, err = c.client.Groups().ByGroupId(*args.CalendarID).Events().ByEventId(args.EventID).Get(ctx, &groups.ItemEventsEventItemRequestBuilderGetRequestConfiguration{
				QueryParameters: &groups.ItemEventsEventItemRequestBuilderGetQueryParameters{
					Expand: []string{"attachments"},
				},
			})
		} else {
			event, err = c.client.Me().Calendars().ByCalendarId(*args.CalendarID).Events().ByEventId(args.EventID).Get(ctx, &users.ItemCalendarsItemEventsEventItemRequestBuilderGetRequestConfiguration{
				QueryParameters: &users.ItemCalendarsItemEventsEventItemRequestBuilderGetQueryParameters{
					Expand: []string{"attachments"},
				},
			})
		}
	} else {
		event, err = c.client.Me().Events().ByEventId(args.EventID).Get(ctx, &users.ItemEventsEventItemRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.ItemEventsEventItemRequestBuilderGetQueryParameters{
				Expand: []string{"attachments"},
			},
		})
	}

	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to get event attachments: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	attachments := event.GetAttachments()
	if len(attachments) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "No attachments found for this event.",
				},
			},
		}, nil, nil
	}

	type AttachmentResult struct {
		Name        string `json:"name"`
		ContentType string `json:"content_type,omitempty"`
		Size        int32  `json:"size,omitempty"`
		Content     string `json:"content,omitempty"` // Base64 encoded for file attachments
		Type        string `json:"type"`
	}

	var results []AttachmentResult
	for _, attachment := range attachments {
		attachmentType := deref(attachment.GetOdataType())
		result := AttachmentResult{
			Name: deref(attachment.GetName()),
			Type: attachmentType,
		}

		if contentType := attachment.GetContentType(); contentType != nil {
			result.ContentType = *contentType
		}
		if size := attachment.GetSize(); size != nil {
			result.Size = *size
		}

		if attachmentType == "#microsoft.graph.fileAttachment" {
			if fileAttachment, ok := attachment.(*models.FileAttachment); ok {
				if contentBytes := fileAttachment.GetContentBytes(); contentBytes != nil {
					result.Content = string(contentBytes) // Base64 encoded content
				}
			}
		}

		results = append(results, result)
	}

	resultJSON, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to marshal attachment data: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(resultJSON),
			},
		},
	}, nil, nil
}

// ListCalendars lists all calendars available to the user
func (c *CalendarMCPServer) ListCalendars(ctx context.Context, req *mcp.CallToolRequest, args ListCalendarsArgs) (*mcp.CallToolResult, any, error) {
	// Get user calendars
	calendarsResp, err := c.client.Me().Calendars().Get(ctx, &users.ItemCalendarsRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemCalendarsRequestBuilderGetQueryParameters{
			Top: ptr(int32(100)),
		},
	})
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to list calendars: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	type CalendarInfo struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		OwnerType string `json:"owner_type"`
		CanEdit   bool   `json:"can_edit"`
	}

	var calendars []CalendarInfo
	for _, calendar := range calendarsResp.GetValue() {
		calendars = append(calendars, CalendarInfo{
			ID:        deref(calendar.GetId()),
			Name:      deref(calendar.GetName()),
			OwnerType: "user",
			CanEdit:   deref(calendar.GetCanEdit()),
		})
	}

	// Get group calendars if available
	memberOf, err := c.client.Me().MemberOf().Get(ctx, &users.ItemMemberOfRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemMemberOfRequestBuilderGetQueryParameters{
			Top: ptr(int32(100)),
		},
	})
	if err == nil {
		for _, group := range memberOf.GetValue() {
			groupCalendar, err := c.client.Groups().ByGroupId(deref(group.GetId())).Calendar().Get(ctx, nil)
			if err == nil {
				calendars = append(calendars, CalendarInfo{
					ID:        deref(group.GetId()),
					Name:      deref(groupCalendar.GetName()),
					OwnerType: "group",
					CanEdit:   deref(groupCalendar.GetCanEdit()),
				})
			}
		}
	}

	result, err := json.MarshalIndent(calendars, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to marshal calendar data: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(result),
			},
		},
	}, nil, nil
}

// ListEventsToday lists all events for today in all calendars available to the user
func (c *CalendarMCPServer) ListEventsToday(ctx context.Context, req *mcp.CallToolRequest, args ListEventsTodayArgs) (*mcp.CallToolResult, any, error) {
	timezone := "UTC"
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, loc)

	// Use the calendar view to get today's events
	resp, err := c.client.Me().CalendarView().Get(ctx, &users.ItemCalendarViewRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemCalendarViewRequestBuilderGetQueryParameters{
			StartDateTime: ptr(start.Format(time.RFC3339)),
			EndDateTime:   ptr(end.Format(time.RFC3339)),
			Top:           ptr(int32(50)),
			Orderby:       []string{"start/dateTime"},
		},
	})
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to list today's events: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	type EventInfo struct {
		ID       string `json:"id"`
		Subject  string `json:"subject"`
		Start    string `json:"start"`
		End      string `json:"end"`
		Location string `json:"location,omitempty"`
		IsOnline bool   `json:"is_online"`
		Body     string `json:"body,omitempty"`
	}

	var events []EventInfo
	for _, event := range resp.GetValue() {
		eventInfo := EventInfo{
			ID:       deref(event.GetId()),
			Subject:  deref(event.GetSubject()),
			IsOnline: deref(event.GetIsOnlineMeeting()),
		}

		// Format start and end times
		if startTime := event.GetStart(); startTime != nil {
			eventInfo.Start = deref(startTime.GetDateTime())
		}
		if endTime := event.GetEnd(); endTime != nil {
			eventInfo.End = deref(endTime.GetDateTime())
		}
		if location := event.GetLocation(); location != nil {
			eventInfo.Location = deref(location.GetDisplayName())
		}
		if body := event.GetBody(); body != nil {
			eventInfo.Body = deref(body.GetContent())
		}

		events = append(events, eventInfo)
	}

	result, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to marshal event data: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(result),
			},
		},
	}, nil, nil
}

// ModifyEventAttendees adds or removes attendees from an event
func (c *CalendarMCPServer) ModifyEventAttendees(ctx context.Context, req *mcp.CallToolRequest, args ModifyEventAttendeesArgs) (*mcp.CallToolResult, any, error) {
	// Parse attendee lists
	var addRequired, addOptional, remove []string

	if args.AddRequiredAttendees != nil {
		addRequired = readEmailsFromString(*args.AddRequiredAttendees)
		if err := validateEmails("add required attendees", addRequired); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Email validation failed: %v", err),
					},
				},
				IsError: true,
			}, nil, err
		}
	}

	if args.AddOptionalAttendees != nil {
		addOptional = readEmailsFromString(*args.AddOptionalAttendees)
		if err := validateEmails("add optional attendees", addOptional); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Email validation failed: %v", err),
					},
				},
				IsError: true,
			}, nil, err
		}
	}

	if args.RemoveAttendees != nil {
		remove = readEmailsFromString(*args.RemoveAttendees)
		if err := validateEmails("remove attendees", remove); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Email validation failed: %v", err),
					},
				},
				IsError: true,
			}, nil, err
		}
	}

	// Get current event to modify attendees
	var event models.Eventable
	var err error

	if args.CalendarID != nil && args.OwnerType != nil {
		if *args.OwnerType == "group" {
			event, err = c.client.Groups().ByGroupId(*args.CalendarID).Events().ByEventId(args.EventID).Get(ctx, nil)
		} else {
			event, err = c.client.Me().Calendars().ByCalendarId(*args.CalendarID).Events().ByEventId(args.EventID).Get(ctx, nil)
		}
	} else {
		event, err = c.client.Me().Events().ByEventId(args.EventID).Get(ctx, nil)
	}

	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to get event: %v", err),
				},
			},
			IsError: true,
		}, nil, err
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
	for _, email := range remove {
		delete(attendeeMap, email)
	}

	// Add new attendees
	for _, attendeeList := range []struct {
		emails []string
		type_  models.AttendeeType
	}{
		{addRequired, models.REQUIRED_ATTENDEETYPE},
		{addOptional, models.OPTIONAL_ATTENDEETYPE},
	} {
		for _, email := range attendeeList.emails {
			if _, exists := attendeeMap[email]; !exists {
				attendee := models.NewAttendee()
				emailAddr := models.NewEmailAddress()
				emailAddr.SetAddress(&email)
				attendee.SetEmailAddress(emailAddr)
				attendee.SetTypeEscaped(ptr(attendeeList.type_))
				attendeeMap[email] = attendee
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
	if args.CalendarID != nil && args.OwnerType != nil {
		if *args.OwnerType == "group" {
			_, err = c.client.Groups().ByGroupId(*args.CalendarID).Events().ByEventId(args.EventID).Patch(ctx, updateBody, nil)
		} else {
			_, err = c.client.Me().Calendars().ByCalendarId(*args.CalendarID).Events().ByEventId(args.EventID).Patch(ctx, updateBody, nil)
		}
	} else {
		_, err = c.client.Me().Events().ByEventId(args.EventID).Patch(ctx, updateBody, nil)
	}

	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to update event attendees: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: "Successfully updated event attendees",
			},
		},
	}, nil, nil
}

// SearchEvents searches for events based on a query string
func (c *CalendarMCPServer) SearchEvents(ctx context.Context, req *mcp.CallToolRequest, args SearchEventsArgs) (*mcp.CallToolResult, any, error) {
	startTime, endTime, err := parseStartEnd(args.Start, args.End, false)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to parse start/end times: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	// Get all events in the time range
	resp, err := c.client.Me().CalendarView().Get(ctx, &users.ItemCalendarViewRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemCalendarViewRequestBuilderGetQueryParameters{
			StartDateTime: ptr(startTime.Format(time.RFC3339)),
			EndDateTime:   ptr(endTime.Format(time.RFC3339)),
			Top:           ptr(int32(100)),
			Orderby:       []string{"start/dateTime"},
		},
	})
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to search events: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	// Filter events based on query
	query := strings.ToLower(args.Query)
	var matchingEvents []models.Eventable

	for _, event := range resp.GetValue() {
		subject := strings.ToLower(deref(event.GetSubject()))
		bodyPreview := strings.ToLower(deref(event.GetBodyPreview()))

		if strings.Contains(subject, query) || strings.Contains(bodyPreview, query) {
			matchingEvents = append(matchingEvents, event)
		}
	}

	if len(matchingEvents) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "No events found matching the search query.",
				},
			},
		}, nil, nil
	}

	type SearchEventInfo struct {
		ID          string   `json:"id"`
		Subject     string   `json:"subject"`
		Start       string   `json:"start"`
		End         string   `json:"end"`
		Location    string   `json:"location,omitempty"`
		IsOnline    bool     `json:"is_online"`
		BodyPreview string   `json:"body_preview,omitempty"`
		Attendees   []string `json:"attendees,omitempty"`
	}

	var events []SearchEventInfo
	for _, event := range matchingEvents {
		eventInfo := SearchEventInfo{
			ID:          deref(event.GetId()),
			Subject:     deref(event.GetSubject()),
			IsOnline:    deref(event.GetIsOnlineMeeting()),
			BodyPreview: deref(event.GetBodyPreview()),
		}

		if startTime := event.GetStart(); startTime != nil {
			eventInfo.Start = deref(startTime.GetDateTime())
		}
		if endTime := event.GetEnd(); endTime != nil {
			eventInfo.End = deref(endTime.GetDateTime())
		}
		if location := event.GetLocation(); location != nil {
			eventInfo.Location = deref(location.GetDisplayName())
		}

		// Extract attendees
		if attendees := event.GetAttendees(); len(attendees) > 0 {
			for _, attendee := range attendees {
				if email := attendee.GetEmailAddress(); email != nil {
					if addr := email.GetAddress(); addr != nil {
						eventInfo.Attendees = append(eventInfo.Attendees, *addr)
					}
				}
			}
		}

		events = append(events, eventInfo)
	}

	result, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to marshal search results: %v", err),
				},
			},
			IsError: true,
		}, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(result),
			},
		},
	}, nil, nil
}

// RespondToEvent accepts, tentatively accepts, or declines an event invitation
func (c *CalendarMCPServer) RespondToEvent(ctx context.Context, req *mcp.CallToolRequest, args RespondToEventArgs) (*mcp.CallToolResult, any, error) {
	var err error

	switch args.Response {
	case "accept":
		requestBody := users.NewItemEventsItemAcceptPostRequestBody()
		requestBody.SetSendResponse(ptr(true))

		if args.CalendarID != nil && args.OwnerType != nil {
			if *args.OwnerType == "group" {
				err = c.client.Groups().ByGroupId(*args.CalendarID).Events().ByEventId(args.EventID).Accept().Post(ctx, requestBody, nil)
			} else {
				err = c.client.Me().Calendars().ByCalendarId(*args.CalendarID).Events().ByEventId(args.EventID).Accept().Post(ctx, requestBody, nil)
			}
		} else {
			err = c.client.Me().Events().ByEventId(args.EventID).Accept().Post(ctx, requestBody, nil)
		}

		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to accept event: %v", err),
					},
				},
				IsError: true,
			}, nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Event accepted successfully",
				},
			},
		}, nil, nil

	case "tentative":
		requestBody := users.NewItemEventsItemTentativelyAcceptPostRequestBody()
		requestBody.SetSendResponse(ptr(true))

		if args.CalendarID != nil && args.OwnerType != nil {
			if *args.OwnerType == "group" {
				err = c.client.Groups().ByGroupId(*args.CalendarID).Events().ByEventId(args.EventID).TentativelyAccept().Post(ctx, requestBody, nil)
			} else {
				err = c.client.Me().Calendars().ByCalendarId(*args.CalendarID).Events().ByEventId(args.EventID).TentativelyAccept().Post(ctx, requestBody, nil)
			}
		} else {
			err = c.client.Me().Events().ByEventId(args.EventID).TentativelyAccept().Post(ctx, requestBody, nil)
		}

		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to tentatively accept event: %v", err),
					},
				},
				IsError: true,
			}, nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Event tentatively accepted successfully",
				},
			},
		}, nil, nil

	case "decline":
		requestBody := users.NewItemEventsItemDeclinePostRequestBody()
		requestBody.SetSendResponse(ptr(true))

		if args.CalendarID != nil && args.OwnerType != nil {
			if *args.OwnerType == "group" {
				err = c.client.Groups().ByGroupId(*args.CalendarID).Events().ByEventId(args.EventID).Decline().Post(ctx, requestBody, nil)
			} else {
				err = c.client.Me().Calendars().ByCalendarId(*args.CalendarID).Events().ByEventId(args.EventID).Decline().Post(ctx, requestBody, nil)
			}
		} else {
			err = c.client.Me().Events().ByEventId(args.EventID).Decline().Post(ctx, requestBody, nil)
		}

		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decline event: %v", err),
					},
				},
				IsError: true,
			}, nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Event declined successfully",
				},
			},
		}, nil, nil

	default:
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Invalid response: %s. Valid responses are: accept, tentative, decline", args.Response),
				},
			},
			IsError: true,
		}, nil, fmt.Errorf("invalid response: %s", args.Response)
	}
}

// ExtractTokenFromRequest extracts the bearer token from HTTP request headers
func ExtractTokenFromRequest(req *http.Request) (string, error) {
	// Try X-Forwarded-Access-Token first
	if token := req.Header.Get("X-Forwarded-Access-Token"); token != "" {
		return token, nil
	}

	// Try Authorization header
	if authHeader := req.Header.Get("Authorization"); authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer "), nil
		}
	}

	return "", fmt.Errorf("no access token found in request headers")
}

func main() {
	flag.Parse()

	// Create server factory that extracts token from each request
	serverFactory := func(req *http.Request) *mcp.Server {
		token, err := ExtractTokenFromRequest(req)
		if err != nil {
			log.Printf("Failed to extract token from request: %v", err)
			// Return a server that will fail gracefully
			server := mcp.NewServer(&mcp.Implementation{Name: "calendar-mcp-server"}, nil)
			return server
		}

		calendarServer, err := NewCalendarMCPServer(token)
		if err != nil {
			log.Printf("Failed to create Calendar MCP server: %v", err)
			// Return a server that will fail gracefully
			server := mcp.NewServer(&mcp.Implementation{Name: "calendar-mcp-server"}, nil)
			return server
		}

		server := mcp.NewServer(&mcp.Implementation{Name: "calendar-mcp-server"}, nil)

		// Create JSON schemas for the tools
		listEventsSchema, _ := jsonschema.For[ListEventsArgs](nil)
		getEventDetailsSchema, _ := jsonschema.For[GetEventDetailsArgs](nil)
		getEventAttachmentsSchema, _ := jsonschema.For[GetEventAttachmentsArgs](nil)
		createEventSchema, _ := jsonschema.For[CreateEventArgs](nil)
		modifyEventAttendeesSchema, _ := jsonschema.For[ModifyEventAttendeesArgs](nil)
		deleteEventSchema, _ := jsonschema.For[DeleteEventArgs](nil)
		searchEventsSchema, _ := jsonschema.For[SearchEventsArgs](nil)
		respondToEventSchema, _ := jsonschema.For[RespondToEventArgs](nil)

		// Register all tools with proper schemas - matching tool.gpt exactly
		mcp.AddTool(server, &mcp.Tool{
			Name:        "list_calendars",
			Description: "List all calendars available to the user.",
		}, calendarServer.ListCalendars)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "list_events_today",
			Description: "List all events for today in all calendars available to the user.",
		}, calendarServer.ListEventsToday)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "list_events",
			Description: "List all events in the given time frame in all calendars available to the user.",
			InputSchema: listEventsSchema,
		}, calendarServer.ListEvents)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "create_event",
			Description: "Create a new calendar event.",
			InputSchema: createEventSchema,
		}, calendarServer.CreateEvent)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "get_event_details",
			Description: "Get the details for a particular event.",
			InputSchema: getEventDetailsSchema,
		}, calendarServer.GetEventDetails)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "get_event_attachments",
			Description: "Download the attachments for a particular event.",
			InputSchema: getEventAttachmentsSchema,
		}, calendarServer.GetEventAttachments)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "modify_event_attendees",
			Description: "Adds or removes attendees from an existing event.",
			InputSchema: modifyEventAttendeesSchema,
		}, calendarServer.ModifyEventAttendees)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "delete_event",
			Description: "Delete a calendar event.",
			InputSchema: deleteEventSchema,
		}, calendarServer.DeleteEvent)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "search_events",
			Description: "Search for events based on a query string.",
			InputSchema: searchEventsSchema,
		}, calendarServer.SearchEvents)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "respond_to_event",
			Description: "Accept, tentatively accept, or decline an event invitation.",
			InputSchema: respondToEventSchema,
		}, calendarServer.RespondToEvent)

		return server
	}

	if *httpAddr != "" {
		handler := mcp.NewStreamableHTTPHandler(serverFactory, nil)
		log.Printf("Calendar MCP server listening at %s", *httpAddr)
		if err := http.ListenAndServe(*httpAddr, handler); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("HTTP address is required")
	}
}