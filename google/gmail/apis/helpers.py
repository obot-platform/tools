
import logging
import os
import sys
from zoneinfo import ZoneInfo
from datetime import timezone
import base64
import os
import re
import gptscript
from filetype import guess_mime
from datetime import datetime, timezone, timedelta
from zoneinfo import ZoneInfo
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
from email.mime.base import MIMEBase
from email import encoders
from bs4 import BeautifulSoup
from google.oauth2.credentials import Credentials
from googleapiclient.discovery import build
from googleapiclient.errors import HttpError


def setup_logger(name):
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
            "[Gmail Tool Debugging Log]: %(asctime)s - %(name)s - %(levelname)s - %(message)s"
        )
        stderr_handler.setFormatter(formatter)
        logger.addHandler(stderr_handler)

    return logger

logger = setup_logger(__name__)

def get_user_timezone():
    user_tz = os.getenv("OBOT_USER_TIMEZONE", "UTC").strip()

    try:
        tz = ZoneInfo(user_tz)
    except:
        tz = timezone.utc

    return tz

obot_user_tz = get_user_timezone()



def client(service_name: str, version: str):
    token = os.getenv('GOOGLE_OAUTH_TOKEN')
    if token is None:
        raise ValueError("GOOGLE_OAUTH_TOKEN environment variable is not set")

    creds = Credentials(token=token)
    try:
        service = build(serviceName=service_name, version=version, credentials=creds)
        return service
    except HttpError as err:
        print(err)
        exit(1)

def extract_message_headers(message):
    subject = None
    sender = None
    to = None
    cc = None
    bcc = None
    date = None

    if message is not None:
        for header in message['payload']['headers']:
            if header['name'].lower() == 'subject':
                subject = header['value']
            if header['name'].lower() == 'from':
                sender = header['value']
            if header['name'].lower() == 'to':
                to = header['value']
            if header['name'].lower() == 'cc':
                cc = header['value']
            if header['name'].lower() == 'bcc':
                bcc = header['value']
            date = datetime.fromtimestamp(int(message['internalDate']) / 1000, timezone.utc).astimezone(obot_user_tz).strftime(
                '%Y-%m-%d %H:%M:%S %Z')

    return subject, sender, to, cc, bcc, date


async def prepend_base_path(base_path: str, file_path: str):
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