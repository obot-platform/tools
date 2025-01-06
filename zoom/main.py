import requests
import os
import sys
import json
from tools.users import get_user
from tools.meetings import create_meeting, list_meeting_templates, get_meeting, delete_meeting, update_meeting, list_meetings

tool_map = {
    "GetUser": get_user,
    "CreateMeeting": create_meeting,
    "ListMeetingTemplates": list_meeting_templates,
    "GetMeeting": get_meeting,
    "DeleteMeeting": delete_meeting,
    "UpdateMeeting": update_meeting,
    "ListMeetings": list_meetings,
}

def main():
    if len(sys.argv) != 2:
        print(f"Error running command: {' '.join(sys.argv)} \nUsage: python3 main.py <command>")
        sys.exit(1)

    command = sys.argv[1]
    try:
        json_response = tool_map[command]()
        print(json.dumps(json_response, indent=4))
    except Exception as e:
        print(f"Error running command: {' '.join(sys.argv)} \nError: {e}")
        sys.exit(1)

    


if __name__ == "__main__":
    main()
