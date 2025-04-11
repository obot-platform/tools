from tools.helper import (
    ZOOM_API_URL,
    ACCESS_TOKEN,
    str_to_bool,
    tool_registry,
    setup_logger,
)
from tools.users import get_user_type
import requests
import os
import re
import string
import random
from datetime import datetime
from zoneinfo import ZoneInfo
import json

logger = setup_logger(__name__)


# ------------------------------------------------------------------------------------------------
# Helper functions
# ------------------------------------------------------------------------------------------------
def _convert_utc_to_local_time(utc_time_str: str, timezone: str) -> str:
    try:
        # Parse ISO 8601 string into a naive datetime object
        utc_time = datetime.strptime(utc_time_str, "%Y-%m-%dT%H:%M:%SZ")

        # Assign UTC timezone using ZoneInfo
        utc_time = utc_time.replace(tzinfo=ZoneInfo("UTC"))

        # Convert to specified local timezone
        local_time = utc_time.astimezone(ZoneInfo(timezone))

        # Format the local time as needed
        output_gmt_format = local_time.strftime(
            "%Y-%m-%dT%H:%M:%S"
        )  # Customize format as necessary
        return output_gmt_format
    except Exception as e:
        logger.error(f"Error converting time: {e}")
        raise ValueError(f"Error converting time: {e}")


def _validate_meeting_start_time(input_time: str) -> bool:
    """
    Validates the input time format for a meeting's start time.

    - GMT Format: yyyy-MM-ddTHH:mm:ssZ (e.g., 2020-03-31T12:02:00Z)
    - Local Timezone Format: yyyy-MM-ddTHH:mm:ss (e.g., 2020-03-31T12:02:00)

    Args:
        input_time (str): The input string to validate.

    Returns:
        bool: True if the input matches one of the valid formats, False otherwise.
    """
    # Regular expression for GMT format (e.g., 2020-03-31T12:02:00Z)
    gmt_format = r"^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$"

    # Regular expression for Local Timezone format (e.g., 2020-03-31T12:02:00)
    local_format = r"^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}$"

    # Validate against both formats
    res = bool(re.match(gmt_format, input_time) or re.match(local_format, input_time))
    if res == False:  # invalid time format
        logger.error(f"Invalid input time format: {input_time}")
    return res


def _validate_invitees(invitees: list) -> bool:
    """
    Validates a list of meeting invitees to ensure all are valid email addresses.

    Args:
        invitees (list): A list of strings to validate as email addresses.

    Returns:
        bool: True if all strings in the list are valid emails, False otherwise.
    """
    # Regular expression for validating email addresses
    email_regex = r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$"

    # Validate each email in the list
    return all(re.match(email_regex, email) for email in invitees)


def _generate_password():
    """
    Generates a random 8-character password containing letters and digits.

    Returns:
        str: An 8-character random password.
    """
    characters = (
        string.ascii_letters + string.digits
    )  # Include uppercase, lowercase, and digits
    password = "".join(random.choices(characters, k=8))  # Generate 8 random characters
    return password


def _trim_meeting_id(meeting_id_or_uuid: str) -> str:
    """
    Trims the meeting ID or UUID to remove any whitespace
    """
    return "".join(meeting_id_or_uuid.split())


def _remove_meeting_series_uuid(response_json: dict) -> dict:
    """
    Remove the meeting series uuid from the response json if APIs:
    - Get Meeting
    - List Meetings
    - List Upcoming Meetings
    - Create Meeting
    - Update Meeting
    """

    if "uuid" in response_json:
        response_json.pop("uuid", None)
    elif "meetings" in response_json:
        for meeting in response_json["meetings"]:
            if "uuid" in meeting:
                meeting.pop("uuid", None)

    return response_json


def _is_meeting_id(value: str) -> bool:
    return value.isdigit() and 9 <= len(value) <= 11


_meeting_types = {
    1: "An instant meeting",
    2: "A scheduled meeting",
    3: "A recurring meeting with no fixed time",
    8: "A recurring meeting with fixed time",
    10: "A screen share only meeting",
}


