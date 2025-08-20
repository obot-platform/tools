"""Configuration management for WordPress MCP Server."""

import os
from datetime import datetime
from typing import Optional
from urllib.parse import urlparse
import requests
from requests.auth import HTTPBasicAuth
from dotenv import load_dotenv

load_dotenv()


class WordPressConfig:
    """WordPress configuration and client management."""
    
    def __init__(self):
        self.site_url = self._get_required_env("WORDPRESS_SITE")
        self.username = self._get_required_env("WORDPRESS_USERNAME") 
        self.password = self._get_required_env("WORDPRESS_PASSWORD")
        
        # Clean and validate site URL
        self.api_url = self._clean_site_url(self.site_url)
        
        # Set up authentication
        self.auth = HTTPBasicAuth(self.username, self.password)
    
    def _get_required_env(self, key: str) -> str:
        """Get required environment variable or raise error."""
        value = os.getenv(key)
        if not value:
            raise ValueError(f"Required environment variable {key} is not set")
        return value
    
    def _clean_site_url(self, site_url: str) -> str:
        """Clean and validate WordPress site URL."""
        if not site_url.startswith(("https://", "http://")):
            raise ValueError(
                f"Invalid site URL: {site_url}. Must start with https:// or http://"
            )
        
        # Remove trailing slash and wp-json if present
        site_url = site_url.rstrip("/")
        if site_url.endswith("/wp-json"):
            site_url = site_url.split("/wp-json")[0]
        
        # Add WordPress API path
        return f"{site_url}/wp-json/wp/v2"
    
    def create_session(self) -> requests.Session:
        """Create a configured requests session."""
        session = requests.Session()
        session.auth = self.auth
        session.headers.update({
            "User-Agent": "WordPress-MCP-Server/1.0",
            "Content-Type": "application/json"
        })
        return session


def is_valid_iso8601(date_string: str) -> bool:
    """Validate ISO 8601 date string."""
    try:
        datetime.fromisoformat(date_string.replace('Z', '+00:00'))
        return True
    except ValueError:
        return False


def str_to_bool(value: str) -> bool:
    """Convert string to boolean."""
    return str(value).lower() in ("true", "1", "yes")


# Global config instance
config = WordPressConfig()