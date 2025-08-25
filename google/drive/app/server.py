from starlette.requests import Request
from starlette.responses import JSONResponse
from .apis.shared_drives import list_drives
from .apis.files import list_files
from fastmcp import FastMCP
from pydantic import Field
from typing import Annotated, Literal
import os


# Import all the command functions
from .apis.files import (
    copy_file,
    get_file,
    update_file,
    delete_file,
    create_folder,
    download_file,
)
from .apis.permissions import (
    list_permissions,
    get_permission,
    create_permission,
    update_permission,
    delete_permission,
    transfer_ownership,
)
from .apis.shared_drives import create_drive, update_drive, delete_drive
from fastmcp.server.dependencies import get_http_headers
from markitdown import MarkItDown, StreamInfo, DocumentConverterResult
from io import BytesIO
from .apis.helper import get_client
from googleapiclient.errors import HttpError
from fastmcp.exceptions import ToolError


# Configure server-specific settings
PORT = int(os.getenv("PORT", 9000))
MCP_PATH = os.getenv("MCP_PATH", "/mcp/google-drive")
GOOGLE_OAUTH_TOKEN = os.getenv("GOOGLE_OAUTH_TOKEN")

mcp = FastMCP(
    name="GoogleDriveMCPServer",
    on_duplicate_tools="error",  # Handle duplicate registrations
    on_duplicate_resources="warn",
    on_duplicate_prompts="replace",
)


@mcp.custom_route("/health", methods=["GET"])
async def health_check(request: Request):
    return JSONResponse({"status": "healthy"})


def _get_access_token() -> str:
    headers = get_http_headers()
    access_token = headers.get("x-forwarded-access-token", None)
    if not access_token:
        raise ToolError("No access token found in headers")
    return access_token


@mcp.tool(
    name="list_files",
    annotations={
        "readOnlyHint": True,
    },
)
def list_files_tool(
    drive_id: Annotated[
        str | None,
        Field(
            description="ID of the Google Drive to list files from. If unset, default to the user's personal drive."
        ),
    ] = None,
    parent_id: Annotated[
        str | None,
        Field(
            description="ID of the parent folder to list files from. If unset, default to the root folder of user's personal drive."
        ),
    ] = None,
    mime_type: Annotated[
        str | None,
        Field(
            description="Filter files by MIME type (e.g., 'application/pdf' for PDFs, 'image/jpeg' for JPEG images, 'application/vnd.google-apps.folder' for folders). If unset, returns all file types."
        ),
    ] = None,
    file_name_contains: Annotated[
        str | None,
        Field(
            description="Case-insensitive search string to filter files by name. Returns files containing this string in their name."
        ),
    ] = None,
    modified_time_after: Annotated[
        str | None,
        Field(
            description="Return only files modified after this timestamp (RFC 3339 format: YYYY-MM-DDTHH:MM:SSZ, e.g., '2024-03-20T10:00:00Z')."
        ),
    ] = None,
    max_results: Annotated[
        int,
        Field(
            description="Maximum number of files to return", ge=1, le=1000, default=50
        ),
    ] = 50,
) -> list[dict]:
    """
    List or search for files in the user's Google Drive. Returns up to 50 files by default, sorted by last modified date.
    """
    try:
        client = get_client(_get_access_token())
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
        raise ToolError(f"Failed to list files, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")


@mcp.tool(
    name="copy_file",
)
def copy_file_tool(
    file_id: Annotated[str, Field(description="ID of the file to copy")],
    new_name: Annotated[
        str | None,
        Field(
            description='New name for the copied file. If not provided, the copied file will be named "Copy of [original name]".'
        ),
    ] = None,
    new_parent_id: Annotated[
        str | None,
        Field(
            description="New parent folder ID for the copied file. Provide this if you want to have the copied file in a different folder."
        ),
    ] = None,
) -> dict:
    """
    Create a copy of a Google Drive file.
    """
    try:
        client = get_client(_get_access_token())
        file = copy_file(
            client, file_id=file_id, new_name=new_name, parent_id=new_parent_id
        )
        return file
    except HttpError as error:
        raise ToolError(f"Failed to copy file, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")


@mcp.tool(
    name="get_file",
    annotations={
        "readOnlyHint": True,
    },
)
def get_file_tool(
    file_id: Annotated[str, Field(description="ID of the file to get")],
) -> dict:
    """
    Get a Google Drive file from user's Google Drive
    """
    try:
        client = get_client(_get_access_token())
        file = get_file(client, file_id)
        return file
    except HttpError as error:
        raise ToolError(f"Failed to get file, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")


