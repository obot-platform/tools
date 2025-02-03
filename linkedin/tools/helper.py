import os
import gptscript

ACCESS_TOKEN = os.getenv("LINKEDIN_OAUTH_TOKEN")
if ACCESS_TOKEN is None or ACCESS_TOKEN == "":
    raise Exception("Error: LINKEDIN_OAUTH_TOKEN is not set properly.")


def str_to_bool(value):
    """Convert a string to a boolean."""
    return str(value).lower() in ("true", "1", "yes")


def _prepend_base_path(base_path: str, file_path: str):
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
async def save_to_gptscript_workspace(filepath: str, content: bytes) -> None:
    gptscript_client = gptscript.GPTScript()
    wksp_file_path = _prepend_base_path("files", filepath)
    await gptscript_client.write_file_in_workspace(wksp_file_path, content)


async def load_from_gptscript_workspace(filepath: str) -> bytes:
    gptscript_client = gptscript.GPTScript()
    wksp_file_path = _prepend_base_path("files", filepath)
    file_content = await gptscript_client.read_file_in_workspace(wksp_file_path)
    return file_content
