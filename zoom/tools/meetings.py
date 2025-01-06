from tools.helper import ZOOM_API_URL, ACCESS_TOKEN, str_to_bool
import requests
import os
import re
import string
import random


def validate_meeting_start_time(input_time: str) -> bool:
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
    return bool(re.match(gmt_format, input_time) or re.match(local_format, input_time))


def validate_invitees(invitees: list) -> bool:
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


def generate_password():
    """
    Generates a random 8-character password containing letters and digits.
    
    Returns:
        str: An 8-character random password.
    """
    characters = string.ascii_letters + string.digits  # Include uppercase, lowercase, and digits
    password = ''.join(random.choices(characters, k=8))  # Generate 8 random characters
    return password


    
meeting_types = {
    1: "An instant meeting",
    2: "A scheduled meeting",
    3: "A recurring meeting with no fixed time",
    8: "A recurring meeting with fixed time",
    10: "A screen share only meeting"
}

def create_meeting():
    url = f"{ZOOM_API_URL}/users/me/meetings"
    meeting_invitees = os.getenv("MEETING_INVITEES", "") # a list of emails separated by commas
    if meeting_invitees != "" and not validate_invitees(meeting_invitees.split(",")):
        raise ValueError(f"Invalid invitees: {meeting_invitees}. Must be a list of valid email addresses separated by commas.")
    agenda = os.getenv("AGENDA", "My Meeting")
    default_password = str_to_bool(os.getenv("DEFAULT_PASSWORD", "false"))
    duration = int(os.getenv("DURATION", 60))
    password = os.getenv("PASSWORD", "")
    if password == "":
        password = generate_password()
    pre_schedule = str_to_bool(os.getenv("PRE_SCHEDULE", "false"))
    schedule_for = os.environ["SCHEDULE_FOR"]
    audio_recording = os.getenv("AUDIO_RECORDING", "none")
    contact_email = os.getenv("CONTACT_EMAIL", "")
    contact_name = os.getenv("CONTACT_NAME", "")
    private_meeting = str_to_bool(os.getenv("PRIVATE_MEETING", "false"))
    start_time = os.getenv("START_TIME", "")
    if start_time != "" and not validate_meeting_start_time(start_time):
        raise ValueError(f"Invalid start time format: {start_time}. Must be in GMT or local timezone format.")
    meeting_template_id = os.getenv("MEETING_TEMPLATE_ID", "")
    timezone = os.getenv("TIMEZONE", "")
    topic = os.getenv("TOPIC", "")

    meeting_type = int(os.getenv("MEETING_TYPE", 2))
    if meeting_type not in meeting_types:
        raise ValueError(f"Invalid meeting type: {meeting_type}. Must be one of: {meeting_types.keys()}")
    # TODO: support recurrence and more settings in the future
    payload = {
        "agenda": agenda,
        "default_password": default_password,
        "duration": duration,
        "password": password,
        "pre_schedule": pre_schedule,
        "schedule_for": schedule_for,
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
            "global_dial_in_countries": ["US"],
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
                "attendees_can_upvote": True
            },
            "meeting_authentication": False,
            "meeting_invitees": [{"email": invitee} for invitee in meeting_invitees.split(",")],
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
                "auto_add_meeting_participants": True
            },
            "participant_focused_meeting": False,
            "push_change_to_calendar": False,
            "auto_start_meeting_summary": False,
            "auto_start_ai_companion_questions": False,
            "device_testing": False
        },
        "start_time": start_time,
        "template_id": meeting_template_id,
        "timezone": timezone,
        "topic": topic,
        "type": meeting_type
    }
    headers = {
        "Content-Type": "application/json",
        "Authorization": "Bearer YOUR_SECRET_TOKEN"
    }

    response = requests.post(url, json=payload, headers=headers)
    if response.status_code != 201:
        raise Exception(f"Error creating meeting: {response.json()}")   
    return response.json()