@mcp.tool(
    name="update_file",
)
def update_file_tool(
    file_id: Annotated[str, Field(description="ID of the file or folder to update")],
    new_name: Annotated[
        str | None, Field(description="New name for the file or folder")
    ] = None,
    new_parent_id: Annotated[
        str | None,
        Field(
            description="New parent folder ID. Provide this if you want to move the item to a different folder, use `root` to move to the root folder."
        ),
    ] = None,
    # new_workspace_file_path: Annotated[str, Field(description="Path to the new content of the file (not applicable for folders)")] = None,
) -> dict:
    """
    Update an existing file or folder in user's Google Drive. Can rename and/or move to a different location.
    """
    try:
        client = get_client(_get_access_token())

        mime_type = None
        new_content = None

        file = update_file(
            client,
            file_id=file_id,
            new_name=new_name,
            new_content=new_content,
            mime_type=mime_type,
            new_parent_id=new_parent_id,
        )
        return file
    except HttpError as error:
        raise ToolError(f"Failed to update file, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")


@mcp.tool(
    name="create_folder",
)
def create_folder_tool(
    folder_name: Annotated[str, Field(description="Name of the new folder")],
    parent_id: Annotated[
        str | None,
        Field(
            description="ID of the parent folder for the new folder. If not provided, the folder will be created in the root folder."
        ),
    ] = None,
) -> dict:
    """
    Create a new folder in user's Google Drive.
    """
    try:
        client = get_client(_get_access_token())
        folder = create_folder(client, name=folder_name, parent_id=parent_id)
        return folder
    except HttpError as error:
        raise ToolError(f"Failed to create folder, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")


@mcp.tool(
    name="delete_file",
)
def delete_file_tool(
    file_id: Annotated[str, Field(description="ID of the file or folder to delete")],
) -> str:
    """
    Delete an existing file or folder from user's Google Drive
    ALWAYS ask for user's confirmation before proceeding this tool.
    """
    try:
        client = get_client(_get_access_token())
        success = delete_file(client, file_id)
        if success:
            return f"Successfully deleted file: {file_id}"
        else:
            return f"Failed to delete file: {file_id}"
    except HttpError as error:
        raise ToolError(f"Failed to delete file, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")


@mcp.tool(
    name="transfer_ownership",
)
def transfer_ownership_tool(
    file_id: Annotated[
        str, Field(description="ID of the file to transfer ownership of")
    ],
    new_owner_email: Annotated[
        str, Field(description="Email address of the new owner")
    ],
) -> dict:
    """
    Transfer ownership of a Google Drive file to another user. Can only transfer ownership to a user in the same domain.
    """
    try:
        client = get_client(_get_access_token())
        permission = transfer_ownership(client, file_id, new_owner_email)
        return permission
    except HttpError as error:
        raise ToolError(f"Failed to transfer ownership, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")


@mcp.tool(
    name="list_permissions",
    annotations={
        "readOnlyHint": True,
    },
)
def list_permissions_tool(
    file_id: Annotated[
        str,
        Field(
            description="ID of the file, folder, or shared drive to list permissions for"
        ),
    ],
) -> list[dict]:
    """
    List all permissions for a Google Drive file, folder, or shared drive.
    """
    try:
        client = get_client(_get_access_token())
        permissions = list_permissions(client, file_id)
        return permissions
    except HttpError as error:
        raise ToolError(f"Failed to list permissions, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")


@mcp.tool(
    name="get_permission",
    annotations={
        "readOnlyHint": True,
    },
)
def get_permission_tool(
    file_id: Annotated[
        str,
        Field(
            description="ID of the file, folder, or shared drive to get permission for"
        ),
    ],
    permission_id: Annotated[str, Field(description="ID of the permission to get")],
) -> dict:
    """
    Get a specific permission for a Google Drive file, folder, or shared drive.
    """
    try:
        client = get_client(_get_access_token())
        permission = get_permission(client, file_id, permission_id)
        return permission
    except HttpError as error:
        raise ToolError(f"Failed to get permission, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")


@mcp.tool(
    name="create_permission",
)
def create_permission_tool(
    file_id: Annotated[
        str,
        Field(
            description="ID of the file, folder, or shared drive to create permission for"
        ),
    ],
    role: Annotated[
        Literal["owner", "organizer", "fileOrganizer", "writer", "commenter", "reader"],
        Field(
            description="Role for the new permission, must be one of [owner(for My Drive), organizer(for shared drive), fileOrganizer(for shared drive), writer, commenter, reader]"
        ),
    ],
    type: Annotated[
        Literal["user", "group", "domain", "anyone"],
        Field(
            description="Type of the new permission, must be one of [user, group, domain, anyone]"
        ),
    ],
    email_address: Annotated[
        str | None,
        Field(
            description="Email address for user/group permission, required if type is user or group"
        ),
    ] = None,
    domain: Annotated[
        str | None,
        Field(description="Domain for domain permission, required if type is domain"),
    ] = None,
) -> dict:
    """
    Create a new permission for a Google Drive file, folder, or shared drive.
    """
    try:
        client = get_client(_get_access_token())

        if type in ["user", "group"] and not email_address:
            raise ToolError("EMAIL_ADDRESS is required for user/group permission")
        if type == "domain" and not domain:
            raise ToolError("DOMAIN is required for domain permission")

        permission = create_permission(
            client,
            file_id=file_id,
            role=role,
            type=type,
            email_address=email_address,
            domain=domain,
        )
        return permission
    except HttpError as error:
        raise ToolError(f"Failed to create permission, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")


