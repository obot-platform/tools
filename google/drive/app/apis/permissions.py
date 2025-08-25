from typing import Optional, List
from googleapiclient.discovery import Resource
from googleapiclient.errors import HttpError
from fastmcp.exceptions import ToolError


def list_permissions(service: Resource, file_id: str) -> List[dict]:
    """
    List all permissions for a file.

    Args:
        service: Google Drive API service instance
        file_id: ID of the file to get permissions for

    Returns:
        List of permission objects
    """
    try:
        permissions = []
        page_token = None

        while True:
            response = (
                service.permissions()
                .list(
                    fileId=file_id,
                    fields="nextPageToken, permissions(id, type, role, emailAddress, domain)",
                    supportsAllDrives=True,
                    pageToken=page_token,
                )
                .execute()
            )

            permissions.extend(response.get("permissions", []))
            page_token = response.get("nextPageToken")

            if not page_token:
                break

        return permissions
    except HttpError as error:
        error_details = error.error_details[0] if error.error_details else {}
        raise ToolError(f"Failed to list permissions, HttpError: {error_details}")


def create_permission(
    service: Resource,
    file_id: str,
    role: str,
    type: str,
    email_address: Optional[str] = None,
    domain: Optional[str] = None,
) -> Optional[dict]:
    """
    Create a new permission for a file.

    Args:
        service: Google Drive API service instance
        file_id: ID of the file to add permission to
        role: The role granted (reader, writer, owner)
        type: The type of grantee (user, group, domain, anyone)
        email_address: Email address for user/group permission
        domain: Domain name for domain permission

    Returns:
        Created permission object if successful, None otherwise
    """
    try:
        permission = {"role": role, "type": type}

        if email_address and type in ["user", "group"]:
            permission["emailAddress"] = email_address
        elif domain and type == "domain":
            permission["domain"] = domain

        return (
            service.permissions()
            .create(
                fileId=file_id,
                body=permission,
                fields="id, type, role, emailAddress, domain",
                supportsAllDrives=True,
                sendNotificationEmail=True,
            )
            .execute()
        )
    except HttpError as error:
        error_details = error.error_details[0] if error.error_details else {}
        raise ToolError(f"Failed to create permission, HttpError: {error_details}")


def update_permission(
    service: Resource, file_id: str, permission_id: str, role: str
) -> Optional[dict]:
    """
    Update an existing permission.

    Args:
        service: Google Drive API service instance
        file_id: ID of the file
        permission_id: ID of the permission to update
        role: The new role to assign (reader, writer, owner)

    Returns:
        Updated permission object if successful, None otherwise
    """
    try:
        return (
            service.permissions()
            .update(
                fileId=file_id,
                permissionId=permission_id,
                body={"role": role},
                fields="id, type, role, emailAddress, domain",
                supportsAllDrives=True,
            )
            .execute()
        )
    except HttpError as error:
        error_details = error.error_details[0] if error.error_details else {}
        raise ToolError(f"Failed to update permission, HttpError: {error_details}")


def delete_permission(service: Resource, file_id: str, permission_id: str) -> bool:
    """
    Delete a permission.

    Args:
        service: Google Drive API service instance
        file_id: ID of the file
        permission_id: ID of the permission to delete

    Returns:
        True if successful, False otherwise
    """
    try:
        service.permissions().delete(
            fileId=file_id, permissionId=permission_id, supportsAllDrives=True
        ).execute()
        return True
    except HttpError as error:
        error_details = error.error_details[0] if error.error_details else {}
        raise ToolError(f"Failed to delete permission, HttpError: {error_details}")


def get_permission(
    service: Resource, file_id: str, permission_id: str
) -> Optional[dict]:
    """
    Get a specific permission.

    Args:
        service: Google Drive API service instance
        file_id: ID of the file
        permission_id: ID of the permission to retrieve

    Returns:
        Permission object if found, None otherwise
    """
    try:
        return (
            service.permissions()
            .get(
                fileId=file_id,
                permissionId=permission_id,
                fields="id, type, role, emailAddress, domain",
                supportsAllDrives=True,
            )
            .execute()
        )
    except HttpError as error:
        error_details = error.error_details[0] if error.error_details else {}
        raise ToolError(f"Failed to get permission, HttpError: {error_details}")


def transfer_ownership(
    service: Resource, file_id: str, new_owner_email: str
) -> Optional[dict]:
    """
    Transfer ownership of a file to another user.

    Args:
        service: Google Drive API service instance
        file_id: ID of the file
        new_owner_email: Email address of the new owner

    Returns:
        New owner permission object if successful, None otherwise
    """
    try:
        permission = {"role": "owner", "type": "user", "emailAddress": new_owner_email}

        return (
            service.permissions()
            .create(
                fileId=file_id,
                body=permission,
                transferOwnership=True,
                fields="id, type, role, emailAddress",
                supportsAllDrives=True,
            )
            .execute()
        )
    except HttpError as error:
        error_details = error.error_details[0] if error.error_details else {}
        raise ToolError(f"Failed to transfer ownership, HttpError: {error_details}")
