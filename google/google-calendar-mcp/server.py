from fastmcp import FastMCP
from pydantic import Field
from typing import Annotated, Literal
import os
from tools.helper import setup_logger, get_client, get_user_timezone
from googleapiclient.errors import HttpError
from fastmcp.exceptions import ToolError
from rfc3339_validator import validate_rfc3339
from tools.event import (
    DEFAULT_MAX_RESULTS,
    GOOGLE_EVENT_TYPE_OPTIONS,
    MOVABLE_EVENT_TYPES,
    get_current_time_rfc3339,
    validate_recurrence_list,
    is_valid_date,
    is_valid_iana_timezone,
    get_current_user_email,
    can_update_property,
    has_calendar_write_access,
)
logger = setup_logger(__name__)

# Configure server-specific settings
PORT = os.getenv("PORT", 9000)
MCP_PATH = os.getenv("MCP_PATH", "/mcp/google-drive")

mcp = FastMCP(
    name="GoogleDriveMCPServer",
    on_duplicate_tools="error",                  # Handle duplicate registrations
    on_duplicate_resources="warn",
    on_duplicate_prompts="replace",
    mask_error_details=True, # only include details for ToolError and convert other errors to generic error\
)

@mcp.tool(
    annotations={
        "readOnlyHint": True,
        "destructiveHint": False,
    },
    exclude_args=["cred_token"], # the access token, is excluded to LLM
)
def list_calendars(cred_token: str = None) -> list:
    """Lists all calendars for the authenticated user."""
    try:
        service = get_client(cred_token)
        calendars = service.calendarList().list().execute()
        return calendars.get("items", [])
    except HttpError as err:
        raise ToolError(f"Failed to list google calendars. HttpError: {err}")


@mcp.tool(
    annotations={
        "readOnlyHint": True,
        "destructiveHint": False,
    },
    exclude_args=["cred_token"], # the access token, is excluded to LLM
)
def get_calendar(calendar_id: Annotated[str, Field(description="calendar id to get")],
                 cred_token: str = None) -> dict:
    """Get details of a specific google calendar"""
    if calendar_id == "":
        raise ValueError("argument `calendar_id` can't be empty")
    service = get_client(cred_token)
    try:
        calendar = service.calendars().get(calendarId=calendar_id).execute()
        return calendar
    except HttpError as err:
        raise ToolError(f"Failed to get google calendar. HttpError: {err}")
    except Exception as e:
        raise ToolError(f"Failed to get google calendar. Exception: {e}")


@mcp.tool(
    exclude_args=["cred_token"], # the access token, is excluded to LLM
)
def create_calendar(summary: Annotated[str, Field(description="calendar title to create")],
                    time_zone: Annotated[str | None, Field(description="calendar timezone to create")] = None,
                    cred_token: str = None) -> dict:
    """Creates a new google calendar."""
    if summary == "":
        raise ValueError(f"argument `summary` can't be empty")
    service = get_client(cred_token)
    if time_zone is None:
        time_zone = get_user_timezone(service)
    elif not is_valid_iana_timezone(time_zone):
        raise ValueError(
            f"Invalid time_zone: {time_zone}. It must be a valid IANA timezone string."
        )

    calendar_body = {"summary": summary, "timeZone": time_zone}
    try:
        created_calendar = service.calendars().insert(body=calendar_body).execute()
        return created_calendar
    except HttpError as err:
        raise ToolError(f"Failed to create google calendar. HttpError: {err}")
    except Exception as e:
        raise ToolError(f"Failed to create google calendar. Exception: {e}")



