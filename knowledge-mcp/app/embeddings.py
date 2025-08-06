import asyncio
from typing import List
import httpx
import time
import random

# Import configuration from centralized config
from app.config import (
    OPENAI_API_KEY,
    EMBEDDING_MODEL,
    EMBEDDING_DIMENSION,
    SKIP_OPENAI_VALIDATION,
)


async def validate_openai_api_key() -> bool:
    """
    Validate the OpenAI API key by making a lightweight API call.
    Returns True if valid, raises ValueError if invalid.
    """
    try:
        async with httpx.AsyncClient() as client:
            # Use a minimal request to test the API key
            response = await client.post(
                "https://api.openai.com/v1/embeddings",
                headers={
                    "Authorization": f"Bearer {OPENAI_API_KEY}",
                    "Content-Type": "application/json",
                },
                json={"input": "test", "model": EMBEDDING_MODEL},
                timeout=10.0,
            )

            if response.status_code == 200:
                return True
            elif response.status_code == 401:
                raise ValueError("Invalid OpenAI API key - authentication failed")
            elif response.status_code == 403:
                raise ValueError(
                    "OpenAI API key does not have permission for embeddings"
                )
            elif response.status_code == 429:
                # Rate limited but key is valid
                print(
                    "Warning: OpenAI API rate limited during validation, but key appears valid"
                )
                return True
            else:
                raise ValueError(
                    f"OpenAI API validation failed: {response.status_code} - {response.text}"
                )

    except httpx.TimeoutException:
        raise ValueError(
            "OpenAI API validation timed out - please check your connection"
        )
    except httpx.ConnectError:
        raise ValueError(
            "Could not connect to OpenAI API - please check your connection"
        )


# Validate the API key on module import
def _validate_api_key_sync():
    """Synchronous wrapper for API key validation during module import."""
    try:
        # Run the async validation
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        loop.run_until_complete(validate_openai_api_key())
        loop.close()
        print("✓ OpenAI API key validated successfully")
    except Exception as e:
        raise ValueError(f"OpenAI API key validation failed: {e}")


# Only validate if we're not in a test environment
if not SKIP_OPENAI_VALIDATION:
    _validate_api_key_sync()


async def generate_embedding_openai(text: str, max_retries: int = 3) -> List[float]:
    """Generate embeddings using OpenAI API with retry logic."""
    for attempt in range(max_retries):
        try:
            async with httpx.AsyncClient() as client:
                response = await client.post(
                    "https://api.openai.com/v1/embeddings",
                    headers={
                        "Authorization": f"Bearer {OPENAI_API_KEY}",
                        "Content-Type": "application/json",
                    },
                    json={"input": text, "model": EMBEDDING_MODEL},
                    timeout=30.0,
                )

                if response.status_code == 200:
                    data = response.json()
                    return data["data"][0]["embedding"]

                # Check if it's a retryable error
                if response.status_code in [429, 500, 502, 503, 504]:
                    if attempt < max_retries - 1:
                        # Exponential backoff with jitter
                        delay = (2**attempt) + random.uniform(0, 1)
                        print(
                            f"OpenAI API error {response.status_code}, retrying in {delay:.2f}s (attempt {attempt + 1}/{max_retries})"
                        )
                        await asyncio.sleep(delay)
                        continue

                # Non-retryable error or max retries reached
                raise ValueError(
                    f"OpenAI API error: {response.status_code} - {response.text}"
                )

        except httpx.TimeoutException:
            if attempt < max_retries - 1:
                delay = (2**attempt) + random.uniform(0, 1)
                print(
                    f"OpenAI API timeout, retrying in {delay:.2f}s (attempt {attempt + 1}/{max_retries})"
                )
                await asyncio.sleep(delay)
                continue
            else:
                raise ValueError("OpenAI API timeout after all retries")

        except httpx.ConnectError:
            if attempt < max_retries - 1:
                delay = (2**attempt) + random.uniform(0, 1)
                print(
                    f"OpenAI API connection error, retrying in {delay:.2f}s (attempt {attempt + 1}/{max_retries})"
                )
                await asyncio.sleep(delay)
                continue
            else:
                raise ValueError("OpenAI API connection error after all retries")

    # This shouldn't be reached, but just in case
    raise ValueError("OpenAI API failed after all retries")


async def generate_embedding(text: str) -> List[float]:
    """
    Generate embeddings for text using OpenAI API.
    Requires OPENAI_API_KEY environment variable to be set.
    """
    return await generate_embedding_openai(text)


async def generate_embeddings_batch(
    texts: List[str], batch_size: int = 10
) -> List[List[float]]:
    """Generate embeddings for multiple texts in batches."""
    embeddings = []

    for i in range(0, len(texts), batch_size):
        batch = texts[i : i + batch_size]
        batch_embeddings = await asyncio.gather(
            *[generate_embedding(text) for text in batch]
        )
        embeddings.extend(batch_embeddings)

    return embeddings
