import os
from tools.helper import WORDPRESS_API_URL, tool_registry
from typing import Union

def _format_users_response(response_json: Union[dict, list]) -> Union[dict, list]:
    # response is either a list of dict or a single dict
    if isinstance(response_json, list):
        return [_format_users_response(user) for user in response_json]
    else:
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


@tool_registry.register("GetMe")
def get_me(client):
    url = f"{WORDPRESS_API_URL}/users/me"
    query_param = {"context": "edit"}
    response = client.get(url, params=query_param)
    return response.json()


@tool_registry.register("ListUsers")
def list_users(client):
    url = f"{WORDPRESS_API_URL}/users"
    response = client.get(url)
    if response.status_code >= 200 and response.status_code < 300:
        return _format_users_response(response.json())
    else:
        print(f"Error: {response.status_code}, {response.text}")