@mcp.tool(
    exclude_args=["cred_token"], # the access token, is excluded to LLM
)
def update_calendar(calendar_id: Annotated[str, Field(description="calendar id to update")],
                    summary: Annotated[str | None, Field(description="calendar title to update")] = None,
                    time_zone: Annotated[str | None, Field(description="calendar timezone to update")] = None,
                    description: Annotated[str | None, Field(description="calendar description to update")] = None,
                    location: Annotated[str | None, Field(description="Geographic location of the calendar as free-form text to update")] = None,
                    cred_token: str = None) -> dict:
    """Updates an existing google calendar"""
    if calendar_id == "":
        raise ValueError("argument `calendar_id` can't be empty")
    service = get_client(cred_token)
    try:
        calendar = service.calendars().get(calendarId=calendar_id).execute()
        if summary:
            calendar["summary"] = summary
        if time_zone:
            calendar["timeZone"] = time_zone
        if description:
            calendar["description"] = description
        if location:
            calendar["location"] = location

        updated_calendar = (
            service.calendars().update(calendarId=calendar_id, body=calendar).execute()
        )
        return updated_calendar
    except HttpError as err:
        raise ToolError(f"Failed to update google calendar. HttpError: {err}")
    except Exception as e:
        raise ToolError(f"Failed to update google calendar. Exception: {e}")


@mcp.tool(
    exclude_args=["cred_token"], # the access token, is excluded to LLM
)
def delete_calendar(calendar_id: Annotated[str, Field(description="calendar id to delete")],
                    cred_token: str = None) -> str:
    """Deletes a google calendar"""
    if calendar_id == "":
        raise ValueError("argument `calendar_id` can't be empty")
    service = get_client(cred_token)
    try:
        service.calendars().delete(calendarId=calendar_id).execute()
        return f"Google calendar {calendar_id} deleted successfully."
    except HttpError as err:
        raise ToolError(f"Failed to delete google calendar. HttpError: {err}")
    except Exception as e:
        raise ToolError(f"Failed to delete google calendar. Exception: {e}")


@mcp.tool(
    annotations={
        "readOnlyHint": True,
        "destructiveHint": False,
    },
    exclude_args=["cred_token"], # the access token, is excluded to LLM
)
def list_events(calendar_id: Annotated[str, Field(description="calendar id")],
                event_type: Annotated[Literal["birthday", "default", "focusTime", "fromGmail", "outOfOffice", "workingLocation"], Field(description="The type of event to list. Defaults to 'default'")] = "default",
                single_event: Annotated[bool, Field(description="Whether to list only single event.")] = False,
                time_min: Annotated[str | None, Field(description="Upper bound (exclusive) for an event's start time to filter by. if both time_min and time_max are None, time_min will be set to the current time. Must be an RFC3339 timestamp with mandatory time zone offset.")] = None,
                time_max: Annotated[str | None, Field(description="Lower bound (exclusive) for an event's end time to filter by. Must be an RFC3339 timestamp with mandatory time zone offset.")] = None,
                order_by: Annotated[Literal["updated"] | None, Field(description="Order by which to sort events. Valid options are: updated. If set, results will be returned in ascending order.")] = None,
                q: Annotated[str | None, Field(description="Free text search terms to find events by")] = None,
                max_results: Annotated[int, Field(description="Maximum number of events to return.", ge=1, le=500)] = 250,
                cred_token: str = None) -> list:
    """Lists events for a specific google calendar."""
    if calendar_id == "":
        raise ValueError("argument `calendar_id` can't be empty")
    service = get_client(cred_token)

    # optional parameters
    params = {}
    params["singleEvents"] = single_event
    params["eventTypes"] = event_type
    if time_min:
        if not validate_rfc3339(time_min):
            raise ValueError(
                f"Invalid time_min: {time_min}. It must be a valid RFC 3339 formatted date/time string, for example, 2011-06-03T10:00:00-07:00, 2011-06-03T10:00:00Z"
            )
        params["timeMin"] = time_min
    if time_max:
        if not validate_rfc3339(time_max):
            raise ValueError(
                f"Invalid time_max: {time_max}. It must be a valid RFC 3339 formatted date/time string, for example, 2011-06-03T10:00:00-07:00, 2011-06-03T10:00:00Z"
            )
        params["timeMax"] = time_max

    if (
        not time_min and not time_max
    ):  # if no time_min or time_max is provided, default to the current time, so it will list all upcoming events
        time_min = get_current_time_rfc3339()
        params["timeMin"] = time_min

    if order_by:
        params["orderBy"] = order_by

    q = os.getenv("Q")
    if q:
        params["q"] = q

    # Filter out None or empty values
    params = {k: v for k, v in params.items() if v not in [None, ""]}

    max_results_to_return = max_results

    try:
        page_token = None
        all_events = []
        while True:
            events_result = (
                service.events()
                .list(calendarId=calendar_id, **params, pageToken=page_token)
                .execute()
            )
            events_result_list = events_result.get("items", [])
            if len(events_result_list) >= max_results_to_return:
                all_events.extend(events_result_list[:max_results_to_return])
                break
            else:
                all_events.extend(events_result_list)
                max_results_to_return -= len(events_result_list)

            page_token = events_result.get("nextPageToken")
            if not page_token:
                break
        return all_events
    except HttpError as err:
        raise ToolError(f"Failed to list events from calendar {calendar_id}. HttpError: {err}")
    except Exception as e:
        raise ToolError(f"Failed to list events from calendar {calendar_id}. Exception: {e}")


