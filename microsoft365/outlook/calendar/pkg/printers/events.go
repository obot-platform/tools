package printers

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/jaytaylor/html2text"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/util"
)

func EventToString(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, calendar graph.CalendarInfo, event models.Eventable) string {
	var calendarName string
	if calendar.Owner == graph.OwnerTypeUser {
		calendarName = util.Deref(calendar.Calendar.GetName())
	} else {
		groupName, err := graph.GetGroupNameFromID(ctx, client, calendar.ID)
		if err != nil {
			calendarName = calendar.ID
		} else {
			calendarName = groupName
		}
	}

	var sb strings.Builder
	sb.WriteString("Subject: " + util.Deref(event.GetSubject()) + "\n")
	sb.WriteString("  ID: " + util.Deref(event.GetId()) + "\n")

	isAllDay := util.Deref(event.GetIsAllDay())
	if isAllDay {
		sb.WriteString("  Time: All Day Event\n")
	} else {
		startTimeConverted, startTZDisplay, endTimeConverted, endTZDisplay := convertEventTimesToUserTimezone(event)
		sb.WriteString("  Start Time: " + startTimeConverted + " " + startTZDisplay + "\n")
		sb.WriteString("  End Time: " + endTimeConverted + " " + endTZDisplay + "\n")
	}

	isRecurring := "No"
	if event.GetSeriesMasterId() != nil {
		isRecurring = "Yes"
	}
	sb.WriteString("  Recurring: " + isRecurring + "\n")

	sb.WriteString("  In calendar: " + calendarName + " (ID " + calendar.ID + ")\n")
	if calendar.Calendar.GetOwner() != nil {
		sb.WriteString("  Owner: " + util.Deref(calendar.Calendar.GetOwner().GetName()) + " (" + util.Deref(calendar.Calendar.GetOwner().GetAddress()) + ")\n")
		sb.WriteString("  Owner Type: " + string(calendar.Owner) + "\n")
	}
	return sb.String()
}

func PrintEvent(event models.Eventable, detailed bool) {
	fmt.Printf("Subject: %s\n", util.Deref(event.GetSubject()))
	fmt.Printf("  ID: %s\n", util.Deref(event.GetId()))

	isAllDay := util.Deref(event.GetIsAllDay())
	if isAllDay {
		fmt.Printf("  Time: All Day Event\n")
	} else {
		startTimeConverted, startTZDisplay, endTimeConverted, endTZDisplay := convertEventTimesToUserTimezone(event)

		fmt.Printf("  Start Time: %s %s\n", startTimeConverted, startTZDisplay)
		fmt.Printf("  End Time: %s %s\n", endTimeConverted, endTZDisplay)
	}

	if event.GetSeriesMasterId() != nil {
		fmt.Printf("  Recurring: Yes\n")
	} else {
		fmt.Printf("  Recurring: No\n")
	}

	if detailed {
		fmt.Printf("  Location: %s\n", util.Deref(event.GetLocation().GetDisplayName()))
		fmt.Printf("  Is Cancelled: %t\n", util.Deref(event.GetIsCancelled()))
		fmt.Printf("  Is Online Meeting: %t\n", util.Deref(event.GetIsOnlineMeeting()))
		fmt.Printf("  Response Status: %s\n", event.GetResponseStatus().GetResponse().String())
		fmt.Printf("  Attendees: %s\n", strings.Join(util.Map(event.GetAttendees(), func(a models.Attendeeable) string {
			return fmt.Sprintf("%s (%s), Response: %s", util.Deref(a.GetEmailAddress().GetName()), util.Deref(a.GetEmailAddress().GetAddress()), a.GetStatus().GetResponse().String())
		}), ", "))
		body, err := html2text.FromString(util.Deref(event.GetBody().GetContent()), html2text.Options{
			PrettyTables: true,
		})
		if err == nil {
			fmt.Printf("  Body: %s\n", strings.ReplaceAll(body, "\n", "\n  "))
			fmt.Printf("  (End Body)\n")
		}
		attachments := event.GetAttachments()
		if len(attachments) > 0 {
			for _, attachment := range attachments {
				attachmentType := util.Deref(attachment.GetOdataType())
				if attachmentType == "#microsoft.graph.fileAttachment" {
					fileAttachment := attachment.(*models.FileAttachment)
					fmt.Printf("File Attachment: %s, Size: %d bytes, Content Type: %s\n", *fileAttachment.GetName(), *fileAttachment.GetSize(), *fileAttachment.GetContentType())
				} else if attachmentType == "#microsoft.graph.itemAttachment" {
					itemAttachment := attachment.(*models.ItemAttachment)
					fmt.Printf("Item Attachment: %s\n", *itemAttachment.GetName())
				}
			}
		}
		fmt.Printf("You can open the event using this link: %s\n", util.Deref(event.GetWebLink()))
	}
	fmt.Println()
}

func EventDisplayTimeZone(event models.Eventable) (string, string) {
	// No TZ for all day events to avoid messing up the start/end times during conversion
	startTZ, endTZ := "", ""
	if util.Deref(event.GetIsAllDay()) {
		return startTZ, endTZ
	}
	// Assume that timestamps are UTC by default, but verify
	if util.Deref(event.GetStart().GetTimeZone()) == "UTC" {
		startTZ = "Z"
	} else {
		startTZ = " " + util.Deref(event.GetStart().GetTimeZone())
	}
	if util.Deref(event.GetEnd().GetTimeZone()) == "UTC" {
		endTZ = "Z"
		endTZ = " " + util.Deref(event.GetEnd().GetTimeZone())
	}
	return startTZ, endTZ
}

// convertTimeStringToUserTimezone converts a time string from source timezone to user timezone
func convertTimeStringToUserTimezone(timeStr, sourceTZ, userTZ string) (string, string) {
	// If it's empty or no conversion is needed
	if timeStr == "" || userTZ == sourceTZ {
		return timeStr, sourceTZ
	}

	layout := "2006-01-02T15:04:05.0000000" // a hardcoded layout for parsing the time string

	// Load source location
	srcLoc, err := time.LoadLocation(sourceTZ)
	if err != nil {
		fmt.Printf("Error loading source timezone: %s, error: %s\n", sourceTZ, err)
		return timeStr, sourceTZ
	}

	// Parse the time string *in* the source location
	t, err := time.ParseInLocation(layout, timeStr, srcLoc)
	if err != nil {
		fmt.Printf("Error parsing time string: %s\n", err)
		return timeStr, sourceTZ
	}

	// Load user location
	userLoc, err := time.LoadLocation(userTZ)
	if err != nil {
		fmt.Printf("Error loading user timezone: %s, error: %s\n", userTZ, err)
		return timeStr, sourceTZ
	}

	// Convert to user's timezone
	t = t.In(userLoc)
	return t.Format(layout), userTZ
}

// convertEventTimesToUserTimezone converts both start and end times of an event to the user's timezone
func convertEventTimesToUserTimezone(event models.Eventable) (start, startTZ, end, endTZ string) {
	userTZ := os.Getenv("OBOT_USER_TIMEZONE")
	if userTZ == "" {
		userTZ = "UTC"
	}

	startTime := util.Deref(event.GetStart().GetDateTime())
	startTZSource := util.Deref(event.GetStart().GetTimeZone())
	start, startTZ = convertTimeStringToUserTimezone(startTime, startTZSource, userTZ)

	endTime := util.Deref(event.GetEnd().GetDateTime())
	endTZSource := util.Deref(event.GetEnd().GetTimeZone())
	end, endTZ = convertTimeStringToUserTimezone(endTime, endTZSource, userTZ)

	return start, startTZ, end, endTZ
}
