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


def _format_category_response(response_json: Union[dict, list]) -> Union[dict, list]:
    # response is either a list of dict or a single dict
    try:
        if isinstance(response_json, list):
            return [_format_category_response(category) for category in response_json]
        else:
            keys = [
                "id",
                "count",
                "description",
                "name",
                "parent",
                "slug",
                "taxonomy",
            ]
            return {key: response_json[key] for key in keys if key in response_json}
    except Exception as e:
        logger.error(f"Error formatting category response: {e}")
        return response_json


@tool_registry.register("ListCategories")
def list_categories(client):
    url = f"{WORDPRESS_API_URL}/categories"
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

    parent = os.getenv("PARENT_ID", None)
    if parent:
        query_params["parent"] = parent

    post = os.getenv("POST_ID", None)
    if post:
        query_params["post"] = post

    slug = os.getenv("SLUG", None)
    if slug:
        query_params["slug"] = slug

    response = client.get(url, params=query_params)
    if response.status_code == 200:
        return _format_category_response(response.json())
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
            f"Failed to list categories. Error: {response.status_code}. Error Message: {response.text}"
        )
        logger.error(
            f"Failed to list categories. Error Code: {response.status_code}. Error Message: {json.dumps(response.text)}",
        )


def create_category(client):
    url = f"{WORDPRESS_API_URL}/categories"
    category_data = {}

    # Retrieve category details from environment variables
    name = os.getenv("CATEGORY_NAME", None)
    if not name:
        raise ValueError("CATEGORY_NAME environment variable is required.")
    category_data["name"] = name

    description = os.getenv("DESCRIPTION", None)
    if description:
        category_data["description"] = description

    slug = os.getenv("SLUG", None)
    if slug:
        category_data["slug"] = slug

    parent = os.getenv("PARENT_ID", None)
    if parent:
        if not parent.isdigit():
            raise ValueError("PARENT_ID must be an integer.")
        category_data["parent"] = int(parent)

    # Make the POST request to create the category
    response = client.post(url, json=category_data)
    if response.status_code == 201:
        return _format_category_response(response.json())
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
            f"Failed to create category. Error: {response.status_code}. Error Message: {response.text}"
        )
        logger.error(
            f"Failed to create category. Error Code: {response.status_code}. Error Message: {json.dumps(response.text)}",
        )


@tool_registry.register("UpdateCategory")
def update_category(client):
    # Get category ID from environment variable
    category_id = os.getenv("CATEGORY_ID", None)
    if not category_id:
        raise ValueError("CATEGORY_ID environment variable is required.")
    if not category_id.isdigit():
        raise ValueError("CATEGORY_ID must be an integer.")

    url = f"{WORDPRESS_API_URL}/categories/{category_id}"
    category_data = {}

    # Retrieve optional category details from environment variables
    name = os.getenv("NAME", None)
    if name:
        category_data["name"] = name

    description = os.getenv("DESCRIPTION", None)
    if description:
        category_data["description"] = description

    slug = os.getenv("SLUG", None)
    if slug:
        category_data["slug"] = slug

    parent = os.getenv("PARENT_ID", None)
    if parent:
        if not parent.isdigit():
            raise ValueError("PARENT_ID must be an integer.")
        category_data["parent"] = int(parent)

    # Make the POST request to update the category
    response = client.post(url, json=category_data)
    if response.status_code == 200:
        return _format_category_response(response.json())
    elif response.status_code == 401:
        print(
            f"Authentication failed: {response.status_code}. Error Message: {response.text}"
        )
    elif response.status_code == 403:
        print(
            f"Permission denied: {response.status_code}. Error Message: {response.text}"
        )
    elif response.status_code == 404:
        print(f"Category_id not found: {category_id}")
    else:
        print(
            f"Failed to update category. Error: {response.status_code}. Error Message: {response.text}"
        )
        logger.error(
            f"Failed to update category. Error Code: {response.status_code}. Error Message: {json.dumps(response.text)}",
        )


@tool_registry.register("DeleteCategory")
def delete_category(client):
    # Get category ID from environment variable
    category_id = os.getenv("CATEGORY_ID", None)
    if not category_id:
        raise ValueError("CATEGORY_ID environment variable is required.")
    if not category_id.isdigit():
        raise ValueError("CATEGORY_ID must be an integer.")

    url = f"{WORDPRESS_API_URL}/categories/{category_id}"

    # Optional force parameter to reassign posts to default category
    force = str_to_bool(os.getenv("FORCE", "True"))
    params = {"force": force}

    # Make the DELETE request
    response = client.delete(url, params=params)
    if response.status_code == 200:
        return {
            "message": f"{response.status_code}. Category {category_id} deleted successfully"
        }
    elif response.status_code == 401:
        print(
            f"Authentication failed: {response.status_code}. Error Message: {response.text}"
        )
    elif response.status_code == 403:
        print(
            f"Permission denied: {response.status_code}. Error Message: {response.text}"
        )
    elif response.status_code == 404:
        print(f"Category_id not found: {category_id}")
    else:
        print(
            f"Failed to delete category. Error: {response.status_code}. Error Message: {response.text}"
        )
        logger.error(
            f"Failed to create post. Error Code: {response.status_code}. Error Message: {json.dumps(response.text)}",
        )
