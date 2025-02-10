from tools import posts, users, site, media  # import tool registry
from tools.helper import tool_registry, create_session
import json
import sys


def main():
    if len(sys.argv) < 2:
        print("Usage: python main.py <command>")
        sys.exit(1)

    command = sys.argv[1]
    try:
        client = create_session()

        json_response = tool_registry.get(command)(client)
        print(json.dumps(json_response, indent=4))
    except Exception as e:
        print(f"Running command: {' '.join(sys.argv)} failed. Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