@mcp.tool(
    annotations={
        "readOnlyHint": True,
        "destructiveHint": False,
    },
    exclude_args=["cred_token"], # the access token, is excluded to LLM
)
def get_event(calendar_id: Annotated[str, Field(description="calendar id to get event from")],
              event_id: Annotated[str, Field(description="event id to get")],
              cred_token: str = None) -> dict:
    """Gets details of a specific google event."""
    if calendar_id == "":
        raise ValueError("argument `calendar_id` can't be empty")
    if event_id == "":
        raise ValueError("argument `event_id` can't be empty")
    service = get_client(cred_token)

    try:
        event = service.events().get(calendarId=calendar_id, eventId=event_id).execute()
        return event
    except HttpError as err:
        raise ToolError(f"Failed to get event {event_id} from calendar {calendar_id}. HttpError: {err}")
    except Exception as e:
        raise ToolError(f"Failed to get event {event_id} from calendar {calendar_id}. Exception: {e}")


@mcp.tool(
    exclude_args=["cred_token"], # the access token, is excluded to LLM
)
def move_event(calendar_id: Annotated[str, Field(description="calendar id to move event from")],
              event_id: Annotated[str, Field(description="event id to move")],
              new_calendar_id: Annotated[str, Field(description="calendar id to move the event to")],
              cred_token: str = None) -> dict:
    """Moves an event to a different google calendar."""
    if calendar_id == "":
        raise ValueError("argument `calendar_id` can't be empty")
    if event_id == "":
        raise ValueError("argument `event_id` can't be empty")
    if new_calendar_id == "":
        raise ValueError("argument `new_calendar_id` can't be empty")
    service = get_client(cred_token)

    try:
        existing_event = (
            service.events().get(calendarId=calendar_id, eventId=event_id).execute()
        )
        if (
            existing_event_type := existing_event.get("eventType")
        ) not in MOVABLE_EVENT_TYPES:
            raise ValueError(
                f"Events with type '{existing_event_type}' can not be moved."
            )

        event = (
            service.events()
            .move(calendarId=calendar_id, eventId=event_id, destination=new_calendar_id)
            .execute()
        )
        return event
    except HttpError as err:
        raise ToolError(f"Failed to move event {event_id} to calendar {new_calendar_id}. HttpError: {err}")
    except Exception as e:
        raise ToolError(f"Failed to move event {event_id} to calendar {new_calendar_id}. Exception: {e}")


@mcp.tool(
    exclude_args=["cred_token"], # the access token, is excluded to LLM
)
def quick_add_event(text: Annotated[str, Field(description="The text of the event to add")],
                    calendar_id: Annotated[str, Field(description="The ID of the calendar to add event for")] = "primary",
                    cred_token: str = None) -> dict:
    """Quickly adds an event to the calendar based on a simple text string."""
    if text == "":
        raise ValueError("argument `text` can't be empty")
    service = get_client(cred_token)

    try:
        event = service.events().quickAdd(calendarId=calendar_id, text=text).execute()
        return event
    except HttpError as err:
        raise ToolError(f"Failed to quick add event to calendar {calendar_id}. HttpError: {err}")
    except Exception as e:
        raise ToolError(f"Failed to quick add event to calendar {calendar_id}. Exception: {e}")


