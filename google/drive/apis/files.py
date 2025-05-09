from typing import Optional, List
from googleapiclient.discovery import Resource
from googleapiclient.errors import HttpError
from googleapiclient.http import MediaIoBaseUpload, MediaIoBaseDownload
from io import BytesIO

# Constants
MAX_DOWNLOAD_SIZE = 100 * 1024 * 1024  # Download file not larger than 100MB
FOLDER_MIME_TYPE = "application/vnd.google-apps.folder"


def list_files(
    service: Resource,
    drive_id: Optional[str] = None,
    parent_id: Optional[str] = None,
    mime_type: Optional[str] = None,
    file_name_contains: Optional[str] = None,
    modified_time_after: Optional[str] = None,
    max_results: Optional[int] = None,
    trashed: Optional[bool] = False,
) -> List[dict]:
    """
    List files accessible to the user, optionally filtered by drive_id and query.

    Args:
        service: Google Drive API service instance
        drive_id: Optional ID of the shared drive to list files from
        parent_id: Optional ID of the parent folder to list files from
        query: Optional search query (see Google Drive API documentation for syntax)
        max_results: Optional maximum number of total results to return

    Returns:
        List of file objects containing file information
    """
    try:
        files = []
        page_token = None

        # Prepare parameters
        params = {
            "pageSize": min(100, max_results) if max_results else 100,
            "fields": "nextPageToken, files(id, name, mimeType, parents, modifiedTime, webViewLink, shared)",
            "pageToken": page_token,
            "supportsAllDrives": True,
        }

        if drive_id:
            params.update(
                {
                    "driveId": drive_id,
                    "includeItemsFromAllDrives": True,
                    "supportsAllDrives": True,
                    "corpora": "drive",
                }
            )

        # Build query conditions list
        query_conditions = ["trashed = true" if trashed else "trashed = false"]

        if parent_id:
            query_conditions.append(f"'{parent_id}' in parents")

        if mime_type:
            query_conditions.append(f"mimeType = '{mime_type}'")

        if file_name_contains:
            query_conditions.append(f"name contains '{file_name_contains}'")

        if modified_time_after:
            query_conditions.append(f"modifiedTime > '{modified_time_after}'")

        # Combine all conditions with AND
        if query_conditions:
            params["q"] = " and ".join(query_conditions)

        while True:
            response = service.files().list(**params).execute()
            files.extend(response.get("files", []))

            # Break if we've reached max_results
            if max_results and len(files) >= max_results:
                files = files[:max_results]
                break

            page_token = response.get("nextPageToken")
            if not page_token:
                break

            params["pageToken"] = page_token

        return files
    except HttpError as error:
        error_details = error.error_details[0] if error.error_details else {}
        print(
            f"An error occurred. Error code: {error.resp.status}, Error message: {error_details}"
        )
        return []


def get_file(
    service: Resource,
    file_id: str,
    fields: str = "id, name, mimeType, parents, modifiedTime, size, createdTime, description, webViewLink, shared, owners, lastModifyingUser, capabilities",
    supports_all_drives: bool = True,
) -> Optional[dict]:
    """
    Get information about a specific file.

    Args:
        service: Google Drive API service instance
        file_id: ID of the file
        fields: Comma-separated list of file fields to return

    Returns:
        File object if found, None otherwise
    """
    try:
        return (
            service.files()
            .get(fileId=file_id, fields=fields, supportsAllDrives=supports_all_drives)
            .execute()
        )
    except HttpError as error:
        error_details = error.error_details[0] if error.error_details else {}
        print(
            f"An error occurred. Error code: {error.resp.status}, Error message: {error_details}"
        )
        return None


# https://developers.google.com/workspace/drive/api/guides/manage-uploads
def _prepare_media_upload(file_content: bytes, mime_type: str) -> MediaIoBaseUpload:
    media = MediaIoBaseUpload(BytesIO(file_content), mimetype=mime_type, resumable=True)
    return media


def create_file(
    service: Resource,
    name: str,
    mime_type: str,
    parent_id: Optional[str] = None,
    file_content: Optional[bytes] = None,
) -> Optional[dict]:
    """
    Create a new file in Google Drive.

    Args:
        service: Google Drive API service instance
        name: Name of the new file
        mime_type: MIME type of the file
        parent_id: Optional ID of the parent folder
        file_content: Optional file content as bytes

    Returns:
        Newly created file object if successful, None otherwise
    """
    try:
        file_metadata = {"name": name, "mimeType": mime_type}

        if parent_id:
            file_metadata["parents"] = [parent_id]

        if file_content:
            return (
                service.files()
                .create(
                    body=file_metadata,
                    media_body=_prepare_media_upload(file_content, mime_type),
                    fields="id, name, mimeType, parents",
                    supportsAllDrives=True,
                )
                .execute()
            )
        else:
            return (
                service.files()
                .create(
                    body=file_metadata,
                    fields="id, name, mimeType, parents",
                    supportsAllDrives=True,
                )
                .execute()
            )
    except HttpError as error:
        error_details = error.error_details[0] if error.error_details else {}
        print(
            f"An error occurred. Error code: {error.resp.status}, Error message: {error_details}"
        )
        return None


