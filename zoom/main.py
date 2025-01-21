import sys
import json
from tools import meetings, recordings, users, timezone
from tools.helper import tool_registry


def main():

    if len(sys.argv) != 2:
        print(
            f"Error running command: {' '.join(sys.argv)} \nUsage: python3 main.py <command>"
        )
        sys.exit(1)

    command = sys.argv[1]
    try:
        json_response = tool_registry.get(command)()
        print(json.dumps(json_response, indent=4))
    except Exception as e:
        print(f"Error running command: {' '.join(sys.argv)} \nError: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
