from tools.helper import ZOOM_API_URL, ACCESS_TOKEN
import requests
import os

def get_user():
    url = f"{ZOOM_API_URL}/users/me"
    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
    }

    response = requests.get(url, headers=headers)
    if response.status_code != 200:
        raise Exception(f"Error fetching user info: {response.json()}")

    return response.json()

