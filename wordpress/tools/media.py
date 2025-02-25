from tools.helper import (
    WORDPRESS_API_URL,
    tool_registry,
    is_valid_iso8601,
    load_from_gptscript_workspace,
    setup_logger,
)
import os
import urllib.parse
import io
import mimetypes
from typing import Union

logger = setup_logger(__name__)


def _format_media_response(response_json: Union[dict, list]) -> Union[dict, list]:
    # response is either a list of dict or a single dict
    try:
        if isinstance(response_json, list):
            return [_format_media_response(media) for media in response_json]
        else:
            keys = [
                "id",
                "date",
                "date_gmt",
                "modified",
                "modified_gmt",
                "slug",
                "status",
                "type",
                "link",
                "title",
                "author",
                "media_type",
                "mime_type",
            ]
            return {key: response_json[key] for key in keys if key in response_json}
    except Exception as e:
        logger.error(f"Error formatting media response: {e}")
        return response_json


@tool_registry.register("RetrieveMedia")
def retrieve_media(client):
    media_id = os.environ["MEDIA_ID"]
    url = f"{WORDPRESS_API_URL}/media/{media_id}"

    context = os.getenv("CONTEXT", "view").lower()
    context_enum = {"view", "embed", "edit"}

    query_params = {}
    if context not in context_enum:
        raise ValueError(
            f"Error: Invalid context: {context}. context must be one of: {context_enum}."
        )
    query_params["context"] = context
    password = os.getenv("PASSWORD", None)
    if password:
        query_params["password"] = password

    response = client.get(url, params=query_params)
    if response.status_code == 200:
        return response.json()
    else:
        print(
            f"Failed to retrieve media. Error: {response.status_code}, {response.text}"
        )


@tool_registry.register("ListMedia")
def list_media(client):
    url = f"{WORDPRESS_API_URL}/media"

    query_params = {}
    context = os.getenv("CONTEXT", "view").lower()
    context_enum = {"view", "embed", "edit"}
    if context not in context_enum:
        raise ValueError(
            f"Invalid context. Valid context must be one of: {context_enum}"
        )
    query_params["context"] = context
    page = os.getenv("PAGE", 1)
    query_params["page"] = page
    per_page = os.getenv("PER_PAGE", 10)
    query_params["per_page"] = per_page
    media_type = os.getenv("MEDIA_TYPE", "")
    if media_type != "":
        media_type_enum = ["image", "video", "text", "application", "audio"]
        if media_type not in media_type_enum:
            raise ValueError(
                f"Error: Invalid media_type: {media_type}. media_type must be one of: {media_type_enum}."
            )
        query_params["media_type"] = media_type
    author_ids = os.getenv("AUTHOR_IDS", None)  # a list of comma separated author ids
    if author_ids:
        for author_id in author_ids.split(","):
            if not author_id.isdigit():
                raise ValueError(
                    f"Error: Invalid author_ids: {author_id}. Author_ids must be a list of integers separated by commas."
                )
        query_params["author"] = author_ids
    search_query = os.getenv("SEARCH_QUERY", None)
    if search_query:
        query_params["search"] = search_query

    publish_after = os.getenv("PUBLISH_AFTER", None)
    if publish_after:
        if not is_valid_iso8601(publish_after):
            raise ValueError(
                f"Error: Invalid publish_after: {publish_after}. publish_after must be a valid ISO 8601 date string, in the format of YYYY-MM-DDTHH:MM:SS, or YYYY-MM-DDTHH:MM:SS+HH:MM."
            )
        query_params["after"] = urllib.parse.quote(publish_after)
    publish_before = os.getenv("PUBLISH_BEFORE", None)
    if publish_before:
        if not is_valid_iso8601(publish_before):
            raise ValueError(
                f"Error: Invalid publish_before: {publish_before}. publish_before must be a valid ISO 8601 date string, in the format of YYYY-MM-DDTHH:MM:SS, or YYYY-MM-DDTHH:MM:SS+HH:MM."
            )
        query_params["before"] = urllib.parse.quote(publish_before)

    modified_after = os.getenv("MODIFIED_AFTER", None)
    if modified_after:
        if not is_valid_iso8601(modified_after):
            raise ValueError(
                f"Error: Invalid modified_after: {modified_after}. modified_after must be a valid ISO 8601 date string, in the format of YYYY-MM-DDTHH:MM:SS, or YYYY-MM-DDTHH:MM:SS+HH:MM."
            )
        query_params["modified_after"] = urllib.parse.quote(modified_after)

    modified_before = os.getenv("MODIFIED_BEFORE", None)
    if modified_before:
        if not is_valid_iso8601(modified_before):
            raise ValueError(
                f"Error: Invalid modified_before: {modified_before}. modified_before must be a valid ISO 8601 date string, in the format of YYYY-MM-DDTHH:MM:SS, or YYYY-MM-DDTHH:MM:SS+HH:MM."
            )
        query_params["modified_before"] = urllib.parse.quote(modified_before)
    order = os.getenv("ORDER", "desc").lower()
    order_enum = ["asc", "desc"]
    if order not in order_enum:
        raise ValueError(
            f"Error: Invalid order: {order}. order must be one of: {order_enum}."
        )
    query_params["order"] = order

    response = client.get(url, params=query_params)
    if response.status_code == 200:
        return _format_media_response(response.json())
    elif response.status_code == 401:
        print(
            f"Authentication failed: {response.status_code}. Error Message: {response.text}"
        )
    elif response.status_code == 403:
        print(
            f"Permission denied: {response.status_code}. Error Message: {response.text}"
        )
    else:
        print(f"Failed to list media. Error code: {response.status_code}")


