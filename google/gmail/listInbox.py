import asyncio
import os

from apis.helpers import client
from apis.messages import list_messages


async def list_inbox():
    max_results = os.getenv('MAX_RESULTS', '100')
    if max_results is not None:
        max_results = int(max_results)
    category = os.getenv('CATEGORY', 'primary')
    label = os.getenv('LABEL', 'inbox')
    query = os.getenv('QUERY', "")
    query += f' label:{label}'
    if label == 'inbox':
        query += f' category:{category}'
    
    service = client('gmail', 'v1')
    await list_messages(service, query, max_results)


if __name__ == "__main__":
    asyncio.run(list_inbox())
