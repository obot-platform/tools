import os
from apis.files import list_files
from apis.helper import get_client
from apis.workspace_file import add_files_to_dataset_elements


def list_files_tool() -> None:
    """
    List or search for files in Google Drive.

    Environment Variables:
        DRIVE_ID (optional): ID of the shared drive to search in
        QUERY (optional): Search query using Google Drive query syntax. Examples:
            - name contains 'report'           : Search by name
            - mimeType = 'application/pdf'     : Search by file type
            - modifiedTime > '2024-01-01'     : Search by modification date
            - '1234' in parents               : Search in specific folder
            - trashed = true                  : Search in trash
            - Multiple conditions can be combined with 'and'/'or':
              "name contains 'report' and mimeType = 'application/pdf'"
        MAX_RESULTS (optional): Maximum number of results to return (default: 50)

    Common MIME types:
        - Google Docs    : application/vnd.google-apps.document
        - Google Sheets  : application/vnd.google-apps.spreadsheet
        - Google Slides  : application/vnd.google-apps.presentation
        - PDF           : application/pdf
        - Folder        : application/vnd.google-apps.folder
    """
    client = get_client("drive", "v3")
    drive_id = os.getenv("DRIVE_ID")  # Optional
    parent_id = os.getenv("PARENT_ID")  # Optional
    mime_type = os.getenv("MIME_TYPE")  # Optional
    file_name_contains = os.getenv("FILE_NAME_CONTAINS")  # Optional
    modified_time_after = os.getenv("MODIFIED_TIME_AFTER")  # Optional
    max_results = os.getenv("MAX_RESULTS")

    if max_results:
        if max_results.isdigit():
            max_results = int(max_results)
        else:
            print("MAX_RESULTS is not a valid integer")
            return
    else:
        max_results = 50  # default

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
    add_files_to_dataset_elements(files)  # this will print out the result
