import asyncio

from apis.helpers import client


async def main():
    service = client("gmail", "v1")
    try:
        profile = service.users().getProfile(userId="me").execute()
        email_address = profile["emailAddress"]
        print(email_address)
    except Exception as e:
        print(f"Error getting email address: {e}")
        return None


if __name__ == "__main__":
    asyncio.run(main())
