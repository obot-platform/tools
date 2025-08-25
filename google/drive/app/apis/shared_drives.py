from typing import Optional, List
from googleapiclient.discovery import Resource
from googleapiclient.errors import HttpError
from fastmcp.exceptions import ToolError


def _generate_ids(service: Resource, count: int = 1, space: str = "drive") -> List[str]:
    """
    Generate a new file ID using Google Drive API.

    Args:
        service: Google Drive API service instance

    Returns:
        A string containing the generated file ID, or None if an error occurs
    """
    try:
        response = service.files().generateIds(count=count, space=space).execute()
        return response.get("ids", [])
    except HttpError as error:
        error_details = error.error_details[0] if error.error_details else {}
        raise ToolError(f"Failed to generate IDs, HttpError: {error_details}")


def list_drives(service: Resource) -> List[dict]:
    """
    List all shared drives accessible to the user.

    Args:
        service: Google Drive API service instance

    Returns:
        List of drive objects containing drive information
    """
    try:
        drives = []
        page_token = None

        while True:
            response = (
                service.drives()
                .list(pageSize=100, pageToken=page_token, useDomainAdminAccess=False)
                .execute()
            )

            drives.extend(response.get("drives", []))
            page_token = response.get("nextPageToken")

            if not page_token:
                break

        return drives
    except HttpError as error:
        error_details = error.error_details[0] if error.error_details else {}
        raise ToolError(f"Error listing drives: {error_details}")


def get_drive(service: Resource, drive_id: str) -> Optional[dict]:
    """
    Get information about a specific shared drive.

    Args:
        service: Google Drive API service instance
        drive_id: ID of the shared drive

    Returns:
        Drive object if found, None otherwise
    """
    try:
        return service.drives().get(driveId=drive_id).execute()
    except HttpError as error:
        error_details = error.error_details[0] if error.error_details else {}
        raise ToolError(f"Failed to get drive, HttpError: {error_details}")


def create_drive(service: Resource, name: str) -> Optional[dict]:
    """
    Create a new shared drive.

    Args:
        service: Google Drive API service instance
        name: Name of the new shared drive

    Returns:
        Newly created drive object if successful, None otherwise
    """
    try:
        request_id = _generate_ids(service, count=1, space="drive")
        return (
            service.drives().create(requestId=request_id, body={"name": name}).execute()
        )
    except HttpError as error:
        error_details = error.error_details[0] if error.error_details else {}
        raise ToolError(f"Failed to create shared drive, HttpError: {error_details}")


def delete_drive(service: Resource, drive_id: str) -> bool:
    """
    Delete a shared drive. Note: Drive must be empty before deletion.

    Args:
        service: Google Drive API service instance
        drive_id: ID of the shared drive to delete

    Returns:
        True if successful, False otherwise
    """
    try:
        service.drives().delete(driveId=drive_id).execute()
        return True
    except HttpError as error:
        error_details = error.error_details[0] if error.error_details else {}
        raise ToolError(f"Failed to delete drive, HttpError: {error_details}")


def update_drive(service: Resource, drive_id: str, new_name: str) -> Optional[dict]:
    """
    Update a shared drive's metadata (currently only name can be updated).

    Args:
        service: Google Drive API service instance
        drive_id: ID of the shared drive to update
        new_name: New name for the shared drive

    Returns:
        Updated drive object if successful, None otherwise
    """
    try:
        return (
            service.drives().update(driveId=drive_id, body={"name": new_name}).execute()
        )
    except HttpError as error:
        error_details = error.error_details[0] if error.error_details else {}
        raise ToolError(f"Failed to update drive, HttpError: {error_details}")
