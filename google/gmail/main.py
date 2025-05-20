import sys
import asyncio
from commands import (
    list_emails_tool,
    list_labels_tool,
    create_label_tool,
    update_label_tool,
    delete_label_tool,
    modify_message_labels_tool,
    list_drafts_tool,
)

async def main():

    if len(sys.argv) != 2:
        print(
            f"Error running command: {' '.join(sys.argv)} \nUsage: python3 main.py <command>"
        )
        sys.exit(1)

    command = sys.argv[1]
    try:
        match command:
            case "list_emails":
                await list_emails_tool()
            case "create_label":
                create_label_tool()
            case "list_labels":
                list_labels_tool()
            case "update_label":
                update_label_tool()
            case "delete_label":
                delete_label_tool()
            case "modify_message_labels":
                modify_message_labels_tool()
            case "list_drafts":
                await list_drafts_tool()
            case _:
                print(f"Command {command} not found")
    except Exception as e:
        print(f"Error running command: {' '.join(sys.argv)} \nError: {e}")
        sys.exit(1)


if __name__ == "__main__":
    asyncio.run(main())
