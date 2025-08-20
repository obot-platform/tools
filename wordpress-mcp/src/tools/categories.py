"""WordPress Categories management tools."""

from typing import Optional, Union, Dict, Any

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
    context: str = "view",
    page: int = 1,
    per_page: int = 10,
    search_query: Optional[str] = None,
    order: str = "asc",
    parent_id: Optional[int] = None,
    post_id: Optional[int] = None,
    slug: Optional[str] = None
) -> Dict[str, Any]:
    """List available categories in WordPress site.
    
    Args:
        context: The context of categories to list (view, embed, edit) - default: view
        page: Page number to list - default: 1
        per_page: Number of categories per page - default: 10
        search_query: Limit results to those matching a string
        order: Sort order (asc, desc) - default: asc
        parent_id: Limit to categories assigned to a specific parent ID
        post_id: Limit to categories assigned to a specific post ID
        slug: Limit to category matching a specific slug
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
    category_name: str,
    description: Optional[str] = None,
    slug: Optional[str] = None,
    parent_id: Optional[int] = None
) -> Dict[str, Any]:
    """Create a new category in WordPress site.
    
    Args:
        category_name: The name of the category
        description: The description of the category (accepts HTML tags)
        slug: The slug for the category
        parent_id: The ID of the parent category
    """
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
    category_id: int,
    name: Optional[str] = None,
    description: Optional[str] = None,
    slug: Optional[str] = None,
    parent_id: Optional[int] = None
) -> Dict[str, Any]:
    """Update an existing category in WordPress site. Only provided fields will be updated.
    
    Args:
        category_id: The ID of the category to update
        name: New name of the category
        description: New description of the category (accepts HTML tags)
        slug: New slug for the category
        parent_id: New parent ID of the category
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
def delete_category(category_id: int) -> Dict[str, Any]:
    """Delete a category in WordPress site.
    
    Args:
        category_id: The ID of the category to delete
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