from apis.shared_drives import list_drives
from apis.files import list_files
from fastmcp import FastMCP
from pydantic import Field
from typing import Annotated, Literal, Union, Optional
import os
from apis.helper import get_client
from googleapiclient.errors import HttpError
from fastmcp.exceptions import ToolError

# Configure server-specific settings
PORT = os.getenv("PORT", 9000)
MCP_PATH = os.getenv("MCP_PATH", "/mcp/drive")

mcp = FastMCP(
    name="GoogleDriveMCPServer",
    on_duplicate_tools="error",                  # Handle duplicate registrations
    on_duplicate_resources="warn",
    on_duplicate_prompts="replace",
)

@mcp.tool(
    exclude_args=["cred_token"],
)
def list_shared_drives(
    cred_token: str = None) -> list[dict]:
    """
    List all shared Google Drives for the user.
    """
    client = get_client(cred_token)
    drives = list_drives(client)
    return drives

@mcp.tool(
    name="list_files",
    exclude_args=["cred_token"],
)
def list_files_tool(
    drive_id: Annotated[str, Field(description="ID of the Google Drive to list files from. If unset, default to the user's personal drive.")] = None,
    parent_id: Annotated[str, Field(description="ID of the parent folder to list files from. If unset, default to the root folder of user's personal drive.")] = None,
    mime_type: Annotated[str, Field(description="Filter files by MIME type (e.g., 'application/pdf' for PDFs, 'image/jpeg' for JPEG images, 'application/vnd.google-apps.folder' for folders). If unset, returns all file types.")] = None,
    file_name_contains: Annotated[str, Field(description="Case-insensitive search string to filter files by name. Returns files containing this string in their name.")] = None,
    modified_time_after: Annotated[str, Field(description="Return only files modified after this timestamp (RFC 3339 format: YYYY-MM-DDTHH:MM:SSZ, e.g., '2024-03-20T10:00:00Z').")] = None,
    max_results: Annotated[int, Field(description="Maximum number of files to return", ge=1, le=1000, default=50)] = 50,
    cred_token: str = None,) -> list[dict]:
    """
    List or search for files in the user's Google Drive. Returns up to 50 files by default, sorted by last modified date.
    """
    try:
        client = get_client(cred_token)

        files = list_files(
            client,
            drive_id=drive_id,
            parent_id=parent_id,
            mime_type=mime_type,
            file_name_contains=file_name_contains,
            modified_time_after=modified_time_after,
            max_results=max_results,
            trashed=False,
        )
        
        return files
    except HttpError as error:
        raise HttpError(f"Failed to list files, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")

def streamable_http_server():
    """Main entry point for the Gmail MCP server."""
    mcp.run(
        transport="streamable-http", # fixed to streamable-http
        host="0.0.0.0",
        port=PORT,
        path=MCP_PATH,
    )

def stdio_server():
    """Main entry point for the Gmail MCP server."""
    mcp.run()


if __name__ == "__main__":
    stdio_server()
