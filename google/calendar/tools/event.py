from datetime import datetime, timezone
from tools.helper import (
    setup_logger,
    get_user_timezone,
    str_to_bool,
    get_obot_user_timezone,
)
import os
from googleapiclient.errors import HttpError
from rfc3339_validator import validate_rfc3339
from zoneinfo import available_timezones, ZoneInfo
import json

logger = setup_logger(__name__)

DEFAULT_MAX_RESULTS = 250
GOOGLE_EVENT_TYPE_OPTIONS = [
    "birthday",
    "default",
    "focusTime",
    "fromGmail",
    "outOfOffice",
    "workingLocation",
]
MOVABLE_EVENT_TYPES = ["default"]
CALENDAR_EVENT_TYPE_RULES = {
    "default": {
        "fully_updatable": True,
        "updatable_properties": [
            "summary",
            "description",
            "location",
            "start",
            "end",
            "attendees",
            "recurrence",
            "reminders",
            "colorId",
            "visibility",
            "transparency",
            "status",
            "extendedProperties",
            "attachments",
            "guestsCanInviteOthers",
            "guestsCanModify",
            "guestsCanSeeOtherGuests",
            "source",
            "sequence",
            # All standard properties can be updated
        ],
        "restrictions": [],
        "notes": "Most flexible event type with virtually no restrictions on updates.",
    },
    "fromGmail": {
        "fully_updatable": False,
        "updatable_properties": [],
        "restrictions": [
            "Cannot create new fromGmail events via the API",
            "Cannot change the organizer",
            "Cannot modify core properties like summary, description, location, or times",
        ],
        "notes": "Limited to updating UI and preference properties only.",
    },
    "birthday": {
        "fully_updatable": False,
        "updatable_properties": [
            "colorId",
            "summary",
            "reminders",
            "start",
            "end",
        ],
        "restrictions": [
            "Cannot modify birthdayProperties",
            "Start/end time updates must remain all-day events spanning exactly one day",
            "Timing updates are restricted if linked to a contact",
            "Cannot change the organizer",
            "Cannot create custom birthday properties via the API",
        ],
        "notes": "Use People API for comprehensive contact birthday management.",
    },
    "focusTime": {
        "fully_updatable": False,
        "updatable_properties": [
            # Standard properties
            "summary",
            "description",
            "start",
            "end",
            "reminders",
            "colorId",
            "visibility",
            "transparency",
            # Focus time specific properties
            "focusTimeProperties",
            "focusTimeProperties.autoDeclineMode",
            "focusTimeProperties.chatStatus",
            "focusTimeProperties.declineMessage",
        ],
        "restrictions": [
            "Only available on primary calendars",
            "Only for specific Google Workspace users",
            "Cannot be created on secondary calendars",
        ],
        "notes": "Used for dedicated focus periods.",
    },
    "outOfOffice": {
        "fully_updatable": False,
        "updatable_properties": [
            # Standard properties
            "summary",
            "description",
            "start",
            "end",
            "reminders",
            "colorId",
            "visibility",
            "transparency",
            # Out of office specific properties
            "outOfOfficeProperties",
            "outOfOfficeProperties.autoDeclineMode",
            "outOfOfficeProperties.declineMessage",
        ],
        "restrictions": [
            "Only available on primary calendars",
            "Only for specific Google Workspace users",
            "Cannot be created on secondary calendars",
        ],
        "notes": "Represents time away from work.",
    },
    "workingLocation": {
        "fully_updatable": False,
        "updatable_properties": [
            # Standard properties
            "summary",
            "description",
            "start",
            "end",
            "reminders",
            "colorId",
            "visibility",
            "transparency",
            # Working location specific properties
            "workingLocationProperties",
            "workingLocationProperties.type",
            "workingLocationProperties.homeOffice",
            "workingLocationProperties.customLocation",
            "workingLocationProperties.customLocation.label",
            "workingLocationProperties.officeLocation",
            "workingLocationProperties.officeLocation.buildingId",
            "workingLocationProperties.officeLocation.floorId",
            "workingLocationProperties.officeLocation.floorSectionId",
            "workingLocationProperties.officeLocation.deskId",
            "workingLocationProperties.officeLocation.label",
        ],
        "restrictions": [
            "Only available on primary calendars",
            "Only for specific Google Workspace users",
            "Cannot be created on secondary calendars",
        ],
        "notes": "Indicates where someone is working.",
    },
}