@tool_registry.register("UploadMedia")
def upload_media(client):
    """Upload a media file to the WordPress site.

    Args:
        client (Client): The client object to use for the request.

    Raises:
        ValueError: If the media file path is not provided or the file is not found.

    Returns:
        dict: The response from the WordPress site. Includes the metadata of the uploaded media.
    """
    upload_url = f"{WORDPRESS_API_URL}/media"

    media_file_path = os.getenv("MEDIA_FILE_PATH", "")
    if media_file_path == "":
        raise ValueError("Error: Media file path is required to upload media file.")

    data = load_from_gptscript_workspace(media_file_path)

    # with open(file_path, "rb") as file:
    #     data = file.read()

    file_obj = io.BytesIO(data)
    file_name = os.path.basename(media_file_path)
    mime_type = mimetypes.guess_type(file_name)[0] or "application/octet-stream"
    files = {"file": (file_name, file_obj, mime_type)}

    response = client.post(upload_url, files=files)

    if response.status_code == 201:
        return _format_media_response(response.json())
    elif response.status_code == 401:
        print(f"Authentication failed: {response.status_code}, {response.text}")
    elif response.status_code == 403:
        print(f"Permission denied: {response.status_code}, {response.text}")
    else:
        print(f"Failed to create/upload media. Error code: {response.status_code}")


@tool_registry.register("UpdateMedia")
def update_media(client):
    """Update the metadata of a media file.

    Args:
        client (Client): The client object to use for the request.

    Raises:
        ValueError: If the media ID is not provided or the title is empty.

    Returns:
        dict: The response from the WordPress site. Includes the metadata of the updated media.
    """
    media_id = os.environ["MEDIA_ID"]
    url = f"{WORDPRESS_API_URL}/media/{media_id}"
    media_data = {}

    if "TITLE" in os.environ:
        title = os.environ["TITLE"]
        if title == "":
            raise ValueError("Error: Title to update cannot be empty.")
        media_data["title"] = title

    if "SLUG" in os.environ:
        slug = os.environ["SLUG"]
        media_data["slug"] = slug

    if "AUTHOR_ID" in os.environ:
        author_id = os.environ["AUTHOR_ID"]
        if not author_id.isdigit():
            raise ValueError(
                f"Error: Invalid author_id: {author_id}. author_id must be an integer."
            )
        media_data["author"] = int(author_id)

    response = client.post(url, json=media_data)
    if response.status_code == 200:
        return _format_media_response(response.json())
    elif response.status_code == 401:
        print(f"Authentication failed: {response.status_code}, {response.text}")
    elif response.status_code == 403:
        print(f"Permission denied: {response.status_code}, {response.text}")
    else:
        print(f"Failed to update media. Error: {response.status_code}, {response.text}")


@tool_registry.register("DeleteMedia")
def delete_media(client):
    media_id = os.environ["MEDIA_ID"]

    query_params = {"force": "true"}
    url = f"{WORDPRESS_API_URL}/media/{media_id}"  # not allowed to put media to trash thru rest api

    response = client.delete(url, params=query_params)
    if response.status_code == 200:
        return {
            "message": f"{response.status_code}. Media {media_id} deleted successfully"
        }
    elif response.status_code == 401:
        print(f"Authentication failed: {response.status_code}, {response.text}")
    elif response.status_code == 403:
        print(f"Permission denied: {response.status_code}, {response.text}")
    else:
        print(f"Failed to delete media. Error: {response.status_code}, {response.text}")
