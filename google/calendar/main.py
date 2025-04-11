import sys
import json
from tools import calendar as calendar_tools
from tools import event as event_tools
from tools.helper import get_client


def main():

    if len(sys.argv) != 2:
        print(
            f"Error running command: {' '.join(sys.argv)} \nUsage: python3 main.py <command>"
        )
        sys.exit(1)

    command = sys.argv[1]
    service = get_client()
    try:
        match command:
            case "quick_add_event":
                json_response = event_tools.quick_add_event(service)
            case "list_events":
                json_response = event_tools.list_events(service)
            case "get_event":
                json_response = event_tools.get_event(service)
            case "move_event":
                json_response = event_tools.move_event(service)
            case "update_event":
                json_response = event_tools.update_event(service)
            case "respond_to_event":
                json_response = event_tools.respond_to_event(service)
            case "create_event":
                json_response = event_tools.create_event(service)
            case "list_recurring_event_instances":
                json_response = event_tools.recurring_event_instances(service)
            case "delete_event":
                json_response = event_tools.delete_event(service)

            case "list_calendars":
                json_response = calendar_tools.list_calendars(service)
            case "get_calendar":
                json_response = calendar_tools.get_calendar(service)
            case "create_calendar":
                json_response = calendar_tools.create_calendar(service)
            case "update_calendar":
                json_response = calendar_tools.update_calendar(service)
            case "delete_calendar":
                json_response = calendar_tools.delete_calendar(service)
            case _:
                raise ValueError(f"Invalid command: {command}")
        print(json.dumps(json_response, indent=4))
    except Exception as e:
        print(f"Error running command: {' '.join(sys.argv)} \nError: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
