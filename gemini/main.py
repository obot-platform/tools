"""
Gemini Image Generator
A tool for generating and manipulating images using Google's Gemini API.
"""
import os
import sys
import json
import hashlib
from pathlib import Path
from google import genai
from google.genai import types
from PIL import Image
from io import BytesIO
import gptscript


# Custom exceptions
class GeminiError(Exception):
    """Base exception for all Gemini Generator errors."""
    pass


class ConfigError(GeminiError):
    """Raised when there's a configuration error."""
    pass


def get_download_base_url(server_url: str | None, thread_id: str | None) -> str | None:
    """Construct the download base URL if server URL and thread ID are available.

    Args:
        server_url: Optional server URL
        thread_id: Optional thread ID

    Returns:
        Download base URL if both parameters are provided, None otherwise
    """
    return f"{server_url}/api/threads/{thread_id}/file" if (thread_id and server_url) else None


def title_to_safe_filename(title: str) -> str:
    """Convert a title to a safe filename in lower_snake_case.

    Args:
        title: The title to convert

    Returns:
        A safe filename with .webp extension
    """
    # Convert to lowercase
    safe_name = title.lower()

    # Replace spaces and special characters with underscores
    safe_name = ''.join(c if c.isalnum() else '_' for c in safe_name)

    # Remove consecutive underscores
    while '__' in safe_name:
        safe_name = safe_name.replace('__', '_')

    # Remove leading/trailing underscores
    safe_name = safe_name.strip('_')

    # Ensure .webp extension
    if not safe_name.endswith('.webp'):
        safe_name = f"{safe_name}.webp"

    return safe_name


async def generate_image(
        gemini: genai.Client,
        model: str,
        prompt: str,
        title: str,
        download_base_url: str | None = None,
        aspect_ratio: str = "1:1",
        safety_filter: str = "BLOCK_MEDIUM_AND_ABOVE",
        allow_people: str = "ALLOW_ADULT"
) -> dict:
    """Generate an image from a text prompt.

    Args:
        gemini: Gemini API client
        model: Model ID of the Imagen model to use for image
        prompt: Text prompt for image generation
        title: Title to use for the image file (will be converted to safe filename)
        download_base_url: Optional base URL for file downloads
        aspect_ratio: Image aspect ratio ("1:1", "3:4", "4:3", "9:16", "16:9")
        safety_filter: Safety filter level ("BLOCK_LOW_AND_ABOVE", "BLOCK_MEDIUM_AND_ABOVE", "BLOCK_ONLY_HIGH")
        allow_people: Person generation setting ("DONT_ALLOW", "ALLOW_ADULT")
    """
    if not prompt:
        raise ConfigError("Prompt is required")
    if not title:
        raise ConfigError("Title is required")
    if aspect_ratio not in {"1:1", "3:4", "4:3", "9:16", "16:9"}:
        raise ConfigError("Invalid aspect ratio")

    # Initialize gptscript client
    gptscript_client = gptscript.GPTScript()

    response = gemini.models.generate_images(
        model=model,
        prompt=prompt,
        config=types.GenerateImagesConfig(
            number_of_images=1,
            aspect_ratio=aspect_ratio,
            safety_filter_level=safety_filter,
            person_generation=allow_people,
        )
    )

    # Convert image bytes to PIL Image
    generated_image = response.generated_images[0]
    image = Image.open(BytesIO(generated_image.image.image_bytes))

    # Convert title to safe filename
    file_name = title_to_safe_filename(title)

    # Convert to WebP format in memory
    webp_buffer = BytesIO()
    image.save(webp_buffer, format='WEBP', quality=100)
    webp_bytes = webp_buffer.getvalue()

    # Save to workspace
    workspace_path = f"{'files/' if download_base_url else ''}{file_name}"
    await gptscript_client.write_file_in_workspace(workspace_path, webp_bytes)

    # Format response as a flat object
    image_details = {
        'prompt': prompt,
        'title': title,
        'workspaceFile': file_name
    }

    # Add downloadUrl if available
    if download_base_url:
        image_details['imageUrl'] = f"{download_base_url}/{file_name}"

    return image_details


if __name__ == "__main__":
    try:
        # Extract and validate all environment variables first
        api_key = os.getenv('GEMINI_API_KEY')
        if not api_key:
            raise ConfigError("GEMINI_API_KEY environment variable is not set")

        title = os.getenv('TITLE')
        if not title:
            raise ConfigError("TITLE environment variable is not set")

        # Initialize client
        client = genai.Client(api_key=api_key)

        # Ensure operation is provided
        if len(sys.argv) < 2:
            raise ConfigError("Command required (generate-image)")

        cmd = sys.argv[1]

        # Switch-like operation handling
        match cmd:
            case 'generate-image':
                import asyncio

                result = asyncio.run(generate_image(
                    client,
                    os.getenv('GEMINI_MODEL', 'imagen-3.0-generate-002'),
                    os.getenv('PROMPT'),
                    title,
                    get_download_base_url(
                        os.getenv('OBOT_SERVER_URL'),
                        os.getenv('OBOT_THREAD_ID')
                    ),
                    os.getenv('ASPECT_RATIO', '1:1'),
                    os.getenv('SAFETY_FILTER', 'BLOCK_LOW_AND_ABOVE'),
                    os.getenv('ALLOW_PEOPLE', 'ALLOW_ADULT')
                ))
                print(json.dumps(result))

            case _:
                raise ConfigError(f"Unknown command '{cmd}'")


    except (ConfigError, GeminiError) as e:
        print(f"Error: {str(e)}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Unexpected error: {str(e)}", file=sys.stderr)
        sys.exit(1)
