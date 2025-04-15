import json
import os
from tools.helper import WORDPRESS_API_URL, tool_registry, setup_logger
from typing import Union
import sys

logger = setup_logger(__name__)


def _format_users_response(response_json: Union[dict, list]) -> Union[dict, list]:
    # response is either a list of dict or a single dict
    try:
        if isinstance(response_json, list):
            return [_format_users_response(user) for user in response_json]
        else:
            keys = [
                "id",
                "name",
                "url",
                "description",
                "link",
                "slug",
                "locale",
                "avatar_urls",
                "roles",
                "capabilities",
                "extra_capabilities",
                "registered_date",
            ]
            return {key: response_json[key] for key in keys if key in response_json}
    except Exception as e:
        logger.error(f"Error formatting users response: {e}")
        return response_json


@tool_registry.register("GetUser")
def get_user(client):
    user_id = os.environ["USER_ID"]
    url = f"{WORDPRESS_API_URL}/users/{user_id}"
    response = client.get(url)
    if response.status_code >= 200 and response.status_code < 300:
        return _format_users_response(response.json())
    else:
        print(f"Error: {response.status_code}, {response.text}")


def _get_current_user_profile(client):
    url = f"{WORDPRESS_API_URL}/users/me"
    query_param = {}
    context = os.getenv("CONTEXT", "edit").lower()
    context_enum = {"view", "embed", "edit"}
    if context not in context_enum:
        raise ValueError(
            f"Error: Invalid context: {context}. context must be one of: {context_enum}."
        )
    query_param["context"] = context
    response = client.get(url, params=query_param)
    return response


@tool_registry.register("GetMe")
def get_me(client):
    response = _get_current_user_profile(client)
    if response.status_code == 200:
        return _format_users_response(response.json())
    else:
        print(
            f"Error Get My Profile: {response.status_code}. Error Message: {response.text}"
        )


@tool_registry.register("ValidateCredential")
def validate_credential(client):
    response = _get_current_user_profile(client)
    if response.status_code == 200:
        sys.exit(0)
    else:
        print(json.dumps({"error": response.text}))
        sys.exit(0)


@tool_registry.register("ListUsers")
def list_users(client):
    url = f"{WORDPRESS_API_URL}/users"
    context = os.getenv("CONTEXT", "view").lower()
    context_enum = {"view", "embed", "edit"}
    if context not in context_enum:
        raise ValueError(
            f"Error: Invalid context: {context}. context must be one of: {context_enum}."
        )
    query_param = {"context": context}
    has_published_posts = os.getenv("HAS_PUBLISHED_POSTS", "true").lower()
    query_param["has_published_posts"] = (
        False if has_published_posts == "false" else True
    )
    response = client.get(url, params=query_param)
    if response.status_code == 200:
        return _format_users_response(response.json())
    elif response.status_code == 403:
        print(
            f"Permission denied: {response.status_code}. Error Message: {response.text}"
        )
    elif response.status_code == 401:
        print(
            f"Authentication failed: {response.status_code}. Error Message: {response.text}"
        )
    else:
        print(f"Error: {response.status_code}. Error Message: {response.text}")
