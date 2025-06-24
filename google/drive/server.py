from apis.shared_drives import list_drives
from apis.files import list_files
from fastmcp import FastMCP
from pydantic import Field
from typing import Annotated, Literal, Union, Optional
import os
from apis.helper import get_client
from googleapiclient.errors import HttpError
from fastmcp.exceptions import ToolError

# Import all the command functions
from apis.files import download_file, copy_file, get_file, create_file, update_file, delete_file, create_folder
from apis.permissions import list_permissions, get_permission, create_permission, update_permission, delete_permission, transfer_ownership
from apis.shared_drives import create_drive, update_drive, delete_drive
from apis.workspace_file import save_to_gptscript_workspace, load_from_gptscript_workspace
import mimetypes

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

## TODO: this tool requires the support of workspace
# @mcp.tool(
#     name="download_file",
#     exclude_args=["cred_token"],
# )
# def download_file_tool(
#     file_id: Annotated[str, Field(description="ID of the file to download")],
#     cred_token: str = None,) -> dict:
#     """
#     Download a Google Drive file to the user's GPTScript workspace
#     """
#     try:
#         client = get_client(cred_token)
#         content, file_name = download_file(client, file_id)
#         if content:
#             try:
#                 save_to_gptscript_workspace(file_name, content)
#                 return {"success": True, "message": f"Successfully downloaded file to: {file_name}", "file_name": file_name}
#             except Exception as e:
#                 raise ToolError(f"Error saving file to workspace: {e}")
#         else:
#             raise ToolError("Failed to download file")
#     except HttpError as error:
#         raise HttpError(f"Failed to download file, HttpError: {error}")
#     except Exception as error:
#         raise ToolError(f"Unexpected ToolError: {error}")

@mcp.tool(
    name="copy_file",
    exclude_args=["cred_token"],
)
def copy_file_tool(
    file_id: Annotated[str, Field(description="ID of the file to copy")],
    new_name: Annotated[str, Field(description="New name for the copied file. If not provided, the copied file will be named \"Copy of [original name]\".")] = None,
    new_parent_id: Annotated[str, Field(description="New parent folder ID for the copied file. Provide this if you want to have the copied file in a different folder.")] = None,
    cred_token: str = None,) -> dict:
    """
    Create a copy of a Google Drive file.
    """
    try:
        client = get_client(cred_token)
        file = copy_file(client, file_id=file_id, new_name=new_name, parent_id=new_parent_id)
        return file
    except HttpError as error:
        raise HttpError(f"Failed to copy file, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")

@mcp.tool(
    name="get_file",
    exclude_args=["cred_token"],
)
def get_file_tool(
    file_id: Annotated[str, Field(description="ID of the file to get")],
    cred_token: str = None,) -> dict:
    """
    Get a Google Drive file from user's Google Drive
    """
    try:
        client = get_client(cred_token)
        file = get_file(client, file_id)
        return file
    except HttpError as error:
        raise HttpError(f"Failed to get file, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")

## TODO: this tool requires the support of workspace
# @mcp.tool(
#     name="create_file",
#     exclude_args=["cred_token"],
# )
# def create_file_tool(
#     file_name: Annotated[str, Field(description="Name of the new file, MUST come with a file extension.")],
#     mime_type: Annotated[str, Field(description="MIME type of the new file, please provide this if you can. If not provided, it will be inferred from the file extension.")] = None,
#     parent_id: Annotated[str, Field(description="ID of the parent folder for the new file. If not provided, the file will be created in the root folder.")] = None,
#     workspace_file_path: Annotated[str, Field(description="Path to a file in the user's GPTScript workspace. If provided, the file will be used as the content of the new file.")] = None,
#     cred_token: str = None,) -> dict:
#     """
#     Create a new file in user's Google Drive. Optionally, can provide the content by providing a path to a Workspace file
#     """
#     try:
#         client = get_client(cred_token)
        
#         if "." not in file_name:
#             raise ValueError("file_name parameter must contain a file extension")
        
#         if not mime_type:
#             # Try to infer MIME type from file extension
#             guessed_type = mimetypes.guess_type(file_name)[0]
#             if guessed_type:
#                 mime_type = guessed_type
#             else:
#                 raise ValueError("Could not determine MIME type from file extension. Please provide mime_type explicitly.")

#         file_content = None
#         if workspace_file_path:
#             try:
#                 file_content = load_from_gptscript_workspace(workspace_file_path)
#             except Exception as e:
#                 raise ToolError(f"Failed to load file from GPTScript workspace: {e}")

#         file = create_file(client, name=file_name, mime_type=mime_type, parent_id=parent_id, file_content=file_content)
#         return file
#     except HttpError as error:
#         raise HttpError(f"Failed to create file, HttpError: {error}")
#     except Exception as error:
#         raise ToolError(f"Unexpected ToolError: {error}")

