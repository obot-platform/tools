import os
from tools.helper import setup_logger, get_user_timezone
from googleapiclient.errors import HttpError

logger = setup_logger(__name__)


def list_calendars(service):
    """Lists all calendars for the authenticated user."""
    try:
        calendars = service.calendarList().list().execute()
        return calendars.get("items", [])
    except HttpError as err:
        logger.error(f"HttpError listing calendars: {err}")
        return []


def get_calendar(service):
    """Gets details of a specific calendar."""
    calendar_id = os.getenv("CALENDAR_ID")
    if not calendar_id or calendar_id == "":
        raise ValueError("CALENDAR_ID environment variable is not set properly")
    try:
        calendar = service.calendars().get(calendarId=calendar_id).execute()
        return calendar
    except HttpError as err:
        raise Exception(f"HttpError retrieving calendar {calendar_id}: {err}")
    except Exception as e:
        raise Exception(f"Exception retrieving calendar {calendar_id}: {e}")


def create_calendar(service):
    """Creates a new calendar."""
    summary = os.getenv("SUMMARY")
    if not summary or summary == "":
        raise ValueError(f"SUMMARY environment variable is not set properly: {summary}")
    time_zone = os.getenv("TIME_ZONE", get_user_timezone(service))
    calendar_body = {"summary": summary, "timeZone": time_zone}
    try:
        created_calendar = service.calendars().insert(body=calendar_body).execute()
        return created_calendar
    except HttpError as err:
        raise Exception(f"HttpError creating calendar: {err}")
    except Exception as e:
        raise Exception(f"Exception creating calendar: {e}")


def update_calendar(service):
    """Updates an existing calendar."""
    calendar_id = os.getenv("CALENDAR_ID")
    if not calendar_id or calendar_id == "":
        raise ValueError("CALENDAR_ID environment variable is not set properly")
    try:
        calendar = service.calendars().get(calendarId=calendar_id).execute()
        if os.environ.get("SUMMARY"):
            calendar["summary"] = os.getenv("SUMMARY")
        if os.environ.get("TIME_ZONE"):
            calendar["timeZone"] = os.getenv("TIME_ZONE")
        if os.environ.get("DESCRIPTION"):
            calendar["description"] = os.getenv("DESCRIPTION")
        if os.environ.get("LOCATION"):
            calendar["location"] = os.getenv("LOCATION")

        updated_calendar = (
            service.calendars().update(calendarId=calendar_id, body=calendar).execute()
        )
        return updated_calendar
    except HttpError as err:
        raise Exception(f"HttpError updating calendar {calendar_id}: {err}")
    except Exception as e:
        raise Exception(f"Exception updating calendar {calendar_id}: {e}")


def delete_calendar(service):
    """Deletes a calendar."""
    calendar_id = os.getenv("CALENDAR_ID")
    if not calendar_id or calendar_id == "":
        raise ValueError("CALENDAR_ID environment variable is not set properly")
    try:
        service.calendars().delete(calendarId=calendar_id).execute()
        return f"Calendar {calendar_id} deleted successfully."
    except HttpError as err:
        raise Exception(f"HttpError deleting calendar {calendar_id}: {err}")
    except Exception as e:
        raise Exception(f"Exception deleting calendar {calendar_id}: {e}")
