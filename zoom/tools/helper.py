import requests
from typing import Dict, Any
import os

ACCESS_TOKEN = os.getenv("ZOOM_OAUTH_TOKEN")

if ACCESS_TOKEN is None or ACCESS_TOKEN == "":
    raise Exception("Error: ZOOM_OAUTH_TOKEN is not set properly.")

ZOOM_API_URL = "https://api.zoom.us/v2"

def str_to_bool(value):
    """Convert a string to a boolean."""
    return str(value).lower() in ('true', '1', 'yes')