@mcp.tool(
    name="update_file",
    exclude_args=["cred_token"],
)
def update_file_tool(
    file_id: Annotated[str, Field(description="ID of the file or folder to update")],
    new_name: Annotated[str, Field(description="New name for the file or folder")] = None,
    new_parent_id: Annotated[str, Field(description="New parent folder ID. Provide this if you want to move the item to a different folder, use `root` to move to the root folder.")] = None,
    # new_workspace_file_path: Annotated[str, Field(description="Path to the new content of the file (not applicable for folders)")] = None,
    cred_token: str = None,) -> dict:
    """
    Update an existing file or folder in user's Google Drive. Can rename and/or move to a different location.
    """
    try:
        client = get_client(cred_token)
        
        mime_type = None
        new_content = None
        
        ## TODO: this argument requires the support of workspace
        # if new_workspace_file_path:
        #     try:
        #         new_content = load_from_gptscript_workspace(new_workspace_file_path)
        #     except Exception as e:
        #         raise ToolError(f"Failed to load file from GPTScript workspace: {e}")

        #     if new_content:
        #         file_info = get_file(client, file_id, "mimeType")
        #         mime_type = file_info.get("mimeType")

        #         # Check if it's a folder
        #         FOLDER_MIME_TYPE = "application/vnd.google-apps.folder"
        #         if mime_type == FOLDER_MIME_TYPE:
        #             raise ToolError("Cannot update content of a folder. Ignoring new content.")

        file = update_file(client, file_id=file_id, new_name=new_name, new_content=new_content, mime_type=mime_type, new_parent_id=new_parent_id)
        return file
    except HttpError as error:
        raise HttpError(f"Failed to update file, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")

@mcp.tool(
    name="create_folder",
    exclude_args=["cred_token"],
)
def create_folder_tool(
    folder_name: Annotated[str, Field(description="Name of the new folder")],
    parent_id: Annotated[str, Field(description="ID of the parent folder for the new folder. If not provided, the folder will be created in the root folder.")] = None,
    cred_token: str = None,) -> dict:
    """
    Create a new folder in user's Google Drive.
    """
    try:
        client = get_client(cred_token)
        folder = create_folder(client, name=folder_name, parent_id=parent_id)
        return folder
    except HttpError as error:
        raise HttpError(f"Failed to create folder, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")

@mcp.tool(
    name="delete_file",
    exclude_args=["cred_token"],
)
def delete_file_tool(
    file_id: Annotated[str, Field(description="ID of the file or folder to delete")],
    cred_token: str = None,) -> str:
    """
    Delete an existing file or folder from user's Google Drive
    """
    try:
        client = get_client(cred_token)
        success = delete_file(client, file_id)
        if success:
            return f"Successfully deleted file: {file_id}"
        else:
            return f"Failed to delete file: {file_id}"
    except HttpError as error:
        raise HttpError(f"Failed to delete file, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")

@mcp.tool(
    name="transfer_ownership",
    exclude_args=["cred_token"],
)
def transfer_ownership_tool(
    file_id: Annotated[str, Field(description="ID of the file to transfer ownership of")],
    new_owner_email: Annotated[str, Field(description="Email address of the new owner")],
    cred_token: str = None,) -> dict:
    """
    Transfer ownership of a Google Drive file to another user. Can only transfer ownership to a user in the same domain.
    """
    try:
        client = get_client(cred_token)
        permission = transfer_ownership(client, file_id, new_owner_email)
        return permission
    except HttpError as error:
        raise HttpError(f"Failed to transfer ownership, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")

@mcp.tool(
    name="list_permissions",
    exclude_args=["cred_token"],
)
def list_permissions_tool(
    file_id: Annotated[str, Field(description="ID of the file, folder, or shared drive to list permissions for")],
    cred_token: str = None,) -> list[dict]:
    """
    List all permissions for a Google Drive file, folder, or shared drive.
    """
    try:
        client = get_client(cred_token)
        permissions = list_permissions(client, file_id)
        return permissions
    except HttpError as error:
        raise HttpError(f"Failed to list permissions, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")

@mcp.tool(
    name="get_permission",
    exclude_args=["cred_token"],
)
def get_permission_tool(
    file_id: Annotated[str, Field(description="ID of the file, folder, or shared drive to get permission for")],
    permission_id: Annotated[str, Field(description="ID of the permission to get")],
    cred_token: str = None,) -> dict:
    """
    Get a specific permission for a Google Drive file, folder, or shared drive.
    """
    try:
        client = get_client(cred_token)
        permission = get_permission(client, file_id, permission_id)
        return permission
    except HttpError as error:
        raise HttpError(f"Failed to get permission, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")

