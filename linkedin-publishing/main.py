from linkedin_api.clients.restli.client import RestliClient
from tools.users import get_user
from tools.posts import create_post
import json
import sys
import asyncio


async def main():
    if len(sys.argv) < 2:
        print("Usage: python main.py <tool_name>")
        sys.exit(1)

    command = sys.argv[1]
    client = RestliClient()

    match command:
        case "GetCurrentUser":
            response = get_user(client)
        case "CreatePost":
            response = await create_post(client)
        case _:
            print(f"Unknown command: {command}")
            sys.exit(1)
    print(json.dumps(response, indent=4))


if __name__ == "__main__":
    asyncio.run(main())