@mcp.tool(
    exclude_args=["cred_token"], # the access token, is excluded to LLM
)
def create_event(calendar_id: Annotated[str, Field(description="Calendar id to create event in. Set to `primary` to create event in the primary calendar")],
                 summary: Annotated[str, Field(description="Event title")] = "My Event",
                 location: Annotated[str, Field(description="Geographic location of the event as free-form text.")] = "",
                 description: Annotated[str, Field(description="Event description")] = "",
                 time_zone: Annotated[str | None, Field(description="Event time zone to create. Defaults to the user's timezone. Must be a valid IANA timezone string")] = None,
                 start_date: Annotated[str | None, Field(description="Event start date in the format 'yyyy-mm-dd', only used if this is an all-day event")] = None,
                 start_datetime: Annotated[str | None, Field(description="Event start date and time to create. Must be a valid RFC 3339 formatted date/time string. A time zone offset is required unless a time zone is explicitly specified in timeZone")] = None,
                 end_date: Annotated[str | None, Field(description="Event end date in the format 'yyyy-mm-dd', only used if this is an all-day event")] = None,
                 end_datetime: Annotated[str | None, Field(description="Event end date and time to create. Must be a valid RFC 3339 formatted date/time string. A time zone offset is required unless a time zone is explicitly specified in timeZone")] = None,
                 recurrence: Annotated[list[str] | None, Field(description='To create a recurring event, provide a list of strings, where each string is an RRULE, EXRULE, RDATE, or EXDATE line as defined by the RFC5545. For example, ["RRULE:FREQ=YEARLY", "EXDATE:20250403T100000Z"]. Note that DTSTART and DTEND are not allowed in this field, because they are already specified in the start_datetime and end_datetime fields.')] = None,
                 attendees: Annotated[list[str] | None, Field(description="A list of email addresses of the attendees")] = None,
                 cred_token: str = None) -> dict:
    """Creates an event in a given google calendar."""
    if calendar_id == "":
        raise ValueError("argument `calendar_id` can't be empty")
    service = get_client(cred_token)
    if time_zone is None:
        time_zone = get_user_timezone(service)
    elif not is_valid_iana_timezone(time_zone):
        raise ValueError(
            f"Invalid time_zone: {time_zone}. It must be a valid IANA timezone string."
        )

    start = {}
    if start_date:
        if not is_valid_date(start_date):
            raise ValueError(
                f"Invalid start_date: {start_date}. It must be a valid date string in the format YYYY-MM-DD."
            )
        start["date"] = start_date
    elif start_datetime:
        if not validate_rfc3339(start_datetime):
            raise ValueError(
                f"Invalid start_datetime: {start_datetime}. It must be a valid RFC 3339 formatted date/time string, for example, 2011-06-03T10:00:00-07:00, 2011-06-03T10:00:00Z"
            )
        start["dateTime"] = start_datetime
    else:
        raise ValueError("Either start_date or start_datetime must be provided.")
    start["timeZone"] = time_zone

    end = {}
    if end_date:
        if not is_valid_date(end_date):
            raise ValueError(
                f"Invalid end_date: {end_date}. It must be a valid date string in the format YYYY-MM-DD."
            )
        end["date"] = end_date
    elif end_datetime:
        if not validate_rfc3339(end_datetime):
            raise ValueError(
                f"Invalid end_datetime: {end_datetime}. It must be a valid RFC 3339 formatted date/time string, for example, 2011-06-03T10:00:00-07:00, 2011-06-03T10:00:00Z"
            )
        end["dateTime"] = end_datetime
    else:
        raise ValueError("Either end_date or end_datetime must be provided.")
    end["timeZone"] = time_zone

    event_body = {
        "summary": summary,
        "location": location,
        "description": description,
        "start": start,
        "end": end,
        "reminders": {
            "useDefault": True,  # TODO: make this configurable
        },
    }

    if recurrence:
        event_body["recurrence"] = validate_recurrence_list(recurrence)

    if attendees:
        event_body["attendees"] = attendees

    try:
        event = (
            service.events().insert(calendarId=calendar_id, body=event_body).execute()
        )
        return event
    except HttpError as err:
        raise ToolError(f"Failed to create event in calendar {calendar_id}. HttpError: {err}")
    except Exception as e:
        raise ToolError(f"Failed to create event in calendar {calendar_id}. Exception: {e}")


