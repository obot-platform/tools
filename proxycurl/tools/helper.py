import os
import logging
import sys

API_KEY = os.getenv("PROXYCURL_API_KEY")

if API_KEY is None or API_KEY == "":
    raise Exception("PROXYCURL_API_KEY environment variable is not set")

api_key_headers = {'Authorization': 'Bearer ' + API_KEY}


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
        "[Proxycurl Tool Debugging Log]: %(asctime)s - %(name)s - %(levelname)s - %(message)s"
    )
    stderr_handler.setFormatter(formatter)

    # Add the handler to the logger
    logger.addHandler(stderr_handler)

    return logger


logger = setup_logger(__name__)


def str_to_bool(value):
    """Convert a string to a boolean."""
    return str(value).lower() in ("true", "1", "yes")


def format_url(text: str) -> str:
    """
    Replaces spaces in a string with hyphens for a URL-friendly format.
    """

    return text.replace(" ", "-")


def remove_images_from_profile(data: dict) -> dict:
    """
    Removes images from a get profile result
    """
    fields_to_remove = ["background_cover_image_url", "profile_pic_url"]

    for field in fields_to_remove:
        data.pop(field, None)

    return data


def remove_images_from_search_result(data: dict) -> dict:
    """
    Removes images from a search result
    """
    if "results" not in data:
        print("results not in data")
        return data

    results = data["results"]

    for i in range(len(results)):
        results[i]["profile"] = remove_images_from_profile(results[i]["profile"])

    return data


class ToolRegistry:
    def __init__(self):
        self._tools = {}

    def register(self, name, func):
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

    def decorator(self, name):
        """
        A decorator that automatically registers the decorated function
        under the specified 'name' in the ToolRegistry.
        """

        def wrapper(func):
            self.register(name, func)
            return func

        return wrapper


tool_registry = ToolRegistry()
