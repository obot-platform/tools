
from google.oauth2.credentials import Credentials
from googleapiclient.discovery import build
from googleapiclient.errors import HttpError
import gspread
from fastmcp.exceptions import ToolError


def get_google_client(token: str, service_name: str = "sheets", version: str = "v4"):
    creds = Credentials(token=token)
    try:
        service = build(serviceName=service_name, version=version, credentials=creds)
        return service
    except HttpError as err:
        raise ToolError(f"Error getting Google client: {err}")


def get_gspread_client(token: str) -> gspread.Client:
    creds = Credentials(token=token)
    try:
        service = gspread.authorize(creds)
        return service
    except HttpError as err:
        raise ToolError(f"Error getting Gspread client: {err}")
