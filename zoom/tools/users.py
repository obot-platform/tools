from tools.helper import ZOOM_API_URL, ACCESS_TOKEN, tool_registry
import requests


@tool_registry.decorator("GetUser")
def get_user():
    url = f"{ZOOM_API_URL}/users/me"
    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
    }

    response = requests.get(url, headers=headers)
    if response.status_code != 200:
        raise Exception(f"Error fetching user info: {response.json()}")

    return response.json()


def get_user_type() -> int:
    user_info = get_user()
    return user_info["type"]
