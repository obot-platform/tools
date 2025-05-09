import os
from apis.files import copy_file
from apis.helper import get_client


def copy_file_tool() -> None:
    client = get_client("drive", "v3")
    file_id = os.getenv("FILE_ID")
    new_name = os.getenv("NEW_NAME")  # Optional
    new_parent_id = os.getenv("NEW_PARENT_ID")  # Optional

    if not file_id:
        print("Error: FILE_ID environment variable is required but not set")
        return

    file = copy_file(
        client, file_id=file_id, new_name=new_name, parent_id=new_parent_id
    )
    print(file)
