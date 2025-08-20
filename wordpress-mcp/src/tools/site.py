"""WordPress Site management tools."""

from typing import Dict, Any

from src.server import mcp
from src.config import config


@mcp.tool
def get_site_settings() -> Dict[str, Any]:
    """Get the settings of the WordPress site. Only admin users have permission to do this.
    
    Returns basic site information and settings that are publicly available
    or accessible to authenticated users with appropriate permissions.
    """
    session = config.create_session()
    response = session.get(f"{config.api_url}/settings")
    
    if response.status_code == 200:
        return {"settings": response.json()}
    elif response.status_code == 401:
        raise Exception(f"Authentication failed: {response.text}")
    elif response.status_code == 403:
        raise Exception(f"Permission denied - admin access required: {response.text}")
    else:
        raise Exception(f"Failed to get site settings: {response.status_code} - {response.text}")