# ------------------------------------------------------------------------------------------------
# Tool functions
# ------------------------------------------------------------------------------------------------
@tool_registry.decorator("CreateMeeting")
def create_meeting():
    url = f"{ZOOM_API_URL}/users/me/meetings"
    user_type = get_user_type()
    meeting_invitees = os.getenv(
        "MEETING_INVITEES", ""
    )  # a list of emails separated by commas
    if meeting_invitees != "" and not _validate_invitees(meeting_invitees.split(",")):
        logger.error(
            f"Invalid invitees: {meeting_invitees}. Must be a list of valid email addresses separated by commas."
        )
        raise ValueError(
            f"Invalid invitees: {meeting_invitees}. Must be a list of valid email addresses separated by commas."
        )
    agenda = os.getenv("AGENDA", "My Meeting")
    default_password = str_to_bool(os.getenv("DEFAULT_PASSWORD", "false"))
    duration = int(os.getenv("DURATION", 60))
    password = os.getenv("PASSWORD", "")
    if password == "":
        password = _generate_password()
    pre_schedule = str_to_bool(os.getenv("PRE_SCHEDULE", "false"))
    # schedule_for = os.environ["SCHEDULE_FOR"] # only for account level app
    audio_recording = os.getenv("AUDIO_RECORDING", "none")
    contact_email = os.getenv("CONTACT_EMAIL", "")
    contact_name = os.getenv("CONTACT_NAME", "")
    private_meeting = str_to_bool(os.getenv("PRIVATE_MEETING", "false"))
    start_time = os.getenv("START_TIME", "")
    if start_time != "" and not _validate_meeting_start_time(start_time):
        raise ValueError(
            f"Invalid start time format: {start_time}. Must be in GMT or local timezone format."
        )

    meeting_template_id = os.getenv("MEETING_TEMPLATE_ID", "")
    timezone = os.getenv("TIMEZONE", "")
    topic = os.getenv("TOPIC", "")

    meeting_type = int(os.getenv("MEETING_TYPE", 2))
    recurrence = os.getenv("RECURRENCE", "")
    auto_start_meeting_summary = str_to_bool(
        os.getenv("AUTO_START_MEETING_SUMMARY", "false")
    )

    if meeting_type not in _meeting_types:
        raise ValueError(
            f"Invalid meeting type: {meeting_type}. Must be one of: {_meeting_types.keys()}"
        )
    payload = {
        "agenda": agenda,
        "default_password": default_password,
        "duration": duration,
        "password": password,
        "pre_schedule": pre_schedule,
        "settings": {
            "allow_multiple_devices": True,
            "approval_type": 2,
            "audio": "both",
            "auto_recording": audio_recording,
            "calendar_type": 1,
            "close_registration": False,
            "cn_meeting": False,
            "contact_email": contact_email,
            "contact_name": contact_name,
            "email_notification": True,
            "encryption_type": "enhanced_encryption",
            "focus_mode": True,
            "host_video": True,
            "in_meeting": False,
            "jbh_time": 0,
            "join_before_host": True,
            "question_and_answer": {
                "enable": True,
                "allow_submit_questions": True,
                "allow_anonymous_questions": True,
                "question_visibility": "all",
                "attendees_can_comment": True,
                "attendees_can_upvote": True,
            },
            "meeting_authentication": False,
            "meeting_invitees": [
                {"email": invitee} for invitee in meeting_invitees.split(",")
            ],
            "mute_upon_entry": True,
            "participant_video": False,
            "private_meeting": private_meeting,
            "registrants_confirmation_email": True,
            "registrants_email_notification": True,
            "registration_type": 1,
            "show_share_button": True,
            "use_pmi": False,
            "waiting_room": False,
            "watermark": False,
            "host_save_video_order": True,
            "alternative_host_update_polls": True,
            "internal_meeting": False,
            "continuous_meeting_chat": {
                "enable": True,
                "auto_add_invited_external_users": True,
                "auto_add_meeting_participants": True,
            },
            "participant_focused_meeting": False,
            "push_change_to_calendar": False,
            "auto_start_meeting_summary": auto_start_meeting_summary,
            "auto_start_ai_companion_questions": False,
            "device_testing": False,
        },
        "start_time": start_time,
        "template_id": meeting_template_id,
        "timezone": timezone,
        "topic": topic,
        "type": meeting_type,
    }

    if recurrence != "":
        try:
            recurrence_object = json.loads(recurrence)
            payload["recurrence"] = recurrence_object
        except Exception as e:
            logger.error(
                f"Exception {e}: Invalid recurrence: {recurrence}. Must be a valid JSON object."
            )
            raise ValueError(
                f"Exception {e}: Invalid recurrence: {recurrence}. Must be a valid JSON object."
            )
    # features for licensed users
    if user_type == 2:
        payload["settings"]["global_dial_in_countries"] = ["US"]

    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {ACCESS_TOKEN}",
    }

    response = requests.post(url, json=payload, headers=headers)
    if response.status_code != 201:
        return {"message": f"Error creating meeting: {response.text}"}

    cleaned_json = _remove_meeting_series_uuid(response.json())
    return cleaned_json


def _get_meeting():
    meeting_id = _trim_meeting_id(os.environ["MEETING_ID"])
    url = f"{ZOOM_API_URL}/meetings/{meeting_id}"
    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
    }
    response = requests.get(url, headers=headers)
    if response.status_code != 200:
        raise ValueError(f"Error getting meeting: {response.text}")

    return response.json()


