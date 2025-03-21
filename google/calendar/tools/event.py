from datetime import datetime, timezone
from tools.helper import setup_logger, get_user_timezone, str_to_bool, get_obot_user_timezone
import os
from googleapiclient.errors import HttpError
from rfc3339_validator import validate_rfc3339
from zoneinfo import available_timezones, ZoneInfo
import json
from dateutil.rrule import rrulestr

logger = setup_logger(__name__)

DEFAULT_MAX_RESULTS = 250


# Private helper functions
def _is_valid_date(date_string: str) -> bool:
    try:
        datetime.strptime(date_string, "%Y-%m-%d")
        return True
    except ValueError:
        return False


def _is_valid_iana_timezone(timezone: str) -> bool:
    return timezone in available_timezones()


def _validate_rrule(rrule_str: str) -> bool:
    """Validates an RRULE string for recurrence rules"""
    try:
        rrulestr(rrule_str)  # If it doesn't raise an error, it's valid
        return True
    except ValueError:
        return False


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
    event_type_options = [
        "birthday",
        "default",
        "focusTime",
        "fromGmail",
        "outOfOffice",
        "workingLocation",
    ]
    single_event = str_to_bool(os.getenv("SINGLE_EVENT", "false"))
    params["singleEvents"] = single_event

    if event_type:
        if event_type not in event_type_options:
            raise ValueError(
                f"Invalid event type: {event_type}. Valid options are: {event_type_options}"
            )
        params["eventType"] = event_type
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
        try:
            recurrence_list = json.loads(recurrence)
        except json.JSONDecodeError:
            raise ValueError(
                f"Invalid recurrence list: {recurrence}. It must be a valid JSON array."
            )
        finally:
            for r in recurrence_list:
                if not _validate_rrule(r):
                    raise ValueError(
                        f"Invalid recurrence rule: {r}. It must be a valid RRULE string."
                    )
            event_body["recurrence"] = recurrence_list

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


def update_event(service):
    """Updates an existing event."""
    calendar_id = os.getenv("CALENDAR_ID")
    if not calendar_id:
        raise ValueError("CALENDAR_ID environment variable is not set properly")
    event_id = os.getenv("EVENT_ID")
    if not event_id:
        raise ValueError("EVENT_ID environment variable is not set properly")

    event_body = {}
    summary = os.getenv("SUMMARY")
    if summary:
        event_body["summary"] = summary
    location = os.getenv("LOCATION")
    if location:
        event_body["location"] = location
    description = os.getenv("DESCRIPTION")
    if description:
        event_body["description"] = description

    time_zone = os.getenv("TIME_ZONE")
    if time_zone and not _is_valid_iana_timezone(time_zone):
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
    if time_zone:
        start["timeZone"] = time_zone
    if start != {}:
        event_body["start"] = start

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
    if time_zone:
        end["timeZone"] = time_zone
    if end != {}:
        event_body["end"] = end

    recurrence = os.getenv("RECURRENCE")
    if recurrence:
        try:
            recurrence_list = json.loads(recurrence)
        except json.JSONDecodeError:
            raise ValueError(
                f"Invalid recurrence list: {recurrence}. It must be a valid JSON array."
            )
        finally:
            for r in recurrence_list:
                if not _validate_rrule(r):
                    raise ValueError(
                        f"Invalid recurrence rule: {r}. It must be a valid RRULE string."
                    )
            event_body["recurrence"] = recurrence_list

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
        event = service.events().get(calendarId=calendar_id, eventId=event_id).execute()

        updated_event = (
            service.events()
            .update(calendarId=calendar_id, eventId=event_id, body=event)
            .execute()
        )
        return updated_event
    except HttpError as err:
        raise Exception(f"HttpError updating event {event_id}: {err}")
    except Exception as e:
        raise Exception(f"Exception updating event {event_id}: {e}")


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
