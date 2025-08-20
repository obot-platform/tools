"""WordPress Users management tools."""

from typing import Optional, Dict, Any, Union, Annotated

from src.server import mcp
from src.config import config


def _format_users_response(response_json: Union[dict, list]) -> Union[dict, list]:
    """Format users response to include only relevant fields."""
    try:
        if isinstance(response_json, list):
            return [_format_users_response(user) for user in response_json]
        else:
            keys = [
                "id", "name", "url", "description", "link", "slug", "locale",
                "avatar_urls", "roles", "capabilities", "extra_capabilities", 
                "registered_date"
            ]
            return {key: response_json[key] for key in keys if key in response_json}
    except Exception:
        return response_json


@mcp.tool
def list_users(
    context: Annotated[str, "The context of users to list (view, embed, edit) - default: view"] = "view",
    has_published_posts: Annotated[bool, "Whether to show users who haven't published posts - default: true"] = True
) -> Dict[str, Any]:
    """List users in WordPress site. Only admin users have permission to do this."""
    if context not in ["view", "embed", "edit"]:
        raise ValueError(f"Invalid context: {context}")
    
    params = {
        "context": context,
        "has_published_posts": has_published_posts
    }
    
    session = config.create_session()
    response = session.get(f"{config.api_url}/users", params=params)
    
    if response.status_code == 200:
        return {"users": _format_users_response(response.json())}
    elif response.status_code == 401:
        raise Exception(f"Authentication failed: {response.text}")
    elif response.status_code == 403:
        raise Exception(f"Permission denied: {response.text}")
    else:
        raise Exception(f"Failed to list users: {response.status_code} - {response.text}")


@mcp.tool
def get_me(context: Annotated[str, "The context of user info to retrieve (view, embed, edit) - default: edit"] = "edit") -> Dict[str, Any]:
    """Get all metadata of the current user in WordPress site, including roles and capabilities.
    
    Failed to get user info indicates authentication is not working correctly.
    """
    if context not in ["view", "embed", "edit"]:
        raise ValueError(f"Invalid context: {context}")
    
    params = {"context": context}
    
    session = config.create_session()
    response = session.get(f"{config.api_url}/users/me", params=params)
    
    if response.status_code == 200:
        return _format_users_response(response.json())
    else:
        raise Exception(f"Failed to get current user info: {response.status_code} - {response.text}")


@mcp.tool
def validate_credential() -> Dict[str, Any]:
    """Validate WordPress credentials by attempting to get current user profile.
    This is useful for testing authentication before performing other operations.
    """
    try:
        session = config.create_session()
        response = session.get(f"{config.api_url}/users/me", params={"context": "view"})
        
        if response.status_code == 200:
            return {"valid": True, "message": "Credentials are valid"}
        else:
            return {"valid": False, "message": f"Authentication failed: {response.text}"}
    except Exception as e:
        return {"valid": False, "message": f"Validation error: {str(e)}"}