@tool_registry.decorator("GetMeeting")
def get_meeting():
    res_json = _get_meeting()
    if (
        "start_time" in res_json
        and "timezone" in res_json
        and res_json["timezone"] != "UTC"
        and res_json["timezone"] != ""
    ):
        res_json["start_time_local"] = _convert_utc_to_local_time(
            res_json["start_time"], res_json["timezone"]
        )
        res_json["start_time_utc"] = res_json.pop("start_time")

    return _remove_meeting_series_uuid(res_json)


@tool_registry.decorator("DeleteMeeting")
def delete_meeting():
    meeting_id = _trim_meeting_id(os.environ["MEETING_ID"])
    url = f"{ZOOM_API_URL}/meetings/{meeting_id}"
    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
    }
    response = requests.delete(url, headers=headers)
    if response.status_code != 204:
        return {"message": f"Error deleting meeting: {response.text}"}

    return {"message": f"successfully deleted meeting, ID: {meeting_id}"}


@tool_registry.decorator("ListMeetings")
def list_meetings():
    url = f"{ZOOM_API_URL}/users/me/meetings"
    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
    }
    params = {}
    type = os.getenv("TYPE", "")
    type_enums = [
        "scheduled",
        "live",
        "upcoming",
        "upcoming_meetings",
        "previous_meetings",
    ]
    if type != "":
        if type not in type_enums:
            raise ValueError(f"Invalid type: {type}. Must be one of: {type_enums}")
        params["type"] = type

    response = requests.get(url, headers=headers, params=params)
    if response.status_code != 200:
        return {
            "message": f"{response.status_code} Error listing meetings: {response.text}"
        }

    res_json = response.json()
    for meeting in res_json["meetings"]:
        if (
            "start_time" in meeting
            and "timezone" in meeting
            and meeting["timezone"] != "UTC"
            and meeting["timezone"] != ""
        ):
            meeting["start_time_local"] = _convert_utc_to_local_time(
                meeting["start_time"], meeting["timezone"]
            )
            meeting["start_time_utc"] = meeting.pop("start_time")


    return _remove_meeting_series_uuid(res_json)


@tool_registry.decorator("ListUpcomingMeetings")
def list_upcoming_meetings():
    url = f"{ZOOM_API_URL}/users/me/upcoming_meetings"
    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
    }
    response = requests.get(url, headers=headers)
    if response.status_code != 200:
        return {"message": f"Error listing upcoming meetings: {response.text}"}

    res_json = response.json()
    for meeting in res_json["meetings"]:
        if (
            "start_time" in meeting
            and "timezone" in meeting
            and meeting["timezone"] != "UTC"
            and meeting["timezone"] != ""
        ):
            meeting["start_time_local"] = _convert_utc_to_local_time(
                meeting["start_time"], meeting["timezone"]
            )
            meeting["start_time_utc"] = meeting.pop("start_time")
    return _remove_meeting_series_uuid(res_json)


@tool_registry.decorator("GetMeetingInvitation")
def get_meeting_invitation():
    meeting_id = _trim_meeting_id(os.environ["MEETING_ID"])
    url = f"{ZOOM_API_URL}/meetings/{meeting_id}/invitation"
    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
    }
    response = requests.get(url, headers=headers)
    if response.status_code != 200:
        return {"message": f"Error getting meeting invitation: {response.text}"}
    return response.json()