def _can_update_property(event_type, property_name):
    """
    Check if a specific property can be updated for a given event type.

    Args:
        event_type (str): The event type ('default', 'fromGmail', etc.)
        property_name (str): The property to check

    Returns:
        bool: True if the property can be updated, False otherwise
    """
    if event_type not in CALENDAR_EVENT_TYPE_RULES:
        raise ValueError(f"Unknown event type: {event_type}")

    # Default events can update all standard properties
    if event_type == "default":
        return True

    # For other event types, check the specific list
    return (
        property_name in CALENDAR_EVENT_TYPE_RULES[event_type]["updatable_properties"]
    )


def _get_event_type_restrictions(event_type):
    """
    Get the list of restrictions for a given event type.

    Args:
        event_type (str): The event type ('default', 'fromGmail', etc.)

    Returns:
        list: List of restriction strings
    """
    if event_type not in CALENDAR_EVENT_TYPE_RULES:
        raise ValueError(f"Unknown event type: {event_type}")

    return CALENDAR_EVENT_TYPE_RULES[event_type]["restrictions"]


def _get_updatable_properties(event_type):
    if event_type not in CALENDAR_EVENT_TYPE_RULES:
        raise ValueError(f"Unknown event type: {event_type}")

    return CALENDAR_EVENT_TYPE_RULES[event_type]["updatable_properties"]


# Private helper functions
def _is_valid_date(date_string: str) -> bool:
    try:
        datetime.strptime(date_string, "%Y-%m-%d")
        return True
    except ValueError:
        return False


def _is_valid_iana_timezone(timezone: str) -> bool:
    return timezone in available_timezones()


def _is_valid_recurrence_line_syntax(line: str) -> bool:
    return any(
        line.startswith(prefix) for prefix in ("RRULE:", "EXRULE:", "RDATE", "EXDATE")
    )


def _get_current_time_rfc3339():
    try:
        timezone = ZoneInfo(get_obot_user_timezone())
    except ValueError:
        # Invalid timezone, fallback to UTC
        timezone = ZoneInfo("UTC")
    return datetime.now(timezone).isoformat()


# Public functions
def list_events(service):
    """Lists events for a specific calendar."""
    calendar_id = os.getenv("CALENDAR_ID")
    if not calendar_id:
        raise ValueError("CALENDAR_ID environment variable is not set properly")

    # optional parameters
    params = {}
    event_type = os.getenv("EVENT_TYPE")
    event_type_options = GOOGLE_EVENT_TYPE_OPTIONS
    single_event = str_to_bool(os.getenv("SINGLE_EVENT", "false"))
    params["singleEvents"] = single_event

    if event_type:
        if event_type not in event_type_options:
            raise ValueError(
                f"Invalid event type: {event_type}. Valid options are: {event_type_options}"
            )
        params["eventTypes"] = event_type
    time_min = os.getenv("TIME_MIN")
    if time_min:
        if not validate_rfc3339(time_min):
            raise ValueError(
                f"Invalid time_min: {time_min}. It must be a valid RFC 3339 formatted date/time string, for example, 2011-06-03T10:00:00-07:00, 2011-06-03T10:00:00Z"
            )
        params["timeMin"] = time_min
    time_max = os.getenv("TIME_MAX")
    if time_max:
        if not validate_rfc3339(time_max):
            raise ValueError(
                f"Invalid time_max: {time_max}. It must be a valid RFC 3339 formatted date/time string, for example, 2011-06-03T10:00:00-07:00, 2011-06-03T10:00:00Z"
            )
        params["timeMax"] = time_max

    if (
        not time_min and not time_max
    ):  # if no time_min or time_max is provided, default to the current time, so it will list all upcoming events
        time_min = _get_current_time_rfc3339()
        params["timeMin"] = time_min

    order_by = os.getenv("ORDER_BY")
    order_by_options = [
        "updated"
    ]  # TODO: add startTime, but that requires `singleEvents` to be True
    if order_by:
        if order_by not in order_by_options:
            raise ValueError(
                f"Invalid order_by: {order_by}. Valid options are: startTime, updated"
            )
        params["orderBy"] = order_by

    q = os.getenv("Q")
    if q:
        params["q"] = q

    # Filter out None or empty values
    params = {k: v for k, v in params.items() if v not in [None, ""]}

    max_results_to_return = os.getenv("MAX_RESULTS")
    if max_results_to_return:
        if not max_results_to_return.isdigit() or int(max_results_to_return) <= 0:
            raise ValueError(
                f"Invalid MAX_RESULTS: {max_results_to_return}. It must be a positive integer."
            )
        else:
            max_results_to_return = int(max_results_to_return)
    else:
        max_results_to_return = DEFAULT_MAX_RESULTS

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
        raise Exception(f"HttpError listing events from calendar {calendar_id}: {err}")
    except Exception as e:
        raise Exception(f"Exception listing events from calendar {calendar_id}: {e}")


