import os
# import gptscript
from pathlib import Path
from apis.helper import setup_logger
# from gptscript.datasets import DatasetElement
import asyncio

logger = setup_logger(__name__)

FILES_DIR = "files"


def _prepend_base_path(file_path: str, base_path: str = FILES_DIR):
    """
    Prepend a base path to a file path if it's not already rooted in the base path.

    Args:
        base_path (str): The base path to prepend.
        file_path (str): The file path to check and modify.

    Returns:
        str: The modified file path with the base path prepended if necessary.

    Examples:
      >>> prepend_base_path("files", "my-file.txt")
      'files/my-file.txt'

      >>> prepend_base_path("files", "files/my-file.txt")
      'files/my-file.txt'

      >>> prepend_base_path("files", "foo/my-file.txt")
      'files/foo/my-file.txt'

      >>> prepend_base_path("files", "bar/files/my-file.txt")
      'files/bar/files/my-file.txt'

      >>> prepend_base_path("files", "files/bar/files/my-file.txt")
      'files/bar/files/my-file.txt'
    """
    # Split the file path into parts for checking
    file_parts = os.path.normpath(file_path).split(os.sep)

    # Check if the base path is already at the root
    if file_parts[0] == base_path:
        return file_path

    # Prepend the base path
    return os.path.join(base_path, file_path)


def _run_async(coroutine):
    """
    Helper function to run coroutines either in a new event loop or in an existing one.

    Args:
        coroutine: The coroutine to execute

    Returns:
        The result of the coroutine execution
    """
    try:
        return asyncio.run(coroutine)
    except RuntimeError:  # If there's already an event loop running
        loop = asyncio.get_running_loop()
        return loop.run_until_complete(coroutine)


# for gptscript workspace S/L, see https://github.com/gptscript-ai/py-gptscript/blob/main/gptscript/gptscript.py
def save_to_gptscript_workspace(filepath: str, content: bytes) -> bool:
    try:
        gptscript_client = gptscript.GPTScript()
        wksp_file_path = _prepend_base_path(filepath, FILES_DIR)
        _run_async(gptscript_client.write_file_in_workspace(wksp_file_path, content))
        return True
    except Exception as e:
        logger.error(f"Failed to write file to GPTScript workspace: {e}")
        return False


def load_from_gptscript_workspace(filepath: str) -> bytes:
    try:
        gptscript_client = gptscript.GPTScript()
        wksp_file_path = _prepend_base_path(filepath, FILES_DIR)
        return _run_async(gptscript_client.read_file_in_workspace(wksp_file_path))
    except Exception as e:
        logger.error(
            f"Failed to load file {filepath} from GPTScript workspace. Exception: {e}"
        )
        raise Exception(
            f"Failed to load file {filepath} from GPTScript workspace. Exception: {e}"
        )


def add_files_to_dataset_elements(response_files: list[dict]) -> None:
    try:
        gptscript_client = gptscript.GPTScript()

        elements = []
        if len(response_files) == 0:
            print("No files found")
            return

        for file in response_files:
            file_id = file.get("id")
            file_name = file.get("name")
            file_content = "\n".join([f"{k}: {v}" for k, v in file.items()])
            elements.append(
                DatasetElement(
                    name=file_id,
                    description=f"list_files_{file_name}",
                    contents=file_content,
                )
            )

        dataset_id = _run_async(
            gptscript_client.add_dataset_elements(
                elements,
                name=f"google_drive_list_files_{file_id}",
                description=f"list of files in Google Drive",
            )
        )

        print(
            f"Created dataset with ID {dataset_id} with {len(elements)} google drive files"
        )
    except Exception as e:
        print("An error occurred while creating the dataset:", e)
