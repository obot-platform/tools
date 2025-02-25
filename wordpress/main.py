from tools import posts, users, site, media, categories, tags  # import tool registry to register tools
from tools.helper import tool_registry, create_session, setup_logger
import json
import sys

logger = setup_logger(__name__)

logger.info(f"Registered WordPress tools: {tool_registry.list_tools()}")


def main():
    if len(sys.argv) < 2:
        print("Usage: python main.py <command>")
        sys.exit(1)

    command = sys.argv[1]
    try:
        client = create_session()

        logger.info(f"Calling tool: {command}")
        json_response = tool_registry.get(command)(client)
        print(json.dumps(json_response, indent=4))
    except Exception as e:
        print(f"Running command: {' '.join(sys.argv)} failed. Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
