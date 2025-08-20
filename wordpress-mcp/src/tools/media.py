"""WordPress Media management tools."""

from typing import Optional, Union, Dict, Any, Annotated
from urllib.parse import quote

from src.server import mcp
from src.config import config, is_valid_iso8601


def _format_media_response(response_json: Union[dict, list]) -> Union[dict, list]:
    """Format media response to include only relevant fields."""
    try:
        if isinstance(response_json, list):
            return [_format_media_response(media) for media in response_json]
        else:
            keys = [
                "id", "date", "date_gmt", "modified", "modified_gmt",
                "slug", "status", "type", "link", "title", "author",
                "media_type", "mime_type"
            ]
            return {key: response_json[key] for key in keys if key in response_json}
    except Exception:
        return response_json


@mcp.tool
def list_media(
    context: Annotated[str, "The context of media files to list (view, embed, edit) - default: view"] = "view",
    page: Annotated[int, "Page number to list - default: 1"] = 1,
    per_page: Annotated[int, "Number of media files per page - default: 10"] = 10,
    media_type: Annotated[Optional[str], "Limit to specific media type (image, video, text, application, audio) - default: None"] = None,
    author_ids: Annotated[Optional[str], "Comma-separated list of author IDs - default: None"] = None,
    search_query: Annotated[Optional[str], "Limit results to those matching a string - default: None"] = None,
    publish_after: Annotated[Optional[str], "ISO 8601 date to filter media files uploaded after - default: None"] = None,
    publish_before: Annotated[Optional[str], "ISO 8601 date to filter media files uploaded before - default: None"] = None,
    modified_after: Annotated[Optional[str], "ISO 8601 date to filter media files modified after - default: None"] = None,
    modified_before: Annotated[Optional[str], "ISO 8601 date to filter media files modified before - default: None"] = None,
    order: Annotated[str, "Sort order (asc, desc) - default: desc"] = "desc"
) -> Dict[str, Any]:
    """List media files in WordPress site and get basic information of each media file.
    
    Date parameters must be valid ISO 8601 date strings (YYYY-MM-DDTHH:MM:SS).
    """
    # Validate parameters
    if context not in ["view", "embed", "edit"]:
        raise ValueError(f"Invalid context: {context}")
    
    if order not in ["asc", "desc"]:
        raise ValueError(f"Invalid order: {order}")
    
    if media_type:
        media_type_enum = ["image", "video", "text", "application", "audio"]
        if media_type not in media_type_enum:
            raise ValueError(f"Invalid media_type: {media_type}")
    
    # Validate date parameters
    for date_param, date_value in [
        ("publish_after", publish_after),
        ("publish_before", publish_before),
        ("modified_after", modified_after),
        ("modified_before", modified_before)
    ]:
        if date_value and not is_valid_iso8601(date_value):
            raise ValueError(f"Invalid {date_param}: must be ISO 8601 format")
    
    # Validate author_ids
    if author_ids:
        for author_id in author_ids.split(","):
            if not author_id.strip().isdigit():
                raise ValueError(f"Invalid author_id: {author_id}")
    
    # Build query parameters
    params = {
        "context": context,
        "page": page,
        "per_page": per_page,
        "order": order
    }
    
    if media_type:
        params["media_type"] = media_type
    if author_ids:
        params["author"] = author_ids
    if search_query:
        params["search"] = search_query
    if publish_after:
        params["after"] = quote(publish_after)
    if publish_before:
        params["before"] = quote(publish_before)
    if modified_after:
        params["modified_after"] = quote(modified_after)
    if modified_before:
        params["modified_before"] = quote(modified_before)
    
    session = config.create_session()
    response = session.get(f"{config.api_url}/media", params=params)
    
    if response.status_code == 200:
        return {"media": _format_media_response(response.json())}
    elif response.status_code == 401:
        raise Exception(f"Authentication failed: {response.text}")
    elif response.status_code in [400, 403]:
        raise Exception(f"Permission denied: {response.text}")
    else:
        raise Exception(f"Failed to list media: {response.status_code} - {response.text}")


@mcp.tool
def update_media(
    media_id: Annotated[int, "The ID of the media file"],
    title: Annotated[Optional[str], "New title for the media file - default: None"] = None,
    slug: Annotated[Optional[str], "New slug for the media file - default: None"] = None,
    author_id: Annotated[Optional[int], "New author ID for the media file - default: None"] = None
) -> Dict[str, Any]:
    """Update the metadata of a media file in WordPress site.
    
    At least one field must be provided to update.
    """
    # Build update data (only include provided fields)
    media_data = {}
    
    if title is not None:
        media_data["title"] = title
    if slug is not None:
        media_data["slug"] = slug
    if author_id is not None:
        media_data["author"] = author_id
    
    if not media_data:
        raise ValueError("At least one field must be provided to update")
    
    session = config.create_session()
    response = session.post(f"{config.api_url}/media/{media_id}", json=media_data)
    
    if response.status_code == 200:
        return _format_media_response(response.json())
    elif response.status_code == 401:
        raise Exception(f"Authentication failed: {response.text}")
    elif response.status_code == 403:
        raise Exception(f"Permission denied: {response.text}")
    else:
        raise Exception(f"Failed to update media: {response.status_code} - {response.text}")


@mcp.tool
def delete_media(media_id: Annotated[int, "The ID of the media file to delete"]) -> Dict[str, Any]:
    """Delete a media file in WordPress site."""
    session = config.create_session()
    response = session.delete(f"{config.api_url}/media/{media_id}")
    
    if response.status_code == 200:
        return {
            "message": f"Media file {media_id} deleted successfully. "
                      "Note: The file may still appear temporarily due to caching. "
                      "This is normal and should resolve shortly."
        }
    elif response.status_code == 401:
        raise Exception(f"Authentication failed: {response.text}")
    elif response.status_code == 403:
        raise Exception(f"Permission denied: {response.text}")
    else:
        raise Exception(f"Failed to delete media: {response.status_code} - {response.text}")