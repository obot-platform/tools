from tools.helper import (
    WORDPRESS_API_URL,
    tool_registry,
    str_to_bool,
    setup_logger,
)
import os
from typing import Union
import json

logger = setup_logger(__name__)


def _format_tag_response(response_json: Union[dict, list]) -> Union[dict, list]:
    # response is either a list of dict or a single dict
    try:
        if isinstance(response_json, list):
            return [_format_tag_response(tag) for tag in response_json]
        else:
            keys = [
                "id",
                "count",
                "description",
                "name",
                "slug",
                "taxonomy",
            ]
            return {key: response_json[key] for key in keys if key in response_json}
    except Exception as e:
        logger.error(f"Error formatting tag response: {e}")
        return response_json


@tool_registry.register("ListTags")
def list_tags(client):
    url = f"{WORDPRESS_API_URL}/tags"
    query_params = {}
    context = os.getenv("CONTEXT", "view").lower()
    context_enum = {"view", "embed", "edit"}
    if context not in context_enum:
        raise ValueError(
            f"Invalid context. Valid context must be one of: {context_enum}"
        )
    query_params["context"] = context
    page = os.getenv("PAGE", 1)
    query_params["page"] = page
    per_page = os.getenv("PER_PAGE", 10)
    query_params["per_page"] = per_page
    search_query = os.getenv("SEARCH_QUERY", None)
    if search_query:
        query_params["search"] = search_query

    order = os.getenv("ORDER", "asc").lower()
    order_enum = ["asc", "desc"]
    if order not in order_enum:
        raise ValueError(
            f"Error: Invalid order: {order}. order must be one of: {order_enum}."
        )
    query_params["order"] = order

    post = os.getenv("POST_ID", None)
    if post:
        query_params["post"] = post

    slug = os.getenv("SLUG", None)
    if slug:
        query_params["slug"] = slug

    response = client.get(url, params=query_params)
    if response.status_code == 200:
        return _format_tag_response(response.json())
    elif response.status_code == 401:
        print(
            f"Authentication failed: {response.status_code}. Error Message: {response.text}"
        )
    elif response.status_code == 403:
        print(
            f"Permission denied: {response.status_code}. Error Message: {response.text}"
        )
    else:
        print(
            f"Failed to list tags. Error: {response.status_code}. Error Message: {response.text}"
        )
        logger.error(
            f"Failed to list tags. Error Code: {response.status_code}. Error Message: {json.dumps(response.text)}",
        )


@tool_registry.register("CreateTag")
def create_tag(client):
    url = f"{WORDPRESS_API_URL}/tags"
    tag_data = {}

    name = os.getenv("NAME", None)
    if not name:
        raise ValueError("NAME environment variable is required.")
    tag_data["name"] = name

    description = os.getenv("DESCRIPTION", None)
    if description:
        tag_data["description"] = description

    slug = os.getenv("SLUG", None)
    if slug:
        tag_data["slug"] = slug

    response = client.post(url, json=tag_data)
    if response.status_code == 201:
        return _format_tag_response(response.json())
    elif response.status_code == 401:
        print(
            f"Authentication failed: {response.status_code}. Error Message: {response.text}"
        )
    elif response.status_code == 403:
        print(
            f"Permission denied: {response.status_code}. Error Message: {response.text}"
        )
    else:
        print(
            f"Failed to create tag. Error: {response.status_code}. Error Message: {response.text}"
        )
        logger.error(
            f"Failed to create tag. Error Code: {response.status_code}. Error Message: {json.dumps(response.text)}",
        )


@tool_registry.register("UpdateTag")
def update_tag(client):
    tag_id = os.getenv("TAG_ID", None)
    if not tag_id:
        raise ValueError("TAG_ID environment variable is required.")
    if not tag_id.isdigit():
        raise ValueError("TAG_ID must be an integer.")

    url = f"{WORDPRESS_API_URL}/tags/{tag_id}"
    tag_data = {}

    name = os.getenv("NAME", None)
    if name:
        tag_data["name"] = name

    description = os.getenv("DESCRIPTION", None)
    if description:
        tag_data["description"] = description

    slug = os.getenv("SLUG", None)
    if slug:
        tag_data["slug"] = slug

    response = client.post(url, json=tag_data)
    if response.status_code == 200:
        return _format_tag_response(response.json())
    elif response.status_code == 401:
        print(
            f"Authentication failed: {response.status_code}. Error Message: {response.text}"
        )
    elif response.status_code == 403:
        print(
            f"Permission denied: {response.status_code}. Error Message: {response.text}"
        )
    elif response.status_code == 404:
        print(f"Tag_id not found: {tag_id}")
    else:
        print(
            f"Failed to update tag. Error: {response.status_code}. Error Message: {response.text}"
        )
        logger.error(
            f"Failed to update tag. Error Code: {response.status_code}. Error Message: {json.dumps(response.text)}",
        )


@tool_registry.register("DeleteTag")
def delete_tag(client):
    tag_id = os.getenv("TAG_ID", None)
    if not tag_id:
        raise ValueError("TAG_ID environment variable is required.")
    if not tag_id.isdigit():
        raise ValueError("TAG_ID must be an integer.")

    url = f"{WORDPRESS_API_URL}/tags/{tag_id}"

    force = str_to_bool(os.getenv("FORCE", "True"))
    params = {"force": force}

    # Make the DELETE request
    response = client.delete(url, params=params)
    if response.status_code == 200:
        return {"message": f"{response.status_code}. Tag {tag_id} deleted successfully"}
    elif response.status_code == 401:
        print(
            f"Authentication failed: {response.status_code}. Error Message: {response.text}"
        )
    elif response.status_code == 403:
        print(
            f"Permission denied: {response.status_code}. Error Message: {response.text}"
        )
    elif response.status_code == 404:
        print(f"Tag_id not found: {tag_id}")
    else:
        print(
            f"Failed to delete tag. Error: {response.status_code}. Error Message: {response.text}"
        )
        logger.error(
            f"Failed to delete tag. Error Code: {response.status_code}. Error Message: {json.dumps(response.text)}",
        )