@mcp.tool(
    exclude_args=["cred_token"], # the access token, is excluded to LLM
)
def update_event(calendar_id: Annotated[str, Field(description="Calendar id to update event in.")],
                 event_id: Annotated[str, Field(description="Event id to update")],
                 summary: Annotated[str | None, Field(description="Event title")] = None,
                 location: Annotated[str | None, Field(description="Geographic location of the event as free-form text.")] = None,
                 description: Annotated[str | None, Field(description="Event description")] = None,
                 time_zone: Annotated[str | None, Field(description="Event time zone to update. Defaults to the user's timezone. Must be a valid IANA timezone string")] = None,
                 start_date: Annotated[str | None, Field(description="Event date in the format 'yyyy-mm-dd', only used if this is an all-day event")] = None,
                 start_datetime: Annotated[str | None, Field(description="Event start date and time to update. Must be a valid RFC 3339 formatted date/time string. A time zone offset is required unless a time zone is explicitly specified in timeZone")] = None,
                 end_date: Annotated[str | None, Field(description="Event end date in the format 'yyyy-mm-dd', only used if this is an all-day event")] = None,
                 end_datetime: Annotated[str | None, Field(description="Event end date and time to update. Must be a valid RFC 3339 formatted date/time string. A time zone offset is required unless a time zone is explicitly specified in timeZone")] = None,
                 recurrence: Annotated[list[str] | None, Field(description='For a recurring event, provide a list of strings, where each string is an RRULE, EXRULE, RDATE, or EXDATE line as defined by the RFC5545. For example, ["RRULE:FREQ=YEARLY", "EXDATE:20250403T100000Z"]. Note that DTSTART and DTEND are not allowed in this field, because they are already specified in the start_datetime and end_datetime fields.')] = None,
                 add_attendees: Annotated[list[str] | None, Field(description="A list of email addresses of the attendees to add. This will add the new attendees to the existing attendees list")] = None,
                 replace_attendees: Annotated[list[str] | None, Field(description="A list of email addresses of the attendees to replace. This is only valid when add_attendees is empty. This will replace the existing attendees list with the new list")] = None,
                 cred_token: str = None) -> dict:
    """Updates an existing google calendar event."""
    if calendar_id == "":
        raise ValueError("argument `calendar_id` can't be empty")
    if event_id == "":
        raise ValueError("argument `event_id` can't be empty")
    service = get_client(cred_token)

    try:
        existing_event = (
            service.events().get(calendarId=calendar_id, eventId=event_id).execute()
        )
        existing_event_type = existing_event.get("eventType")
    except HttpError as err:
        raise ToolError(f"Failed to get event {event_id} from calendar {calendar_id}. HttpError: {err}")
    except Exception as e:
        raise ToolError(f"Failed to get event {event_id} from calendar {calendar_id}. Exception: {e}")

    def return_field_update_error(field: str, event_type: str):
        return f"Operation Forbidden: Updating property '{field}' for event type '{event_type}' is not allowed."

    event_body = {}
    if summary:
        if not can_update_property(existing_event_type, "summary"):
            return return_field_update_error("summary", existing_event_type)

        event_body["summary"] = summary

    if location:
        if not can_update_property(existing_event_type, "location"):
            return return_field_update_error("location", existing_event_type)
        event_body["location"] = location

    if description:
        if not can_update_property(existing_event_type, "description"):
            return return_field_update_error("description", existing_event_type)
        event_body["description"] = description

    if time_zone and not is_valid_iana_timezone(time_zone):
        raise ValueError(
            f"Invalid time_zone: {time_zone}. It must be a valid IANA timezone string."
        )

    start = {}
    if not can_update_property(existing_event_type, "start"):
        return return_field_update_error("start", existing_event_type)

    if start_date:
        if not is_valid_date(start_date):
            raise ValueError(
                f"Invalid start_date: {start_date}. It must be a valid date string in the format YYYY-MM-DD."
            )
        start["date"] = start_date
    elif start_datetime:
        if not validate_rfc3339(start_datetime):
            raise ValueError(
                f"Invalid start_datetime: {start_datetime}. It must be a valid RFC 3339 formatted date/time string, for example, 2011-06-03T10:00:00-07:00, 2011-06-03T10:00:00Z"
            )
        start["dateTime"] = start_datetime
    if time_zone:
        start["timeZone"] = time_zone
    if start != {}:
        event_body["start"] = start

    end = {}
    if not can_update_property(existing_event_type, "end"):
        return return_field_update_error("end", existing_event_type)

    if end_date:
        if not is_valid_date(end_date):
            raise ValueError(
                f"Invalid end_date: {end_date}. It must be a valid date string in the format YYYY-MM-DD."
            )
        end["date"] = end_date
    elif end_datetime:
        if not validate_rfc3339(end_datetime):
            raise ValueError(
                f"Invalid end_datetime: {end_datetime}. It must be a valid RFC 3339 formatted date/time string, for example, 2011-06-03T10:00:00-07:00, 2011-06-03T10:00:00Z"
            )
        end["dateTime"] = end_datetime
    if time_zone:
        end["timeZone"] = time_zone
    if end != {}:
        event_body["end"] = end

    if recurrence:
        if not can_update_property(existing_event_type, "recurrence"):
            return return_field_update_error("recurrence", existing_event_type)

        if validate_recurrence_list(recurrence):
            event_body["recurrence"] = recurrence

    if add_attendees or replace_attendees:
        if not can_update_property(existing_event_type, "attendees"):
            return return_field_update_error("attendees", existing_event_type)

        existing_attendees = existing_event.get("attendees", [])
        existing_attendee_map = {
            a["email"]: a for a in existing_attendees if "email" in a
        }
        final_attendees = []

        if add_attendees:
            # ADD mode takes priority if both are present
            new_emails = {
                email.strip() for email in add_attendees if email.strip()
            }
            existing_emails = set(existing_attendee_map.keys())

            final_attendees = existing_attendees.copy()  # preserve full metadata

            for email in new_emails:
                if email not in existing_emails:
                    final_attendees.append({"email": email})

        elif replace_attendees:
            new_emails = {
                email.strip() for email in replace_attendees if email.strip()
            }
            for email in new_emails:
                if email in existing_attendee_map:
                    final_attendees.append(existing_attendee_map[email])
                else:
                    final_attendees.append({"email": email})

        event_body["attendees"] = final_attendees

    try:
        existing_event_type = existing_event.get("eventType")

        if not has_calendar_write_access(service, calendar_id):
            raise PermissionError("You do not have write access to this calendar.")

        existing_event.update(event_body)

        updated_event = (
            service.events()
            .update(calendarId=calendar_id, eventId=event_id, body=existing_event)
            .execute()
        )
        return updated_event
    except HttpError as err:
        raise ToolError(f"Failed to update event {event_id} in calendar {calendar_id}. HttpError: {err}")
    except Exception as e:
        raise ToolError(f"Failed to update event {event_id} in calendar {calendar_id}. Exception: {e}")


