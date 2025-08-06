import os
from .helper import setup_logger
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
