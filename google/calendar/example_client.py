import asyncio
from fastmcp import Client
import os
import json

GOOGLE_OAUTH_TOKEN = os.getenv("GOOGLE_OAUTH_TOKEN")
PORT = os.getenv("PORT", "9000")

async def example_list_calendars():
    async with Client(f"http://127.0.0.1:{PORT}/mcp/google-drive") as client:
        res = await client.call_tool(
            name="list_calendars",
            arguments={"cred_token": GOOGLE_OAUTH_TOKEN},
        )
        print("list_calendars result:")
        print(res[0].text)

async def example_get_calendar():
    calendar_id = "primary"  # Replace with a real calendar ID
    async with Client(f"http://127.0.0.1:{PORT}/mcp/google-drive") as client:
        res = await client.call_tool(
            name="get_calendar",
            arguments={
                "calendar_id": calendar_id,
                "cred_token": GOOGLE_OAUTH_TOKEN,
            },
        )
        print("get_calendar result:")
        print(res[0].text)

async def example_list_events():
    calendar_id = "primary"  # Replace with a real calendar ID
    async with Client(f"http://127.0.0.1:{PORT}/mcp/google-drive") as client:
        res = await client.call_tool(
            name="list_events",
            arguments={
                "calendar_id": calendar_id,
                "event_type": "default",
                "single_event": False,
                # "time_min": "2023-06-03T10:00:00Z",
                "max_results": 10,
                "cred_token": GOOGLE_OAUTH_TOKEN,
            },
        )
        print("list_events result:")
        print(res[0].text)
        return res

async def example_get_event(event_id):
    calendar_id = "primary"  # Replace with a real calendar ID
    async with Client(f"http://127.0.0.1:{PORT}/mcp/google-drive") as client:
        res = await client.call_tool(
            name="get_event",
            arguments={
                "calendar_id": calendar_id,
                "event_id": event_id,
                "cred_token": GOOGLE_OAUTH_TOKEN,
            },
        )
        print("get_event result:")
        print(res[0].text)

async def example_quick_add_event():
    calendar_id = "primary"  # Replace with a real calendar ID
    async with Client(f"http://127.0.0.1:{PORT}/mcp/google-drive") as client:
        res = await client.call_tool(
            name="quick_add_event",
            arguments={
                "text": "Quick add event Test",
                "calendar_id": calendar_id,
                "cred_token": GOOGLE_OAUTH_TOKEN,
            },
        )
        print("quick_add_event result:")
        print(res[0].text)

if __name__ == "__main__":
    asyncio.run(example_list_calendars())
    asyncio.run(example_get_calendar())
    res = asyncio.run(example_list_events())
    events = json.loads(res[0].text)
    if isinstance(events, list):
        event_id = events[0]["id"]
    elif isinstance(events, dict):
        event_id = events["id"]
    else:
        print("No events found")
    asyncio.run(example_get_event(event_id))
    # asyncio.run(example_quick_add_event())