def get_event(service):
    """Gets details of a specific event."""
    calendar_id = os.getenv("CALENDAR_ID")
    if not calendar_id:
        raise ValueError("CALENDAR_ID environment variable is not set properly")
    event_id = os.getenv("EVENT_ID")
    if not event_id:
        raise ValueError("EVENT_ID environment variable is not set properly")

    try:
        event = service.events().get(calendarId=calendar_id, eventId=event_id).execute()
        return event
    except HttpError as err:
        raise Exception(f"HttpError retrieving event {event_id}: {err}")
    except Exception as e:
        raise Exception(f"Exception retrieving event {event_id}: {e}")


def move_event(service):
    """Moves an event to a different calendar."""
    calendar_id = os.getenv("CALENDAR_ID")
    if not calendar_id:
        raise ValueError("CALENDAR_ID environment variable is not set properly")
    event_id = os.getenv("EVENT_ID")
    if not event_id:
        raise ValueError("EVENT_ID environment variable is not set properly")
    new_calendar_id = os.getenv("NEW_CALENDAR_ID")
    if not new_calendar_id:
        raise ValueError("NEW_CALENDAR_ID environment variable is not set properly")

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
        raise Exception(
            f"HttpError moving event {event_id} to calendar {new_calendar_id}: {err}"
        )
    except Exception as e:
        raise Exception(
            f"Exception moving event {event_id} to calendar {new_calendar_id}: {e}"
        )


def quick_add_event(service):
    """Quickly adds an event to the calendar based on a simple text string."""
    calendar_id = os.getenv("CALENDAR_ID")
    if not calendar_id:
        raise ValueError("CALENDAR_ID environment variable is not set properly")

    text = os.getenv("TEXT")
    if not text:
        raise ValueError("TEXT environment variable is not set properly")

    try:
        event = service.events().quickAdd(calendarId=calendar_id, text=text).execute()
        return event
    except HttpError as err:
        raise Exception(
            f"HttpError quick adding event to calendar {calendar_id}: {err}"
        )
    except Exception as e:
        raise Exception(f"Exception quick adding event to calendar {calendar_id}: {e}")


