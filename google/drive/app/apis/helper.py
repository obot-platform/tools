import sys
from google.oauth2.credentials import Credentials
from googleapiclient.discovery import build
from googleapiclient.errors import HttpError
from fastmcp.exceptions import ToolError
import logging


def setup_logger(name, tool_name: str = "Google Drive Tool"):
    """Setup a logger that writes to sys.stderr. Avoid adding duplicate handlers.

    Args:
        name (str): The name of the logger.

    Returns:
        logging.Logger: The logger.
    """
    logger = logging.getLogger(name)
    logger.setLevel(logging.DEBUG)

    if not logger.handlers:
        stderr_handler = logging.StreamHandler(sys.stderr)
        formatter = logging.Formatter(
            f"[{tool_name} Debugging Log]: %(asctime)s - %(name)s - %(levelname)s - %(message)s"
        )
        stderr_handler.setFormatter(formatter)
        logger.addHandler(stderr_handler)

    return logger


logger = setup_logger(__name__)


def get_client(cred_token: str, service_name: str = "drive", version: str = "v3"):

    creds = Credentials(token=cred_token)
    try:
        service = build(serviceName=service_name, version=version, credentials=creds)
        return service
    except HttpError as err:
        raise ToolError(f"HttpError retrieving google {service_name} client: {err}")


def get_user_timezone(service):
    """Fetches the authenticated user's time zone from User's Google Calendar settings."""
    try:
        settings = service.settings().get(setting="timezone").execute()
        return settings.get(
            "value", "UTC"
        )  # Default to UTC if not found
    except HttpError as err:
        if err.status_code == 403:
            raise ToolError(f"HttpError retrieving user timezone: {err}")
        logger.error(f"HttpError retrieving user timezone: {err}")
        return "UTC"
    except Exception as e:
        logger.error(f"Exception retrieving user timezone: {e}")
        return "UTC"
