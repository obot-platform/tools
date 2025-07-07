import asyncio
from fastmcp import Client


async def example():
    async with Client("http://127.0.0.1:9000/mcp/google-drive") as client:
        res = await client.ping()
        res = await client.list_tools()
        for r in res:
            print(r.name)
        # rres = await client.call_tool(  
        #     name="get_weather",
        #     arguments={"city": "New York", "cred_token": "1234567890"},
        # )
        # print(rres)

if __name__ == "__main__":
    asyncio.run(example())