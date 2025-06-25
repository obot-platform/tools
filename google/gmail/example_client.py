import asyncio
from fastmcp import Client
import os
import json

GOOGLE_OAUTH_TOKEN = os.getenv("GOOGLE_OAUTH_TOKEN")
PORT = os.getenv("PORT", "9000")
MCP_PATH = os.getenv("MCP_PATH", "/mcp/gmail")

async def example_list_emails():
    async with Client(f"http://127.0.0.1:{PORT}{MCP_PATH}") as client:
        res = await client.call_tool(
            name="list_emails",
            arguments={"cred_token": GOOGLE_OAUTH_TOKEN},
        )
        print("list_emails result:")
        print(res[0].text)

if __name__ == "__main__":
    asyncio.run(example_list_emails())