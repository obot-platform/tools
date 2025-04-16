import asyncio
import os

from helpers import client, list_messages


async def main():
    max_results = os.getenv('MAX_RESULTS', '100')
    if max_results is not None:
        max_results = int(max_results)
    category = os.getenv('CATEGORY', 'primary')
    label = os.getenv('LABEL', 'inbox')
    if label != 'inbox':
        query = f'label:{label}'
    else:
        query = f'label:{label} category:{category}'

    service = client('gmail', 'v1')
    await list_messages(service, query, max_results)


if __name__ == "__main__":
    asyncio.run(main())
