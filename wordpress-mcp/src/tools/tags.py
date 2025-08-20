"""WordPress Tags management tools."""

from typing import Optional, Union, Dict, Any

from src.server import mcp
from src.config import config


def _format_tag_response(response_json: Union[dict, list]) -> Union[dict, list]:
    """Format tag response to include only relevant fields."""
    try:
        if isinstance(response_json, list):
            return [_format_tag_response(tag) for tag in response_json]
        else:
            keys = [
                "id", "count", "description", "name", "slug", "taxonomy"
            ]
            return {key: response_json[key] for key in keys if key in response_json}
    except Exception:
        return response_json


@mcp.tool
def list_tags(
    context: str = "view",
    page: int = 1,
    per_page: int = 10,
    search_query: Optional[str] = None,
    order: str = "asc",
    post_id: Optional[int] = None,
    slug: Optional[str] = None
) -> Dict[str, Any]:
    """List available tags in WordPress site.
    
    Args:
        context: The context of tags to list (view, embed, edit) - default: view
        page: Page number to list - default: 1
        per_page: Number of tags per page - default: 10
        search_query: Limit results to those matching a string
        order: Sort order (asc, desc) - default: asc
        post_id: Limit to tags assigned to a specific post ID
        slug: Limit to tag matching a specific slug
    """
    # Validate parameters
    if context not in ["view", "embed", "edit"]:
        raise ValueError(f"Invalid context: {context}")
    
    if order not in ["asc", "desc"]:
        raise ValueError(f"Invalid order: {order}")
    
    # Build query parameters
    params = {
        "context": context,
        "page": page,
        "per_page": per_page,
        "order": order
    }
    
    if search_query:
        params["search"] = search_query
    if post_id is not None:
        params["post"] = post_id
    if slug:
        params["slug"] = slug
    
    session = config.create_session()
    response = session.get(f"{config.api_url}/tags", params=params)
    
    if response.status_code == 200:
        return {"tags": _format_tag_response(response.json())}
    elif response.status_code == 401:
        raise Exception(f"Authentication failed: {response.text}")
    elif response.status_code in [400, 403]:
        raise Exception(f"Permission denied: {response.text}")
    else:
        raise Exception(f"Failed to list tags: {response.status_code} - {response.text}")


@mcp.tool
def create_tag(
    name: str,
    description: Optional[str] = None,
    slug: Optional[str] = None
) -> Dict[str, Any]:
    """Create a new tag in WordPress site.
    
    Args:
        name: The name of the tag
        description: The description of the tag (accepts HTML tags)
        slug: The slug for the tag
    """
    if not name.strip():
        raise ValueError("Tag name is required")
    
    tag_data = {"name": name.strip()}
    
    if description is not None:
        tag_data["description"] = description
    if slug:
        tag_data["slug"] = slug
    
    session = config.create_session()
    response = session.post(f"{config.api_url}/tags", json=tag_data)
    
    if response.status_code == 201:
        return _format_tag_response(response.json())
    elif response.status_code == 401:
        raise Exception(f"Authentication failed: {response.text}")
    elif response.status_code == 403:
        raise Exception(f"Permission denied: {response.text}")
    elif response.status_code == 400:
        raise Exception(f"Bad request - tag may already exist: {response.text}")
    else:
        raise Exception(f"Failed to create tag: {response.status_code} - {response.text}")


@mcp.tool
def update_tag(
    tag_id: int,
    name: Optional[str] = None,
    description: Optional[str] = None,
    slug: Optional[str] = None
) -> Dict[str, Any]:
    """Update an existing tag in WordPress site. Only provided fields will be updated.
    
    Args:
        tag_id: The ID of the tag to update
        name: New name of the tag
        description: New description of the tag (accepts HTML tags)
        slug: New slug for the tag
    """
    # Build update data (only include provided fields)
    tag_data = {}
    
    if name is not None:
        if not name.strip():
            raise ValueError("Tag name cannot be empty")
        tag_data["name"] = name.strip()
        
    if description is not None:
        tag_data["description"] = description
    if slug is not None:
        tag_data["slug"] = slug
    
    if not tag_data:
        raise ValueError("At least one field must be provided to update")
    
    session = config.create_session()
    response = session.post(f"{config.api_url}/tags/{tag_id}", json=tag_data)
    
    if response.status_code == 200:
        return _format_tag_response(response.json())
    elif response.status_code == 401:
        raise Exception(f"Authentication failed: {response.text}")
    elif response.status_code == 403:
        raise Exception(f"Permission denied: {response.text}")
    elif response.status_code == 404:
        raise Exception(f"Tag not found: {response.text}")
    else:
        raise Exception(f"Failed to update tag: {response.status_code} - {response.text}")


@mcp.tool
def delete_tag(tag_id: int) -> Dict[str, Any]:
    """Delete a tag in WordPress site.
    
    Args:
        tag_id: The ID of the tag to delete
    """
    session = config.create_session()
    response = session.delete(f"{config.api_url}/tags/{tag_id}")
    
    if response.status_code == 200:
        return {
            "message": f"Tag {tag_id} deleted successfully. "
                      "Posts previously assigned to this tag will no longer have it."
        }
    elif response.status_code == 401:
        raise Exception(f"Authentication failed: {response.text}")
    elif response.status_code == 403:
        raise Exception(f"Permission denied: {response.text}")
    elif response.status_code == 404:
        raise Exception(f"Tag not found: {response.text}")
    else:
        raise Exception(f"Failed to delete tag: {response.status_code} - {response.text}")