def get_meeting():
    meeting_id = os.environ["MEETING_ID"]
    url = f"{ZOOM_API_URL}/meetings/{meeting_id}"
    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
    }
    response = requests.get(url, headers=headers)
    if response.status_code != 200:
        raise Exception(f"Error getting meeting: {response.json()}")
    return response.json()


def delete_meeting():
    meeting_id = os.environ["MEETING_ID"]
    url = f"{ZOOM_API_URL}/meetings/{meeting_id}"
    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
    }
    response = requests.delete(url, headers=headers)
    if response.status_code != 204:
        raise Exception(f"Error deleting meeting: {response.json()}")
    return response.json()


def list_meetings():
    url = f"{ZOOM_API_URL}/users/me/meetings"
    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
    }
    response = requests.get(url, headers=headers)
    if response.status_code != 200:
        raise Exception(f"Error listing meetings: {response.json()}")
    return response.json()


def update_meeting():
    meeting_id = os.environ["MEETING_ID"]
    url = f"{ZOOM_API_URL}/meetings/{meeting_id}"
    meeting_invitees = os.getenv("MEETING_INVITEES", "") # a list of emails separated by commas
    if meeting_invitees != "" and not validate_invitees(meeting_invitees.split(",")):
        raise ValueError(f"Invalid invitees: {meeting_invitees}. Must be a list of valid email addresses separated by commas.")
    agenda = os.getenv("AGENDA", "My Meeting")
    default_password = str_to_bool(os.getenv("DEFAULT_PASSWORD", "false"))
    duration = int(os.getenv("DURATION", 60))
    password = os.getenv("PASSWORD", "123456")
    pre_schedule = str_to_bool(os.getenv("PRE_SCHEDULE", "false"))
    audio_recording = os.getenv("AUDIO_RECORDING", "none")
    contact_email = os.getenv("CONTACT_EMAIL", "")
    contact_name = os.getenv("CONTACT_NAME", "")
    private_meeting = str_to_bool(os.getenv("PRIVATE_MEETING", "false"))
    start_time = os.getenv("START_TIME", "")
    if start_time != "" and not validate_meeting_start_time(start_time):
        raise ValueError(f"Invalid start time format: {start_time}. Must be in GMT or local timezone format.")
    meeting_template_id = os.getenv("MEETING_TEMPLATE_ID", "")
    timezone = os.getenv("TIMEZONE", "")
    topic = os.getenv("TOPIC", "")
    meeting_type = int(os.getenv("MEETING_TYPE", 2))
    if meeting_type not in meeting_types:
        raise ValueError(f"Invalid meeting type: {meeting_type}. Must be one of: {meeting_types.keys()}")
    # TODO: support recurrence and more settings in the future
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
            "global_dial_in_countries": ["US"],
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
                "attendees_can_upvote": True
            },
            "meeting_authentication": False,
            "meeting_invitees": [{"email": invitee} for invitee in meeting_invitees.split(",")],
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
                "auto_add_meeting_participants": True
            },
            "participant_focused_meeting": False,
            "push_change_to_calendar": False,
            "auto_start_meeting_summary": False,
            "auto_start_ai_companion_questions": False,
            "device_testing": False
        },
        "start_time": start_time,
        "template_id": meeting_template_id,
        "timezone": timezone,
        "topic": topic,
        "type": meeting_type
    }
    headers = {
        "Content-Type": "application/json",
        "Authorization": "Bearer YOUR_SECRET_TOKEN"
    }

    response = requests.patch(url, json=payload, headers=headers)
    if response.status_code != 204:
        raise Exception(f"Error updating meeting: {response.json()}")   
    return response.json()


def list_meeting_templates():
    url = f"{ZOOM_API_URL}/users/me/meeting_templates"
    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
    }
    response = requests.get(url, headers=headers)
    if response.status_code != 200:
        raise Exception(f"Error listing meeting templates: {response.json()}")
    return response.json()