def _get_recurrence_list(recurrence: str) -> list:
    """
    Parse a string into a list of recurrence rules.

    Args:
        recurrence (str): The recurrence string to parse

    Raises:
        ValueError: If the recurrence string is not a valid JSON array of strings, where each string is an RRULE, EXRULE, RDATE, or EXDATE line as defined by the RFC5545.
        ValueError: If the recurrence string is not a valid recurrence rule syntax.

    Returns:
        list: A list of recurrence rules
    """
    try:
        recurrence_list = json.loads(recurrence)
    except json.JSONDecodeError:
        try:
            fixed_recurrence = recurrence.encode().decode(
                "unicode_escape"
            )  # try to fix the recurrence string if it's not a valid JSON array of strings, for example a bad input like "[\\\"RRULE:FREQ=YEARLY\\\"]"
            recurrence_list = json.loads(fixed_recurrence)
        except json.JSONDecodeError:
            if _is_valid_recurrence_line_syntax(
                recurrence
            ):  # even if it's not a list,  check if it's a valid recurrence rule syntax, if yes, wrap it in a list
                recurrence_list = [recurrence]
            else:  # if it's not a valid recurrence rule syntax, raise an error
                raise ValueError(
                    f"Invalid recurrence list: {recurrence}. It must be a valid JSON array of strings, where each string is an RRULE, EXRULE, RDATE, or EXDATE line as defined by the RFC5545.."
                )

    for r in recurrence_list:
        if not _is_valid_recurrence_line_syntax(r):
            raise ValueError(
                f"Invalid recurrence rule: {r}. It must be a valid RRULE, EXRULE, RDATE, or EXDATE string."
            )

    return recurrence_list


def create_event(service):
    """Creates an event in the calendar."""
    calendar_id = os.getenv("CALENDAR_ID")
    if not calendar_id:
        raise ValueError("CALENDAR_ID environment variable is not set properly")

    summary = os.getenv("SUMMARY", "My Event")
    location = os.getenv("LOCATION", "")
    description = os.getenv("DESCRIPTION", "")

    user_timezone = get_user_timezone(service)
    time_zone = os.getenv("TIME_ZONE", user_timezone)
    if not _is_valid_iana_timezone(time_zone):
        raise ValueError(
            f"Invalid time_zone: {time_zone}. It must be a valid IANA timezone string."
        )

    start = {}
    start_date = os.getenv("START_DATE")
    start_datetime = os.getenv("START_DATETIME")
    if start_date:
        if not _is_valid_date(start_date):
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
    end_date = os.getenv("END_DATE")
    end_datetime = os.getenv("END_DATETIME")
    if end_date:
        if not _is_valid_date(end_date):
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
    recurrence = os.getenv("RECURRENCE")

    if recurrence:
        event_body["recurrence"] = _get_recurrence_list(recurrence)

    attendees = os.getenv(
        "ATTENDEES"
    )  # comma separated list of email addresses FOR NOW. TODO: support other types of attendees
    if attendees:
        try:
            final_attendees = []
            attendees_list = attendees.split(",")
            for attendee in attendees_list:
                final_attendees.append({"email": attendee})
            event_body["attendees"] = final_attendees
        except Exception as e:
            raise ValueError(
                f"Invalid attendees list: {attendees}. It must be a valid comma-separated list of email addresses."
            )

    try:
        event = (
            service.events().insert(calendarId=calendar_id, body=event_body).execute()
        )
        return event
    except HttpError as err:
        raise Exception(f"HttpError creating event in calendar {calendar_id}: {err}")
    except Exception as e:
        raise Exception(f"Exception creating event in calendar {calendar_id}: {e}")


def _get_current_user_email(service) -> str:
    """
    Gets the email of the current user, by getting the user_id of the primary calendar.
    """
    user_info = service.calendars().get(calendarId="primary").execute()
    return user_info["id"]


def _has_calendar_write_access(service, calendar_id: str) -> bool:
    "Validate if the user has writer access to the calendar"
    try:
        calendar = service.calendarList().get(calendarId=calendar_id).execute()
        return calendar.get("accessRole") in ("owner", "writer")
    except HttpError as e:
        if e.resp.status == 403:
            return False
        raise Exception(
            f"HttpError retrieving calendar For validating user access to {calendar_id}: {e}"
        )


