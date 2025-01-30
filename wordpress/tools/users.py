import os
from tools.helper import WORDPRESS_API_URL, tool_registry


def _format_users_response(response_json: dict):
    keys = ["id", "name", "url", "description", "link", "slug", "avatar_urls"]
    return {key: response_json[key] for key in keys}


@tool_registry.register("GetUser")
def get_user(client):
    user_id = os.environ["USER_ID"]
    url = f"{WORDPRESS_API_URL}/users/{user_id}"
    response = client.get(url)
    if response.status_code >= 200 and response.status_code < 300:
        return _format_users_response(response.json())
    else:
        print(f"Error: {response.status_code}, {response.text}")


@tool_registry.register("ListUsers")
def list_users(client):
    url = f"{WORDPRESS_API_URL}/users"
    response = client.get(url)
    if response.status_code >= 200 and response.status_code < 300:
        return [_format_users_response(user) for user in response.json()]
    else:
        print(f"Error: {response.status_code}, {response.text}")
