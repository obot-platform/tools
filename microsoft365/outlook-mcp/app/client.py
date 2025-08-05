"""Client module for Microsoft Graph API authentication and client creation."""

from typing import List
from msgraph import GraphServiceClient
from azure.core.credentials import AccessToken
from azure.core.credentials_async import AsyncTokenCredential
from fastmcp.exceptions import ToolError
from fastmcp.server.dependencies import get_http_headers


def get_access_token() -> str:
    """Get access token from HTTP headers."""
    headers = get_http_headers()
    access_token = headers.get("x-forwarded-access-token", None)
    if not access_token:
        raise ToolError(
            "No access token found in headers, available headers: " + str(headers)
        )
    return access_token


class StaticTokenCredential(AsyncTokenCredential):
    """Static token credential for Microsoft Graph authentication."""

    def __init__(self, token: str):
        self.token = token

    async def get_token(self, *scopes, **kwargs) -> AccessToken:
        """Get access token."""
        return AccessToken(self.token, 0)  # 0 means no expiration tracking


def create_client(scopes: List[str], token: str) -> GraphServiceClient:
    """Create a Microsoft Graph client with the given scopes."""

    credential = StaticTokenCredential(token)
    return GraphServiceClient(credentials=credential, scopes=scopes)
