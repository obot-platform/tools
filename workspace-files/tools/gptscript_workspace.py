import os
import gptscript
from pathlib import Path
from tools.helper import setup_logger

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


# for gptscript workspace S/L, see https://github.com/gptscript-ai/py-gptscript/blob/main/gptscript/gptscript.py
async def write_file_in_workspace(filepath: str, content: str) -> bool:
    try:
        gptscript_client = gptscript.GPTScript()
        wksp_file_path = _prepend_base_path(filepath, FILES_DIR)
        await gptscript_client.write_file_in_workspace(
            wksp_file_path, content.encode("utf-8")
        )
        return True
    except Exception as e:
        logger.error(f"Failed to write file to GPTScript workspace: {e}")
        return False


async def list_files_in_workspace(directory: str) -> str:
    gptscript_client = gptscript.GPTScript()
    files = await gptscript_client.list_files_in_workspace(
        prefix=str(Path(FILES_DIR) / directory)
    )
    if files is None:
        return ""

    unique_dirs = set()
    for file in files:
        p = str(Path(file).relative_to(FILES_DIR))  # Remove "FILES_DIR/"
        if p is None:
            continue
        parts = p.split("/")
        if len(parts) > 1:
            unique_dirs.add(parts[0] + "/")  # Add top-level dir with "/"
        else:
            unique_dirs.add(parts[0])  # Add filename

    return "\n".join(sorted(unique_dirs))


async def delete_file_in_workspace(filepath: str) -> None:
    gptscript_client = gptscript.GPTScript()
    wksp_file_path = _prepend_base_path(filepath, FILES_DIR)
    await gptscript_client.delete_file_in_workspace(wksp_file_path)


async def read_file_in_workspace(filepath: str) -> bytes:
    gptscript_client = gptscript.GPTScript()
    wksp_file_path = _prepend_base_path(filepath, FILES_DIR)
    file_content: bytes = await gptscript_client.read_file_in_workspace(wksp_file_path)
    return file_content
