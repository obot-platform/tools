package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/commands"
	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/graph"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: calendar <command>")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "listCalendars":
		if err := commands.ListCalendars(context.Background()); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "listEventsToday":
		timezone := os.Getenv("OBOT_USER_TIMEZONE")
		if timezone == "" {
			timezone = "UTC" // default fallback
		}
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			fmt.Printf("Error loading user timezone: %s. Error: %v. Falling back to UTC.\n", timezone, err)
			timezone = "UTC"
			loc = time.UTC // Use UTC location after error
		}
		now := time.Now().In(loc)
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, loc)
		if err := commands.ListEvents(context.Background(), "", start, end, ""); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "listEvents":
		start, end, err := parseStartEnd(os.Getenv("START"), os.Getenv("END"), false)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if err := commands.ListEvents(context.Background(), os.Getenv("CALENDAR_IDS"), start, end, os.Getenv("LIMIT")); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "getEventDetails":
		if err := commands.GetEventDetails(context.Background(), os.Getenv("EVENT_ID"), os.Getenv("CALENDAR_ID"), graph.OwnerType(os.Getenv("OWNER_TYPE"))); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "getEventAttachments":
		if err := commands.GetEventAttachments(context.Background(), os.Getenv("EVENT_ID"), os.Getenv("CALENDAR_ID"), graph.OwnerType(os.Getenv("OWNER_TYPE"))); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "createEvent":
		info := graph.CreateEventInfo{
			Attendees:  strings.Split(os.Getenv("ATTENDEES"), ","),
			OptionalAttendees: strings.Split(os.Getenv("OPTIONAL_ATTENDEES"), ","),
			Subject:    os.Getenv("SUBJECT"),
			Location:   os.Getenv("LOCATION"),
			Body:       os.Getenv("BODY"),
			Recurrence: os.Getenv("RECURRENCE"),
		}

		// Unset the BODY variable so that it does not mess up writing files to the workspace later on.
		if err := os.Unsetenv("BODY"); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		isOnline, err := strconv.ParseBool(os.Getenv("IS_ONLINE"))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		info.IsOnline = isOnline

		if id := os.Getenv("CALENDAR_ID"); id != "" {
			info.ID = id
			info.Owner = graph.OwnerType(os.Getenv("OWNER_TYPE"))

			if info.Owner == "" {
				fmt.Println("Owner type is required")
				os.Exit(1)
			}
		}

		start, end, err := parseStartEnd(os.Getenv("START"), os.Getenv("END"), false)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		info.Start = start
		info.End = end

		if err := commands.CreateEvent(context.Background(), info); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "modifyEventAttendees":
		if err := commands.ModifyEventAttendees(context.Background(), os.Getenv("EVENT_ID"), os.Getenv("CALENDAR_ID"), graph.OwnerType(os.Getenv("OWNER_TYPE")), os.Getenv("ADD_REQUIRED_ATTENDEES"), os.Getenv("ADD_OPTIONAL_ATTENDEES"), os.Getenv("REMOVE_ATTENDEES")); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "deleteEvent":
		if err := commands.DeleteEvent(context.Background(), os.Getenv("EVENT_ID"), os.Getenv("CALENDAR_ID"), graph.OwnerType(os.Getenv("OWNER_TYPE")), os.Getenv("DELETE_SERIES") == "true"); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "searchEvents":
		start, end, err := parseStartEnd(os.Getenv("START"), os.Getenv("END"), false)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if err := commands.SearchEvents(context.Background(), os.Getenv("QUERY"), start, end); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "respondToEvent":
		if err := commands.RespondToEvent(context.Background(), os.Getenv("EVENT_ID"), os.Getenv("CALENDAR_ID"), graph.OwnerType(os.Getenv("OWNER_TYPE")), os.Getenv("RESPONSE")); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %q\n", command)
		os.Exit(1)
	}
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