@tool_registry.decorator("UpdateMeeting")
def update_meeting():
    meeting_id = _trim_meeting_id(os.environ["MEETING_ID"])
    url = f"{ZOOM_API_URL}/meetings/{meeting_id}"

    payload = {}

    if "AGENDA" in os.environ:
        payload["agenda"] = os.environ["AGENDA"]
    if "DEFAULT_PASSWORD" in os.environ:
        payload["default_password"] = str_to_bool(os.environ["DEFAULT_PASSWORD"])
    if "DURATION" in os.environ:
        payload["duration"] = int(os.environ["DURATION"])
    if "PASSWORD" in os.environ:
        password = os.environ["PASSWORD"]
        if password == "":
            password = _generate_password()
        payload["password"] = password

    if "RECURRENCE" in os.environ:
        recurrence = os.environ["RECURRENCE"]
        try:
            recurrence_object = json.loads(recurrence)
        except Exception as e:
            raise ValueError(
                f"Invalid recurrence: {recurrence}. Must be a valid JSON object."
            )
        payload["recurrence"] = recurrence_object

    if "PRE_SCHEDULE" in os.environ:
        payload["pre_schedule"] = str_to_bool(os.environ["PRE_SCHEDULE"])

    if "START_TIME" in os.environ:
        start_time = os.environ["START_TIME"]
        if start_time != "" and not _validate_meeting_start_time(start_time):
            raise ValueError(
                f"Invalid start time format: {start_time}. Must be in GMT or local timezone format."
            )
        payload["start_time"] = start_time

    if "MEETING_TEMPLATE_ID" in os.environ:
        payload["template_id"] = os.environ["MEETING_TEMPLATE_ID"]
    if "TIMEZONE" in os.environ:
        payload["timezone"] = os.environ["TIMEZONE"]
    if "TOPIC" in os.environ:
        payload["topic"] = os.environ["TOPIC"]

    if "MEETING_TYPE" in os.environ:
        meeting_type = int(os.environ["MEETING_TYPE"])
        if meeting_type not in _meeting_types:
            raise ValueError(
                f"Invalid meeting type: {meeting_type}. Must be one of: {_meeting_types.keys()}"
            )
        payload["type"] = meeting_type

    # args of settings
    settings = {}
    if "AUDIO_RECORDING" in os.environ:
        settings["audio_recording"] = os.environ["AUDIO_RECORDING"]
    if "CONTACT_EMAIL" in os.environ:
        settings["contact_email"] = os.environ["CONTACT_EMAIL"]
    if "CONTACT_NAME" in os.environ:
        settings["contact_name"] = os.environ["CONTACT_NAME"]
    if "PRIVATE_MEETING" in os.environ:
        settings["private_meeting"] = str_to_bool(os.environ["PRIVATE_MEETING"])

    if "MEETING_INVITEES" in os.environ:
        meeting_invitees = os.environ[
            "MEETING_INVITEES"
        ]  # a list of emails separated by commas
        if meeting_invitees != "" and not _validate_invitees(
            meeting_invitees.split(",")
        ):
            raise ValueError(
                f"Invalid invitees: {meeting_invitees}. Must be a list of valid email addresses separated by commas."
            )
        meeting_invitees_list = [
            {"email": invitee} for invitee in meeting_invitees.split(",")
        ]
        settings["meeting_invitees"] = meeting_invitees_list

    if "AUTO_START_MEETING_SUMMARY" in os.environ:
        settings["auto_start_meeting_summary"] = str_to_bool(
            os.environ["AUTO_START_MEETING_SUMMARY"]
        )

    if settings:
        payload["settings"] = settings

    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {ACCESS_TOKEN}",
    }

    response = requests.patch(url, json=payload, headers=headers)
    if response.status_code != 204:
        return {"message": f"Error updating meeting: {response.text}"}
    return {"message": "successfully updated meeting"}


@tool_registry.decorator("ListMeetingTemplates")
def list_meeting_templates():
    url = f"{ZOOM_API_URL}/users/me/meeting_templates"
    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
    }
    response = requests.get(url, headers=headers)
    if response.status_code != 200:
        return {"message": f"Error listing meeting templates: {response.text}"}
    return response.json()


@tool_registry.decorator("GetMeetingSummary")
def get_meeting_summary():
    user_type = get_user_type()
    if user_type != 2:
        raise ValueError(
            "The `Get Meeting Summary` feature is only available for licensed users."
        )
    meeting_uuid = _trim_meeting_id(os.environ["MEETING_UUID"])
    if _is_meeting_id(meeting_uuid):
        return {
            "message": "ValueError: Meeting UUID must be provided instead of meeting ID."
        }

    url = f"{ZOOM_API_URL}/meetings/{meeting_uuid}/meeting_summary"
    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
    }
    response = requests.get(url, headers=headers)
    if response.status_code != 200:
        return {
            "message": f"{response.status_code} Error getting meeting summary: {response.text}"
        }
    return response.json()


@tool_registry.decorator("GetPastMeetingDetails")
def get_past_meeting_details():
    meeting_id_or_uuid = _trim_meeting_id(os.environ["MEETING_ID_OR_UUID"])
    url = f"{ZOOM_API_URL}/past_meetings/{meeting_id_or_uuid}"
    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
    }
    response = requests.get(url, headers=headers)
    if response.status_code != 200:
        return {
            "message": f"{response.status_code} Error getting past meeting details: {response.text}"
        }
    return response.json()


@tool_registry.decorator("ListPastMeetingInstances")
def list_past_meeting_instances():
    meeting_id = _trim_meeting_id(os.environ["MEETING_ID"])
    url = f"{ZOOM_API_URL}/past_meetings/{meeting_id}/instances"
    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
    }
    response = requests.get(url, headers=headers)
    if response.status_code != 200:
        return {
            "message": f"{response.status_code} Error listing past meeting instances: {response.text}"
        }
    return response.json()