@mcp.tool(
    exclude_args=["cred_token"], # the access token, is excluded to LLM
)
def respond_to_event(calendar_id: Annotated[str, Field(description="Calendar id to respond to event in.")],
                     event_id: Annotated[str, Field(description="Event id to respond to")],
                     response: Annotated[Literal["accepted", "declined", "tentative"], Field(description="The response to the event invitation")],
                     cred_token: str = None) -> dict:
    """Responds to a calendar event by updating the current user's attendee status."""
    if calendar_id == "":
        raise ValueError("argument `calendar_id` can't be empty")
    if event_id == "":
        raise ValueError("argument `event_id` can't be empty")
    service = get_client(cred_token)

    try:
        # Get current user's email
        user_email = get_current_user_email(service)

        event = service.events().get(calendarId=calendar_id, eventId=event_id).execute()

        # Only update the responseStatus for the current user
        updated = False
        for attendee in event.get("attendees", []):
            if attendee["email"].lower() == user_email.lower():
                attendee["responseStatus"] = response
                updated = True
                break

        if not updated:
            raise ValueError(
                f"User {user_email} is not listed as an attendee on this event."
            )

        updated_event = (
            service.events()
            .patch(
                calendarId=calendar_id,
                eventId=event_id,
                body={"attendees": event["attendees"]},
            )
            .execute()
        )

        return updated_event

    except HttpError as err:
        raise ToolError(f"Failed to respond to event {event_id} in calendar {calendar_id}. HttpError: {err}")
    except Exception as e:
        raise ToolError(f"Failed to respond to event {event_id} in calendar {calendar_id}. Exception: {e}")