def delete_file(service: Resource, file_id: str) -> bool:
    """
    Delete a file from Google Drive.

    Args:
        service: Google Drive API service instance
        file_id: ID of the file to delete

    Returns:
        True if successful, False otherwise
    """
    try:
        service.files().delete(fileId=file_id, supportsAllDrives=True).execute()
        return True
    except HttpError as error:
        error_details = error.error_details[0] if error.error_details else {}
        if error.resp.status == 403:
            reason = error_details.get("reason")
            if reason == "insufficientFilePermissions":
                print(
                    f"Permission denied: You don't have sufficient permissions to delete file {file_id}"
                )
            else:
                print(
                    f"Access denied: Unable to delete file {file_id} (reason: {reason})"
                )
        else:
            print(
                f"An error occurred. Error code: {error.resp.status}, Error message: {error_details}"
            )
        return False


def update_file(
    service: Resource,
    file_id: str,
    new_name: Optional[str] = None,
    new_content: Optional[bytes] = None,
    mime_type: Optional[str] = None,
    new_parent_id: Optional[str] = None,
) -> Optional[dict]:
    """
    Update a file's metadata and/or content.

    Args:
        service: Google Drive API service instance
        file_id: ID of the file to update
        new_name: Optional new name for the file
        new_content: Optional new file content as bytes
        mime_type: Optional MIME type of the file content
        new_parent_id: Optional new parent folder ID

    Returns:
        Updated file object if successful, None otherwise
    """
    try:
        # Get previous parents only if moving the file
        add_parents = None
        remove_parents = None
        if new_parent_id:
            file = get_file(service, file_id, "parents", supports_all_drives=True)
            previous_parents = ",".join(file.get("parents", []))
            add_parents = new_parent_id
            remove_parents = previous_parents

        file_metadata = {}
        if new_name:
            file_metadata["name"] = new_name

        media = None
        if new_content and mime_type:
            media = _prepare_media_upload(new_content, mime_type)

        return (
            service.files()
            .update(
                fileId=file_id,
                body=file_metadata,
                media_body=media,
                addParents=add_parents,
                removeParents=remove_parents,
                fields="id, name, mimeType, parents",
                supportsAllDrives=True,
            )
            .execute()
        )

    except HttpError as error:
        error_details = error.error_details[0] if error.error_details else {}
        print(
            f"An error occurred. Error code: {error.resp.status}, Error message: {error_details}"
        )
        return None


def download_file(service: Resource, file_id: str) -> Optional[bytes]:
    """
    Download a file's content from Google Drive.
    Files larger than 100MB will not be downloaded.

    Args:
        service: Google Drive API service instance
        file_id: ID of the file to download

    Returns:
        File content as bytes if successful, None otherwise
    """
    try:
        # Get the file metadata including size
        file = get_file(service, file_id, "mimeType, size, name")
        if not file:
            return None

        # Check file size. exclude Google Workspace files, those are typically not too large.
        if not file["mimeType"].startswith("application/vnd.google-apps"):
            file_size = int(file.get("size", 0))
            if file_size > MAX_DOWNLOAD_SIZE:
                print(
                    f"File '{file['name']}' is too large ({file_size / (1024 * 1024):.2f}MB). Maximum size is {MAX_DOWNLOAD_SIZE / (1024 * 1024):.0f}MB."
                )
                return None

        # Handle Google Workspace files (Docs, Sheets, Slides, etc.)
        if file["mimeType"].startswith("application/vnd.google-apps"):
            export_formats = {
                "application/vnd.google-apps.document": "application/pdf",
                "application/vnd.google-apps.spreadsheet": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
                "application/vnd.google-apps.presentation": "application/pdf",
            }
            export_mime_type = export_formats.get(file["mimeType"], "application/pdf")

            request = service.files().export_media(
                fileId=file_id, mimeType=export_mime_type
            )
        else:
            # Download regular files
            request = service.files().get_media(fileId=file_id, supportsAllDrives=True)

        file_content = BytesIO()
        downloader = MediaIoBaseDownload(file_content, request)

        done = False
        while not done:
            _, done = downloader.next_chunk()

        return file_content.getvalue()

    except HttpError as error:
        error_details = error.error_details[0] if error.error_details else {}
        print(
            f"An error occurred. Error code: {error.resp.status}, Error message: {error_details}"
        )
        return None


def copy_file(
    service: Resource,
    file_id: str,
    new_name: Optional[str] = None,
    parent_id: Optional[str] = None,
) -> Optional[dict]:
    """
    Create a copy of a file in Google Drive.

    Args:
        service: Google Drive API service instance
        file_id: ID of the file to copy
        new_name: Optional name for the copied file (if None, uses 'Copy of [original name]')
        parent_id: Optional ID of the parent folder for the new copy

    Returns:
        Newly created file object if successful, None otherwise
    """
    try:
        # Prepare file metadata
        file_metadata = {}
        if new_name:
            file_metadata["name"] = new_name
        if parent_id:
            file_metadata["parents"] = [parent_id]

        return (
            service.files()
            .copy(
                fileId=file_id,
                body=file_metadata,
                fields="id, name, mimeType, parents",
                supportsAllDrives=True,
            )
            .execute()
        )

    except HttpError as error:
        error_details = error.error_details[0] if error.error_details else {}
        print(
            f"An error occurred. Error code: {error.resp.status}, Error message: {error_details}"
        )
        return None


def create_folder(
    service: Resource,
    name: str,
    parent_id: Optional[str] = None,
) -> Optional[dict]:
    """
    Create a new folder in Google Drive.

    Args:
        service: Google Drive API service instance
        name: Name of the new folder
        parent_id: Optional ID of the parent folder

    Returns:
        Newly created folder object if successful, None otherwise
    """
    return create_file(
        service=service, name=name, mime_type=FOLDER_MIME_TYPE, parent_id=parent_id
    )
