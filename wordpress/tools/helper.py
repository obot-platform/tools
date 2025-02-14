import requests
import os
import sys
from requests.auth import HTTPBasicAuth
from datetime import datetime
import gptscript
import asyncio
import logging


def setup_logger(name):
    """Setup a logger that writes to sys.stderr. This will eventually show up in GPTScript's debugging logs.

    Args:
        name (str): The name of the logger.

    Returns:
        logging.Logger: The logger.
    """
    # Create a logger
    logger = logging.getLogger(name)
    logger.setLevel(logging.DEBUG)  # Set the logging level

    # Create a stream handler that writes to sys.stderr
    stderr_handler = logging.StreamHandler(sys.stderr)

    # Create a log formatter
    formatter = logging.Formatter(
        "[WordPress Tool Debugging Log]: %(asctime)s - %(name)s - %(levelname)s - %(message)s"
    )
    stderr_handler.setFormatter(formatter)

    # Add the handler to the logger
    logger.addHandler(stderr_handler)

    return logger


logger = setup_logger(__name__)

if "WORDPRESS_USERNAME" not in os.environ:
    raise ValueError("WORDPRESS_USERNAME is not set")

if "WORDPRESS_PASSWORD" not in os.environ:
    raise ValueError("WORDPRESS_PASSWORD is not set")

if "WORDPRESS_SITE" not in os.environ:
    raise ValueError("WORDPRESS_SITE is not set")

username = os.environ["WORDPRESS_USERNAME"]
password = os.environ["WORDPRESS_PASSWORD"]


WORDPRESS_BASIC_AUTH = HTTPBasicAuth(username, password)

WORDPRESS_SITE = os.environ["WORDPRESS_SITE"]
WORDPRESS_API_PATH = "/wp-json/wp/v2"


def clean_wordpress_site_url(site_url):

    if not site_url.startswith("https://") and not site_url.startswith("http://"):
        print(
            f"Error: Invalid site URL: [{site_url}]. No scheme supplied, must start with protocol, e.g. https:// or http://"
        )
        sys.exit(1)

    site_url = site_url.rstrip("/")
    if site_url.endswith("/wp-json"):
        site_url = site_url.split("/wp-json")[0]

    site_url = site_url + WORDPRESS_API_PATH
    return site_url


WORDPRESS_API_URL = clean_wordpress_site_url(WORDPRESS_SITE)


def is_valid_iso8601(date_string):
    try:
        datetime.fromisoformat(date_string)
        return True
    except ValueError:
        logger.error(f"Invalid ISO 8601 date string: {date_string}")
        return False


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


def load_from_gptscript_workspace(filepath: str) -> bytes:
    try:
        gptscript_client = gptscript.GPTScript()
        wksp_file_path = _prepend_base_path("files", filepath)

        try:
            return asyncio.run(gptscript_client.read_file_in_workspace(wksp_file_path))
        except RuntimeError:  # If there's already an event loop running
            loop = asyncio.get_running_loop()
            return loop.run_until_complete(
                gptscript_client.read_file_in_workspace(wksp_file_path)
            )
    except Exception as e:
        logger.error(
            f"Failed to load file {filepath} from GPTScript workspace. Exception: {e}"
        )
        raise Exception(
            f"Failed to load file {filepath} from GPTScript workspace. Exception: {e}"
        )


def str_to_bool(value):
    """Convert a string to a boolean."""
    return str(value).lower() in ("true", "1", "yes")


def create_session():
    session = requests.Session()
    session.auth = WORDPRESS_BASIC_AUTH
    headers = {
        "User-Agent": "curl",
    }
    session.headers.update(headers)
    return session


class ToolRegistry:
    def __init__(self):
        self._tools = {}

    def _register(self, name, func):
        """
        Registers a tool by the given 'name'.
        Raises a ValueError if a tool with the same name is already registered.
        """
        if name in self._tools:
            raise ValueError(f"Tool '{name}' is already registered.")
        self._tools[name] = func

    def get(self, name):
        """
        Retrieves a registered tool by name.
        Raises a ValueError if the tool is not found.
        """
        if name not in self._tools:
            raise ValueError(f"Tool '{name}' not found.")
        return self._tools[name]

    def list_tools(self):
        """
        Returns a list of all registered tool names.
        """
        return list(self._tools.keys())

    def register(self, name):
        """
        A decorator that automatically registers the decorated function
        under the specified 'name' in the ToolRegistry.
        """

        def wrapper(func):
            self._register(name, func)
            return func

        return wrapper


tool_registry = ToolRegistry()