@mcp.tool(
    name="update_permission",
)
def update_permission_tool(
    file_id: Annotated[
        str,
        Field(
            description="ID of the file, folder, or shared drive to update permission for"
        ),
    ],
    permission_id: Annotated[str, Field(description="ID of the permission to update")],
    role: Annotated[
        str,
        Field(
            description="New role for the permission, must be one of [owner(for My Drive), organizer(for shared drive), fileOrganizer(for shared drive), writer, commenter, reader]"
        ),
    ],
) -> dict:
    """
    Update an existing permission for a Google Drive file, folder, or shared drive.
    """
    try:
        client = get_client(_get_access_token())
        permission = update_permission(client, file_id, permission_id, role)
        return permission
    except HttpError as error:
        raise ToolError(f"Failed to update permission, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")


@mcp.tool(
    name="delete_permission",
)
def delete_permission_tool(
    file_id: Annotated[
        str,
        Field(
            description="ID of the file, folder, or shared drive to delete permission from"
        ),
    ],
    permission_id: Annotated[str, Field(description="ID of the permission to delete")],
) -> dict:
    """
    Delete an existing permission for a Google Drive file, folder, or shared drive.
    ALWAYS ask for user's confirmation before proceeding this tool.
    """
    try:
        client = get_client(_get_access_token())
        success = delete_permission(client, file_id, permission_id)
        if success:
            return {"result": f"Successfully deleted permission: {permission_id}"}
        else:
            return {"result": f"Failed to delete permission: {permission_id}"}
    except HttpError as error:
        raise ToolError(f"Failed to delete permission, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")


@mcp.tool(
    name="list_shared_drives",
    annotations={
        "readOnlyHint": True,
    },
)
def list_shared_drives() -> list[dict]:
    """
    List all shared Google Drives for the user.
    """

    client = get_client(_get_access_token())
    drives = list_drives(client)
    return drives


@mcp.tool(
    name="create_shared_drive",
)
def create_shared_drive_tool(
    drive_name: Annotated[str, Field(description="Name of the new shared drive")],
) -> dict:
    """
    Create a new shared Google Drive for the user
    """
    try:
        client = get_client(_get_access_token())
        drive = create_drive(client, drive_name)
        return drive
    except HttpError as error:
        raise ToolError(f"Failed to create shared drive, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")


@mcp.tool(
    name="delete_shared_drive",
)
def delete_shared_drive_tool(
    drive_id: Annotated[str, Field(description="ID of the shared drive to delete")],
) -> dict:
    """
    Delete an existing shared Google Drive.
    ALWAYS ask for user's confirmation before proceeding this tool.
    """
    try:
        client = get_client(_get_access_token())
        delete_drive(client, drive_id)
        return {
            "success": True,
            "message": f"Successfully deleted shared drive: {drive_id}",
        }
    except HttpError as error:
        raise ToolError(f"Failed to delete shared drive, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")


@mcp.tool(
    name="rename_shared_drive",
)
def update_shared_drive_tool(
    drive_id: Annotated[str, Field(description="ID of the shared drive to rename")],
    drive_name: Annotated[str, Field(description="New name for the shared drive")],
) -> dict:
    """
    Rename an existing shared Google Drive
    """
    try:
        client = get_client(_get_access_token())
        drive = update_drive(client, drive_id, drive_name)
        return drive
    except HttpError as error:
        raise ToolError(f"Failed to update shared drive, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")


@mcp.tool(
    name="read_file",
)
def read_file_tool(
    file_id: Annotated[str, Field(description="ID of the file to read")],
) -> DocumentConverterResult:
    """Read the content of a file in Google Drive. Files larger than 100MB will not be read."""
    try:
        client = get_client(_get_access_token())
        content, file_name = download_file(client, file_id)

        # Extract file extension, handling files without extensions
        if "." in file_name and not file_name.startswith("."):
            file_extension = file_name.split(".")[-1]
        else:
            file_extension = None
        md = MarkItDown(enable_plugins=False)
        if file_extension:
            return md.convert(
                BytesIO(content), stream_info=StreamInfo(extension=file_extension)
            )
        else:
            return md.convert(BytesIO(content))
    except HttpError as error:
        raise ToolError(f"Failed to read file, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")


def streamable_http_server():
    """Main entry point for the Gmail MCP server."""
    mcp.run(
        transport="streamable-http",  # fixed to streamable-http
        host="0.0.0.0",
        port=PORT,
        path=MCP_PATH,
    )


def stdio_server():
    """Main entry point for the Gmail MCP server."""
    mcp.run()


if __name__ == "__main__":
    streamable_http_server()
