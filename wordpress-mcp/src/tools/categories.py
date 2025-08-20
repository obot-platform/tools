"""WordPress Categories management tools."""

from typing import Optional, Union, Dict, Any, Annotated

from src.server import mcp
from src.config import config


def _format_category_response(response_json: Union[dict, list]) -> Union[dict, list]:
    """Format category response to include only relevant fields."""
    try:
        if isinstance(response_json, list):
            return [_format_category_response(category) for category in response_json]
        else:
            keys = [
                "id", "count", "description", "name", "parent", "slug", "taxonomy"
            ]
            return {key: response_json[key] for key in keys if key in response_json}
    except Exception:
        return response_json


@mcp.tool
def list_categories(
    context: Annotated[str, "The context of categories to list (view, embed, edit) - default: view"] = "view",
    page: Annotated[int, "Page number to list - default: 1"] = 1,
    per_page: Annotated[int, "Number of categories per page - default: 10"] = 10,
    search_query: Annotated[Optional[str], "Limit results to those matching a string - default: None"] = None,
    order: Annotated[str, "Sort order (asc, desc) - default: asc"] = "asc",
    parent_id: Annotated[Optional[int], "Limit to categories assigned to a specific parent ID - default: None"] = None,
    post_id: Annotated[Optional[int], "Limit to categories assigned to a specific post ID - default: None"] = None,
    slug: Annotated[Optional[str], "Limit to category matching a specific slug - default: None"] = None
) -> Dict[str, Any]:
    """List available categories in WordPress site."""
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
    if parent_id is not None:
        params["parent"] = parent_id
    if post_id is not None:
        params["post"] = post_id
    if slug:
        params["slug"] = slug
    
    session = config.create_session()
    response = session.get(f"{config.api_url}/categories", params=params)
    
    if response.status_code == 200:
        return {"categories": _format_category_response(response.json())}
    elif response.status_code == 401:
        raise Exception(f"Authentication failed: {response.text}")
    elif response.status_code in [400, 403]:
        raise Exception(f"Permission denied: {response.text}")
    else:
        raise Exception(f"Failed to list categories: {response.status_code} - {response.text}")


@mcp.tool
def create_category(
    category_name: Annotated[str, "The name of the category"],
    description: Annotated[Optional[str], "The description of the category (accepts HTML tags) - default: None"] = None,
    slug: Annotated[Optional[str], "The slug for the category - default: None"] = None,
    parent_id: Annotated[Optional[int], "The ID of the parent category - default: None"] = None
) -> Dict[str, Any]:
    """Create a new category in WordPress site."""
    if not category_name.strip():
        raise ValueError("Category name is required")
    
    category_data = {"name": category_name.strip()}
    
    if description is not None:
        category_data["description"] = description
    if slug:
        category_data["slug"] = slug
    if parent_id is not None:
        category_data["parent"] = parent_id
    
    session = config.create_session()
    response = session.post(f"{config.api_url}/categories", json=category_data)
    
    if response.status_code == 201:
        return _format_category_response(response.json())
    elif response.status_code == 401:
        raise Exception(f"Authentication failed: {response.text}")
    elif response.status_code == 403:
        raise Exception(f"Permission denied: {response.text}")
    elif response.status_code == 400:
        raise Exception(f"Bad request - category may already exist: {response.text}")
    else:
        raise Exception(f"Failed to create category: {response.status_code} - {response.text}")


@mcp.tool
def update_category(
    category_id: Annotated[int, "The ID of the category to update"],
    name: Annotated[Optional[str], "New name of the category - default: None"] = None,
    description: Annotated[Optional[str], "New description of the category (accepts HTML tags) - default: None"] = None,
    slug: Annotated[Optional[str], "New slug for the category - default: None"] = None,
    parent_id: Annotated[Optional[int], "New parent ID of the category - default: None"] = None
) -> Dict[str, Any]:
    """Update an existing category in WordPress site. Only provided fields will be updated.
    
    At least one field must be provided to update.
    """
    # Build update data (only include provided fields)
    category_data = {}
    
    if name is not None:
        if not name.strip():
            raise ValueError("Category name cannot be empty")
        category_data["name"] = name.strip()
        
    if description is not None:
        category_data["description"] = description
    if slug is not None:
        category_data["slug"] = slug
    if parent_id is not None:
        category_data["parent"] = parent_id
    
    if not category_data:
        raise ValueError("At least one field must be provided to update")
    
    session = config.create_session()
    response = session.post(f"{config.api_url}/categories/{category_id}", json=category_data)
    
    if response.status_code == 200:
        return _format_category_response(response.json())
    elif response.status_code == 401:
        raise Exception(f"Authentication failed: {response.text}")
    elif response.status_code == 403:
        raise Exception(f"Permission denied: {response.text}")
    elif response.status_code == 404:
        raise Exception(f"Category not found: {response.text}")
    else:
        raise Exception(f"Failed to update category: {response.status_code} - {response.text}")


@mcp.tool
def delete_category(category_id: Annotated[int, "The ID of the category to delete"]) -> Dict[str, Any]:
    """Delete a category in WordPress site.
    
    Posts previously assigned to this category will be moved to 'Uncategorized'.
    """
    session = config.create_session()
    response = session.delete(f"{config.api_url}/categories/{category_id}")
    
    if response.status_code == 200:
        return {
            "message": f"Category {category_id} deleted successfully. "
                      "Posts previously assigned to this category will be moved to 'Uncategorized'."
        }
    elif response.status_code == 401:
        raise Exception(f"Authentication failed: {response.text}")
    elif response.status_code == 403:
        raise Exception(f"Permission denied: {response.text}")
    elif response.status_code == 404:
        raise Exception(f"Category not found: {response.text}")
    else:
        raise Exception(f"Failed to delete category: {response.status_code} - {response.text}")