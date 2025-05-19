import os
from apis.files import download_file
from apis.helper import get_client
from apis.workspace_file import save_to_gptscript_workspace


def download_file_tool() -> None:
    client = get_client("drive", "v3")
    file_id = os.getenv("FILE_ID")

    if not file_id:
        print("Error: FILE_ID environment variable is required but not set")
        return

    content, file_name = download_file(client, file_id)
    if content:
        try:
            save_to_gptscript_workspace(file_name, content)
            print(f"Successfully downloaded file to: {file_name}")
        except Exception as e:
            print(f"Error saving file to workspace: {e}")
    else:
        print("Failed to download file")