@mcp.tool(
    name="create_permission",
    exclude_args=["cred_token"],
)
def create_permission_tool(
    file_id: Annotated[str, Field(description="ID of the file, folder, or shared drive to create permission for")],
    role: Annotated[Literal["owner", "organizer", "fileOrganizer", "writer", "commenter", "reader"], Field(description="Role for the new permission, must be one of [owner(for My Drive), organizer(for shared drive), fileOrganizer(for shared drive), writer, commenter, reader]")],
    type: Annotated[Literal["user", "group", "domain", "anyone"], Field(description="Type of the new permission, must be one of [user, group, domain, anyone]")],
    email_address: Annotated[str, Field(description="Email address for user/group permission, required if type is user or group")] = None,
    domain: Annotated[str, Field(description="Domain for domain permission, required if type is domain")] = None,
    cred_token: str = None,) -> dict:
    """
    Create a new permission for a Google Drive file, folder, or shared drive.
    """
    try:
        client = get_client(cred_token)
        
        if type in ["user", "group"] and not email_address:
            raise ToolError("EMAIL_ADDRESS is required for user/group permission")
        if type == "domain" and not domain:
            raise ToolError("DOMAIN is required for domain permission")

        permission = create_permission(client, file_id=file_id, role=role, type=type, email_address=email_address, domain=domain)
        return permission
    except HttpError as error:
        raise HttpError(f"Failed to create permission, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")

@mcp.tool(
    name="update_permission",
    exclude_args=["cred_token"],
)
def update_permission_tool(
    file_id: Annotated[str, Field(description="ID of the file, folder, or shared drive to update permission for")],
    permission_id: Annotated[str, Field(description="ID of the permission to update")],
    role: Annotated[str, Field(description="New role for the permission, must be one of [owner(for My Drive), organizer(for shared drive), fileOrganizer(for shared drive), writer, commenter, reader]")],
    cred_token: str = None,) -> dict:
    """
    Update an existing permission for a Google Drive file, folder, or shared drive.
    """
    try:
        client = get_client(cred_token)
        permission = update_permission(client, file_id, permission_id, role)
        return permission
    except HttpError as error:
        raise HttpError(f"Failed to update permission, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")

@mcp.tool(
    name="delete_permission",
    exclude_args=["cred_token"],
)
def delete_permission_tool(
    file_id: Annotated[str, Field(description="ID of the file, folder, or shared drive to delete permission from")],
    permission_id: Annotated[str, Field(description="ID of the permission to delete")],
    cred_token: str = None,) -> dict:
    """
    Delete an existing permission for a Google Drive file, folder, or shared drive.
    """
    try:
        client = get_client(cred_token)
        success = delete_permission(client, file_id, permission_id)
        if success:
            return f"Successfully deleted permission: {permission_id}"
        else:
            return f"Failed to delete permission: {permission_id}"
    except HttpError as error:
        raise HttpError(f"Failed to delete permission, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")


@mcp.tool(
    name="list_shared_drives",
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
    name="create_shared_drive",
    exclude_args=["cred_token"],
)
def create_shared_drive_tool(
    drive_name: Annotated[str, Field(description="Name of the new shared drive")],
    cred_token: str = None,) -> dict:
    """
    Create a new shared Google Drive for the user
    """
    try:
        client = get_client(cred_token)
        drive = create_drive(client, drive_name)
        return drive
    except HttpError as error:
        raise HttpError(f"Failed to create shared drive, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")

@mcp.tool(
    name="delete_shared_drive",
    exclude_args=["cred_token"],
)
def delete_shared_drive_tool(
    drive_id: Annotated[str, Field(description="ID of the shared drive to delete")],
    cred_token: str = None,) -> dict:
    """
    Delete an existing shared Google Drive
    """
    try:
        client = get_client(cred_token)
        delete_drive(client, drive_id)
        return {"success": True, "message": f"Successfully deleted shared drive: {drive_id}"}
    except HttpError as error:
        raise HttpError(f"Failed to delete shared drive, HttpError: {error}")
    except Exception as error:
        raise ToolError(f"Unexpected ToolError: {error}")

@mcp.tool(
    name="rename_shared_drive",
    exclude_args=["cred_token"],
)
def update_shared_drive_tool(
    drive_id: Annotated[str, Field(description="ID of the shared drive to rename")],
    drive_name: Annotated[str, Field(description="New name for the shared drive")],
    cred_token: str = None,) -> dict:
    """
    Rename an existing shared Google Drive
    """
    try:
        client = get_client(cred_token)
        drive = update_drive(client, drive_id, drive_name)
        return drive
    except HttpError as error:
        raise HttpError(f"Failed to update shared drive, HttpError: {error}")
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
