from tools.helper import ZOOM_API_URL, ACCESS_TOKEN, tool_registry
import requests
import os


@tool_registry.decorator("GetMeetingRecordings")
def get_meeting_recordings():
    meeting_id = os.environ["MEETING_ID"]
    url = f"{ZOOM_API_URL}/meetings/{meeting_id}/recordings"
    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
    }
    response = requests.get(url, headers=headers)
    if response.status_code != 200:
        raise Exception(f"Error getting meeting recordings: {response.text}")
    return response.json()


@tool_registry.decorator("ListUserRecordings")
def list_user_recordings():
    url = f"{ZOOM_API_URL}/users/me/recordings"
    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
    }
    response = requests.get(url, headers=headers)
    if response.status_code != 200:
        raise Exception(f"Error listing user recordings: {response.text}")
    return response.json()
