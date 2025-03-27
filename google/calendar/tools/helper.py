import sys
from google.oauth2.credentials import Credentials
from googleapiclient.discovery import build
from googleapiclient.errors import HttpError
import os
import logging


def setup_logger(name, tool_name: str = "Google Calendar Tool"):
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
        f"[{tool_name} Debugging Log]: %(asctime)s - %(name)s - %(levelname)s - %(message)s"
    )
    stderr_handler.setFormatter(formatter)

    # Add the handler to the logger
    logger.addHandler(stderr_handler)

    return logger


logger = setup_logger(__name__)


def str_to_bool(value):
    """Convert a string to a boolean."""
    return str(value).lower() in ("true", "1", "yes")


def get_client(service_name: str = "calendar", version: str = "v3"):
    token = os.getenv("GOOGLE_OAUTH_TOKEN")
    if token is None:
        raise ValueError("GOOGLE_OAUTH_TOKEN environment variable is not set")

    creds = Credentials(token=token)
    try:
        service = build(serviceName=service_name, version=version, credentials=creds)
        return service
    except HttpError as err:
        print(err)
        exit(1)


def get_obot_user_timezone():
    return os.getenv("OBOT_USER_TIMEZONE", "UTC").strip()


def get_user_timezone(service):
    """Fetches the authenticated user's time zone from User's Google Calendar settings."""
    try:
        settings = service.settings().get(setting="timezone").execute()
        return settings.get(
            "value", get_obot_user_timezone()
        )  # Default to Obot's user timezone if not found
    except HttpError as err:
        if err.status_code == 403:
            raise Exception(f"HttpError retrieving user timezone: {err}")
        logger.error(f"HttpError retrieving user timezone: {err}")
        return "UTC"
    except Exception as e:
        logger.error(f"Exception retrieving user timezone: {e}")
        return "UTC"
