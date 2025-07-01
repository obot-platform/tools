import asyncio
from fastmcp import Client
import os
import json
from fastmcp.client.transports import StreamableHttpTransport

GOOGLE_OAUTH_TOKEN = os.getenv("GOOGLE_OAUTH_TOKEN")
PORT = os.getenv("PORT", "9000")
MCP_PATH = os.getenv("MCP_PATH", "/mcp/google-drive")

async def example_list_shared_drives():
    async with Client(transport=StreamableHttpTransport(
        f"http://127.0.0.1:{PORT}{MCP_PATH}",
        headers={"x-forwarded-access-token": GOOGLE_OAUTH_TOKEN},
    )) as client:
        res = await client.call_tool(
            name="list_shared_drives",
            arguments={},
        )
        print("list_shared_drives result:")
        print(res[0].text)


async def example_list_files():
    async with Client(transport=StreamableHttpTransport(
        f"http://127.0.0.1:{PORT}{MCP_PATH}",
        headers={"x-forwarded-access-token": GOOGLE_OAUTH_TOKEN},
    )) as client:
        res = await client.call_tool(
            name="list_files",
            arguments={"max_results": 10},
        )
        print("list_files result:")
        print(res[0].text)

if __name__ == "__main__":
    asyncio.run(example_list_shared_drives())
    asyncio.run(example_list_files())