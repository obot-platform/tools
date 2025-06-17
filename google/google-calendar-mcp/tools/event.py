from datetime import datetime, timezone
from tools.helper import (
    setup_logger,
    get_obot_user_timezone,
)
from googleapiclient.errors import HttpError
from zoneinfo import available_timezones, ZoneInfo

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


def can_update_property(event_type, property_name):
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
def is_valid_date(date_string: str) -> bool:
    try:
        datetime.strptime(date_string, "%Y-%m-%d")
        return True
    except ValueError:
        return False


def is_valid_iana_timezone(timezone: str) -> bool:
    return timezone in available_timezones()


def _is_valid_recurrence_line_syntax(line: str) -> bool:
    return any(
        line.startswith(prefix) for prefix in ("RRULE:", "EXRULE:", "RDATE", "EXDATE")
    )


def get_current_time_rfc3339():
    try:
        timezone = ZoneInfo(get_obot_user_timezone())
    except ValueError:
        # Invalid timezone, fallback to UTC
        timezone = ZoneInfo("UTC")
    return datetime.now(timezone).isoformat()


def validate_recurrence_list(recurrence_list: list[str]) -> list[str]:
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

    for r in recurrence_list:
        if not _is_valid_recurrence_line_syntax(r):
            raise ValueError(
                f"Invalid recurrence rule: {r}. It must be a valid RRULE, EXRULE, RDATE, or EXDATE string."
            )

    return recurrence_list

def get_current_user_email(service) -> str:
    """
    Gets the email of the current user, by getting the user_id of the primary calendar.
    """
    user_info = service.calendars().get(calendarId="primary").execute()
    return user_info["id"]


def has_calendar_write_access(service, calendar_id: str) -> bool:
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