@mcp.tool(
    exclude_args=["cred_token"], # the access token, is excluded to LLM
)
def delete_event(calendar_id: Annotated[str, Field(description="Calendar id to delete event from.")],
                 event_id: Annotated[str, Field(description="Event id to delete")],
                 cred_token: str = None) -> dict:
    """Deletes an event from the calendar."""
    if calendar_id == "":
        raise ValueError("argument `calendar_id` can't be empty")
    if event_id == "":
        raise ValueError("argument `event_id` can't be empty")
    service = get_client(cred_token)

    try:
        service.events().delete(calendarId=calendar_id, eventId=event_id).execute()
        return f"Event {event_id} deleted successfully."
    except HttpError as err:
        raise ToolError(f"Failed to delete event {event_id} in calendar {calendar_id}. HttpError: {err}")
    except Exception as e:
        raise ToolError(f"Failed to delete event {event_id} in calendar {calendar_id}. Exception: {e}")


@mcp.tool(
    exclude_args=["cred_token"], # the access token, is excluded to LLM
)
def list_recurring_event_instances(calendar_id: Annotated[str, Field(description="Calendar id to list recurring event instances from.")],
                                   event_id: Annotated[str, Field(description="Event id to list recurring event instances for")],
                                   time_min: Annotated[str | None, Field(description="Upper bound (exclusive) for an event's start time to filter by. Must be an RFC3339 timestamp with mandatory time zone offset.")] = None,
                                   time_max: Annotated[str | None, Field(description="Lower bound (exclusive) for an event's end time to filter by. Must be an RFC3339 timestamp with mandatory time zone offset.")] = None,
                                   max_results: Annotated[int, Field(description="Maximum number of events to return.", ge=1, le=500)] = 250,
                                   cred_token: str = None) -> list:
    """Gets all instances of a recurring event."""
    if calendar_id == "":
        raise ValueError("argument `calendar_id` can't be empty")
    if event_id == "":
        raise ValueError("argument `event_id` can't be empty")
    service = get_client(cred_token)

    params = {}
    if time_min:
        if not validate_rfc3339(time_min):
            raise ValueError(
                f"Invalid time_min: {time_min}. It must be a valid RFC 3339 formatted date/time string, for example, 2011-06-03T10:00:00-07:00, 2011-06-03T10:00:00Z"
            )
        params["timeMin"] = time_min
    if time_max:
        if not validate_rfc3339(time_max):
            raise ValueError(
                f"Invalid time_max: {time_max}. It must be a valid RFC 3339 formatted date/time string, for example, 2011-06-03T10:00:00-07:00, 2011-06-03T10:00:00Z"
            )
        params["timeMax"] = time_max

    try:
        page_token = None
        all_instances = []
        while True:
            instances = (
                service.events()
                .instances(
                    calendarId=calendar_id, eventId=event_id, pageToken=page_token
                )
                .execute()
            )
            instances_result_list = instances.get("items", [])
            if len(instances_result_list) >= max_results:
                all_instances.extend(instances_result_list[:max_results])
                break
            else:
                all_instances.extend(instances_result_list)
                max_results_to_return -= len(instances_result_list)
            page_token = instances.get("nextPageToken")
            if not page_token:
                break
        return all_instances
    except HttpError as err:
        raise ToolError(f"Failed to list recurring event instances for event {event_id} in calendar {calendar_id}. HttpError: {err}")
    except Exception as e:
        raise ToolError(f"Failed to list recurring event instances for event {event_id} in calendar {calendar_id}. Exception: {e}")


if __name__ == "__main__":
    mcp.run(
        transport="streamable-http", # fixed to streamable-http
        host="0.0.0.0",
        port=PORT,
        path=MCP_PATH,
    )