def update_event(service):
    """Updates an existing event."""
    calendar_id = os.getenv("CALENDAR_ID")
    if not calendar_id:
        raise ValueError("CALENDAR_ID environment variable is not set properly")
    event_id = os.getenv("EVENT_ID")
    if not event_id:
        raise ValueError("EVENT_ID environment variable is not set properly")

    try:
        existing_event = (
            service.events().get(calendarId=calendar_id, eventId=event_id).execute()
        )
        existing_event_type = existing_event.get("eventType")
    except HttpError as err:
        raise Exception(f"HttpError retrieving event {event_id}: {err}")
    except Exception as e:
        raise Exception(f"Exception retrieving event {event_id}: {e}")

    def return_field_update_error(field: str, event_type: str):
        return f"Error: Updating property '{field}' for event type '{event_type}' is not allowed."

    event_body = {}
    summary = os.getenv("SUMMARY")
    if summary:
        if not _can_update_property(existing_event_type, "summary"):
            return return_field_update_error("summary", existing_event_type)

        event_body["summary"] = summary
    location = os.getenv("LOCATION")
    if location:
        if not _can_update_property(existing_event_type, "location"):
            return return_field_update_error("location", existing_event_type)
        event_body["location"] = location
    description = os.getenv("DESCRIPTION")
    if description:
        if not _can_update_property(existing_event_type, "description"):
            return return_field_update_error("description", existing_event_type)
        event_body["description"] = description

    time_zone = os.getenv("TIME_ZONE")
    if time_zone and not _is_valid_iana_timezone(time_zone):
        raise ValueError(
            f"Invalid time_zone: {time_zone}. It must be a valid IANA timezone string."
        )

    start = {}
    if not _can_update_property(existing_event_type, "start"):
        return return_field_update_error("start", existing_event_type)

    start_date = os.getenv("START_DATE")
    start_datetime = os.getenv("START_DATETIME")
    if start_date:
        if not _is_valid_date(start_date):
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
    if not _can_update_property(existing_event_type, "end"):
        return return_field_update_error("end", existing_event_type)

    end_date = os.getenv("END_DATE")
    end_datetime = os.getenv("END_DATETIME")
    if end_date:
        if not _is_valid_date(end_date):
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

    recurrence = os.getenv("RECURRENCE")
    if recurrence:
        if not _can_update_property(existing_event_type, "recurrence"):
            return return_field_update_error("recurrence", existing_event_type)

        event_body["recurrence"] = _get_recurrence_list(recurrence)

    add_attendees = os.getenv("ADD_ATTENDEES")
    replace_attendees = os.getenv("REPLACE_ATTENDEES")

    if add_attendees or replace_attendees:
        if not _can_update_property(existing_event_type, "attendees"):
            return return_field_update_error("attendees", existing_event_type)

        existing_attendees = existing_event.get("attendees", [])
        existing_attendee_map = {
            a["email"]: a for a in existing_attendees if "email" in a
        }
        final_attendees = []

        if add_attendees:
            # ADD mode takes priority if both are present
            new_emails = {
                email.strip() for email in add_attendees.split(",") if email.strip()
            }
            existing_emails = set(existing_attendee_map.keys())

            final_attendees = existing_attendees.copy()  # preserve full metadata

            for email in new_emails:
                if email not in existing_emails:
                    final_attendees.append({"email": email})

        elif replace_attendees:
            new_emails = {
                email.strip() for email in replace_attendees.split(",") if email.strip()
            }
            for email in new_emails:
                if email in existing_attendee_map:
                    final_attendees.append(existing_attendee_map[email])
                else:
                    final_attendees.append({"email": email})

        event_body["attendees"] = final_attendees

    try:
        existing_event_type = existing_event.get("eventType")

        if not _has_calendar_write_access(service, calendar_id):
            raise PermissionError("You do not have write access to this calendar.")

        existing_event.update(event_body)

        updated_event = (
            service.events()
            .update(calendarId=calendar_id, eventId=event_id, body=existing_event)
            .execute()
        )
        return updated_event
    except HttpError as err:
        raise Exception(f"HttpError updating event {event_id}: {err}")
    except Exception as e:
        raise Exception(f"Exception updating event {event_id}: {e}")


