import asyncio
import os

from apis.drafts import list_drafts
from apis.helpers import client


async def list_drafts_tool():
    max_results = os.getenv("MAX_RESULTS", "100")
    if max_results is not None:
        max_results = int(max_results)

    service = client("gmail", "v1")

    await list_drafts(service, max_results)
