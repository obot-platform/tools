import os
from tools.helper import setup_logger, get_user_timezone
from googleapiclient.errors import HttpError

logger = setup_logger(__name__)


def list_events(service):
    """Lists events for a specific calendar."""
    calendar_id = os.getenv('CALENDAR_ID')
    if not calendar_id or calendar_id == '':
        raise ValueError("CALENDAR_ID environment variable is not set properly")
    
    try:
        events_result = service.events().list(calendarId=calendar_id).execute()
        return events_result.get('items', [])
    except HttpError as err:
        raise Exception(f"HttpError listing events from calendar {calendar_id}: {err}")
    except Exception as e:
        raise Exception(f"Exception listing events from calendar {calendar_id}: {e}")

def get_event(service, event_id):
    """Gets details of a specific event."""
    calendar_id = os.getenv('CALENDAR_ID')
    if not calendar_id or calendar_id == '':
        raise ValueError("CALENDAR_ID environment variable is not set properly")

    try:
        event = service.events().get(calendarId=calendar_id, eventId=event_id).execute()
        return event
    except HttpError as err:
        raise Exception(f"HttpError retrieving event {event_id}: {err}")
    except Exception as e:
        raise Exception(f"Exception retrieving event {event_id}: {e}")

def create_event(service, summary, start_time, end_time, description='', time_zone=None):
    """Creates an event in the calendar."""
    calendar_id = os.getenv('CALENDAR_ID')
    if not calendar_id or calendar_id == '':
        raise ValueError("CALENDAR_ID environment variable is not set properly")

    if time_zone is None:
        time_zone = get_user_timezone(service)  # Retrieve user's timezone if not provided

    event_body = {
        'summary': summary,
        'description': description,
        'start': {'dateTime': start_time, 'timeZone': time_zone},
        'end': {'dateTime': end_time, 'timeZone': time_zone}
    }

    try:
        event = service.events().insert(calendarId=calendar_id, body=event_body).execute()
        return event
    except HttpError as err:
        raise Exception(f"HttpError creating event in calendar {calendar_id}: {err}")
    except Exception as e:
        raise Exception(f"Exception creating event in calendar {calendar_id}: {e}")

def update_event(service, event_id, summary=None, start_time=None, end_time=None, description=None, time_zone=None):
    """Updates an existing event."""
    calendar_id = os.getenv('CALENDAR_ID')
    if not calendar_id or calendar_id == '':
        raise ValueError("CALENDAR_ID environment variable is not set properly")

    try:
        event = service.events().get(calendarId=calendar_id, eventId=event_id).execute()

        if summary:
            event['summary'] = summary
        if description:
            event['description'] = description
        if start_time:
            event['start']['dateTime'] = start_time
        if end_time:
            event['end']['dateTime'] = end_time
        if time_zone:
            event['start']['timeZone'] = time_zone
            event['end']['timeZone'] = time_zone

        updated_event = service.events().update(calendarId=calendar_id, eventId=event_id, body=event).execute()
        return updated_event
    except HttpError as err:
        raise Exception(f"HttpError updating event {event_id}: {err}")
    except Exception as e:
        raise Exception(f"Exception updating event {event_id}: {e}")

def delete_event(service, event_id):
    """Deletes an event from the calendar."""
    calendar_id = os.getenv('CALENDAR_ID')
    if not calendar_id or calendar_id == '':
        raise ValueError("CALENDAR_ID environment variable is not set properly")

    try:
        service.events().delete(calendarId=calendar_id, eventId=event_id).execute()
        print(f"Event {event_id} deleted successfully.")
        return True
    except HttpError as err:
        raise Exception(f"HttpError deleting event {event_id}: {err}")
    except Exception as e:
        raise Exception(f"Exception deleting event {event_id}: {e}")