def respond_to_event(service):
    """
    Responds to a calendar event by updating the current user's attendee status.

    Args:
        service: An authenticated Google Calendar API service instance.
        calendar_id (str): The ID of the calendar containing the event.
        event_id (str): The ID of the event to respond to.
        response (str): One of 'accepted', 'declined', or 'tentative'.

    Returns:
        dict: The updated event object.
    """
    calendar_id = os.getenv("CALENDAR_ID")
    if not calendar_id:
        raise ValueError("CALENDAR_ID environment variable is not set properly")
    event_id = os.getenv("EVENT_ID")
    if not event_id:
        raise ValueError("EVENT_ID environment variable is not set properly")
    response = os.getenv("RESPONSE")
    if not response:
        raise ValueError("RESPONSE environment variable is not set properly")

    response_options = ["accepted", "declined", "tentative"]
    if response not in response_options:
        raise ValueError(f"Invalid response. Must be one of: {response_options}")

    try:
        # Get current user's email
        user_email = _get_current_user_email(service)

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
        raise Exception(f"HttpError responding to event: {err}")
    except Exception as e:
        raise Exception(f"Exception responding to event: {e}")


def delete_event(service):
    """Deletes an event from the calendar."""
    calendar_id = os.getenv("CALENDAR_ID")
    if not calendar_id:
        raise ValueError("CALENDAR_ID environment variable is not set properly")
    event_id = os.getenv("EVENT_ID")
    if not event_id:
        raise ValueError("EVENT_ID environment variable is not set properly")

    try:
        service.events().delete(calendarId=calendar_id, eventId=event_id).execute()
        print(f"Event {event_id} deleted successfully.")
    except HttpError as err:
        raise Exception(f"HttpError deleting event {event_id}: {err}")
    except Exception as e:
        raise Exception(f"Exception deleting event {event_id}: {e}")


def recurring_event_instances(service):
    """Gets all instances of a recurring event."""
    calendar_id = os.getenv("CALENDAR_ID")
    if not calendar_id:
        raise ValueError("CALENDAR_ID environment variable is not set properly")
    event_id = os.getenv("EVENT_ID")
    if not event_id:
        raise ValueError("EVENT_ID environment variable is not set properly")

    params = {}
    time_min = os.getenv("TIME_MIN")
    if time_min:
        if not validate_rfc3339(time_min):
            raise ValueError(
                f"Invalid time_min: {time_min}. It must be a valid RFC 3339 formatted date/time string, for example, 2011-06-03T10:00:00-07:00, 2011-06-03T10:00:00Z"
            )
        params["timeMin"] = time_min
    time_max = os.getenv("TIME_MAX")
    if time_max:
        if not validate_rfc3339(time_max):
            raise ValueError(
                f"Invalid time_max: {time_max}. It must be a valid RFC 3339 formatted date/time string, for example, 2011-06-03T10:00:00-07:00, 2011-06-03T10:00:00Z"
            )
        params["timeMax"] = time_max

    max_results_to_return = os.getenv("MAX_RESULTS")
    if max_results_to_return:
        if not max_results_to_return.isdigit() or int(max_results_to_return) <= 0:
            raise ValueError(
                f"Invalid MAX_RESULTS: {max_results_to_return}. It must be a positive integer."
            )
        else:
            max_results_to_return = int(max_results_to_return)
    else:
        max_results_to_return = DEFAULT_MAX_RESULTS

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
            if len(instances_result_list) >= max_results_to_return:
                all_instances.extend(instances_result_list[:max_results_to_return])
                break
            else:
                all_instances.extend(instances_result_list)
                max_results_to_return -= len(instances_result_list)
            page_token = instances.get("nextPageToken")
            if not page_token:
                break
        return all_instances
    except HttpError as err:
        raise Exception(f"HttpError retrieving instances of event {event_id}: {err}")
    except Exception as e:
        raise Exception(f"Exception retrieving instances of event {event_id}: {e}")
