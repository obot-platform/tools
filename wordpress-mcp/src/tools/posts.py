"""WordPress Posts management tools."""

from typing import Optional, List, Union, Dict, Any, Annotated
from urllib.parse import quote
import requests

from src.server import mcp
from src.config import config, is_valid_iso8601, str_to_bool


def _format_posts_response(response_json: Union[dict, list]) -> Union[dict, list]:
    """Format posts response to include only relevant fields."""
    try:
        if isinstance(response_json, list):
            return [_format_posts_response(post) for post in response_json]
        else:
            keys = [
                "id", "date", "date_gmt", "modified", "modified_gmt", 
                "slug", "status", "type", "link", "title", "excerpt", 
                "author", "categories", "tags", "featured_media", "format"
            ]
            return {key: response_json[key] for key in keys if key in response_json}
    except Exception:
        return response_json


@mcp.tool
def list_posts(
    context: Annotated[str, "The context of posts to list (view, embed, edit) - default: view"] = "view",
    page: Annotated[int, "Page number to list - default: 1"] = 1,
    per_page: Annotated[int, "Number of posts per page - default: 10"] = 10,
    author_ids: Annotated[Optional[str], "Comma-separated list of author IDs - default: None"] = None,
    search_query: Annotated[Optional[str], "Limit results to those matching a string - default: None"] = None,
    statuses: Annotated[str, "Comma-separated list of statuses - default: publish"] = "publish",
    publish_after: Annotated[Optional[str], "ISO 8601 date to filter posts published after - default: None"] = None,
    publish_before: Annotated[Optional[str], "ISO 8601 date to filter posts published before - default: None"] = None,
    modified_after: Annotated[Optional[str], "ISO 8601 date to filter posts modified after - default: None"] = None,
    modified_before: Annotated[Optional[str], "ISO 8601 date to filter posts modified before - default: None"] = None,
    order: Annotated[str, "Sort order (asc, desc) - default: desc"] = "desc",
    categories: Annotated[Optional[str], "Comma-separated list of category IDs - default: None"] = None,
    tags: Annotated[Optional[str], "Comma-separated list of tag IDs - default: None"] = None
) -> Dict[str, Any]:
    """List posts in WordPress site and get basic information of each post.
    
    Date parameters must be valid ISO 8601 date strings (YYYY-MM-DDTHH:MM:SS).
    Valid statuses: publish, future, draft, pending, private, trash, auto-draft, inherit.
    """
    # Validate parameters
    if context not in ["view", "embed", "edit"]:
        raise ValueError(f"Invalid context: {context}")
    
    if order not in ["asc", "desc"]:
        raise ValueError(f"Invalid order: {order}")
        
    status_enum = [
        "publish", "future", "draft", "pending", "private", "trash",
        "auto-draft", "inherit", "request-pending", "request-confirmed",
        "request-failed", "request-completed"
    ]
    for status in statuses.split(","):
        if status.strip() not in status_enum:
            raise ValueError(f"Invalid status: {status}")
    
    # Validate date parameters
    for date_param, date_value in [
        ("publish_after", publish_after),
        ("publish_before", publish_before), 
        ("modified_after", modified_after),
        ("modified_before", modified_before)
    ]:
        if date_value and not is_valid_iso8601(date_value):
            raise ValueError(f"Invalid {date_param}: must be ISO 8601 format")
    
    # Validate author_ids
    if author_ids:
        for author_id in author_ids.split(","):
            if not author_id.strip().isdigit():
                raise ValueError(f"Invalid author_id: {author_id}")
    
    # Validate categories and tags
    if categories:
        for cat_id in categories.split(","):
            if not cat_id.strip().isdigit():
                raise ValueError(f"Invalid category ID: {cat_id}")
                
    if tags:
        for tag_id in tags.split(","):
            if not tag_id.strip().isdigit():
                raise ValueError(f"Invalid tag ID: {tag_id}")
    
    # Build query parameters
    params = {
        "context": context,
        "page": page,
        "per_page": per_page,
        "status": statuses,
        "order": order
    }
    
    if author_ids:
        params["author"] = author_ids
    if search_query:
        params["search"] = search_query
    if publish_after:
        params["after"] = quote(publish_after)
    if publish_before:
        params["before"] = quote(publish_before)
    if modified_after:
        params["modified_after"] = quote(modified_after)
    if modified_before:
        params["modified_before"] = quote(modified_before)
    if categories:
        params["categories"] = categories
    if tags:
        params["tags"] = tags
    
    # Make request
    session = config.create_session()
    response = session.get(f"{config.api_url}/posts", params=params)
    
    if response.status_code == 200:
        return {"posts": _format_posts_response(response.json())}
    elif response.status_code == 401:
        raise Exception(f"Authentication failed: {response.text}")
    elif response.status_code in [400, 403]:
        raise Exception(f"Permission denied: {response.text}")
    else:
        raise Exception(f"Failed to list posts: {response.status_code} - {response.text}")


@mcp.tool
def retrieve_post(
    post_id: Annotated[int, "The ID of the post"],
    context: Annotated[str, "The context of the post (view, embed, edit) - default: view"] = "view",
    password: Annotated[Optional[str], "Password for protected posts - default: None"] = None
) -> Dict[str, Any]:
    """Retrieve all metadata of a post in WordPress site."""
    if context not in ["view", "embed", "edit"]:
        raise ValueError(f"Invalid context: {context}")
    
    params = {"context": context}
    if password:
        params["password"] = password
    
    session = config.create_session()
    response = session.get(f"{config.api_url}/posts/{post_id}", params=params)
    
    if response.status_code == 200:
        return response.json()
    elif response.status_code == 401:
        raise Exception(f"Authentication failed: {response.text}")
    elif response.status_code == 403:
        raise Exception(f"Permission denied: {response.text}")
    else:
        raise Exception(f"Failed to retrieve post: {response.status_code} - {response.text}")


@mcp.tool
def create_post(
    title: Annotated[str, "The title of the post - default: empty"] = "",
    content: Annotated[str, "The content of the post (use HTML tags for formatting) - default: empty"] = "",
    status: Annotated[str, "Status of the post (publish, future, draft, pending, private) - default: draft"] = "draft",
    comment_status: Annotated[str, "Comment status (open, closed) - default: open"] = "open",
    sticky: Annotated[bool, "Whether the post is sticky - default: false"] = False,
    password: Annotated[Optional[str], "Password for the post - default: None"] = None,
    slug: Annotated[Optional[str], "URL slug for the post - default: None"] = None,
    date: Annotated[Optional[str], "ISO 8601 date string for publishing - default: None"] = None,
    featured_media: Annotated[Optional[int], "ID of featured media file - default: None"] = None,
    format: Annotated[str, "Post format (standard, aside, chat, gallery, link, image, quote, status, video, audio) - default: standard"] = "standard",
    author_id: Annotated[Optional[int], "ID of the author - default: None"] = None,
    excerpt: Annotated[Optional[str], "Post excerpt - default: None"] = None,
    ping_status: Annotated[str, "Ping status (open, closed) - default: open"] = "open",
    categories: Annotated[Optional[str], "Comma-separated list of category IDs - default: None"] = None,
    tags: Annotated[Optional[str], "Comma-separated list of tag IDs - default: None"] = None
) -> Dict[str, Any]:
    """Create a post in WordPress site. By default, creates as draft.
    
    Use HTML tags for content formatting. At least one of title or content must be provided.
    Date must be ISO 8601 format. Future dates will schedule the post.
    """
    if not title and not content:
        raise ValueError("At least one of title or content must be provided")
    
    # Validate parameters
    if status not in ["publish", "future", "draft", "pending", "private"]:
        raise ValueError(f"Invalid status: {status}")
        
    if comment_status not in ["open", "closed"]:
        raise ValueError(f"Invalid comment_status: {comment_status}")
        
    if ping_status not in ["open", "closed"]:
        raise ValueError(f"Invalid ping_status: {ping_status}")
        
    format_enum = [
        "standard", "aside", "chat", "gallery", "link", "image", 
        "quote", "status", "video", "audio"
    ]
    if format not in format_enum:
        raise ValueError(f"Invalid format: {format}")
    
    if date and not is_valid_iso8601(date):
        raise ValueError("Invalid date: must be ISO 8601 format")
    
    if categories:
        for cat_id in categories.split(","):
            if not cat_id.strip().isdigit():
                raise ValueError(f"Invalid category ID: {cat_id}")
                
    if tags:
        for tag_id in tags.split(","):
            if not tag_id.strip().isdigit():
                raise ValueError(f"Invalid tag ID: {tag_id}")
    
    # Build post data
    post_data = {
        "title": title,
        "content": content,
        "status": status,
        "comment_status": comment_status,
        "sticky": sticky,
        "ping_status": ping_status,
        "format": format
    }
    
    if password:
        post_data["password"] = password
    if slug:
        post_data["slug"] = slug
    if date:
        post_data["date"] = date
    if featured_media:
        post_data["featured_media"] = featured_media
    if author_id:
        post_data["author"] = author_id
    if excerpt:
        post_data["excerpt"] = excerpt
    if categories:
        post_data["categories"] = categories
    if tags:
        post_data["tags"] = tags
    
    session = config.create_session()
    response = session.post(f"{config.api_url}/posts", json=post_data)
    
    if response.status_code == 201:
        return _format_posts_response(response.json())
    elif response.status_code == 401:
        raise Exception(f"Authentication failed: {response.text}")
    elif response.status_code == 403:
        raise Exception(f"Permission denied: {response.text}")
    else:
        raise Exception(f"Failed to create post: {response.status_code} - {response.text}")


@mcp.tool
def update_post(
    post_id: Annotated[int, "ID of the post to update"],
    title: Annotated[Optional[str], "New title for the post - default: None"] = None,
    content: Annotated[Optional[str], "New content for the post (use HTML tags for formatting) - default: None"] = None,
    status: Annotated[Optional[str], "New status (publish, future, draft, pending, private) - default: None"] = None,
    comment_status: Annotated[Optional[str], "New comment status (open, closed) - default: None"] = None,
    sticky: Annotated[Optional[bool], "Whether the post should be sticky - default: None"] = None,
    password: Annotated[Optional[str], "New password for the post - default: None"] = None,
    slug: Annotated[Optional[str], "New URL slug - default: None"] = None,
    date: Annotated[Optional[str], "New publication date (ISO 8601 format) - default: None"] = None,
    featured_media: Annotated[Optional[int], "New featured media ID - default: None"] = None,
    format: Annotated[Optional[str], "New post format - default: None"] = None,
    author_id: Annotated[Optional[int], "New author ID - default: None"] = None,
    excerpt: Annotated[Optional[str], "New excerpt - default: None"] = None,
    ping_status: Annotated[Optional[str], "New ping status (open, closed) - default: None"] = None,
    categories: Annotated[Optional[str], "New comma-separated list of category IDs - default: None"] = None,
    tags: Annotated[Optional[str], "New comma-separated list of tag IDs - default: None"] = None
) -> Dict[str, Any]:
    """Update a post in WordPress site. Only provided fields will be updated.
    
    Use HTML tags for content formatting. Date must be ISO 8601 format.
    """
    # Validate parameters
    if status and status not in ["publish", "future", "draft", "pending", "private"]:
        raise ValueError(f"Invalid status: {status}")
        
    if comment_status and comment_status not in ["open", "closed"]:
        raise ValueError(f"Invalid comment_status: {comment_status}")
        
    if ping_status and ping_status not in ["open", "closed"]:
        raise ValueError(f"Invalid ping_status: {ping_status}")
        
    if format:
        format_enum = [
            "standard", "aside", "chat", "gallery", "link", "image", 
            "quote", "status", "video", "audio"
        ]
        if format not in format_enum:
            raise ValueError(f"Invalid format: {format}")
    
    if date and not is_valid_iso8601(date):
        raise ValueError("Invalid date: must be ISO 8601 format")
    
    if categories:
        for cat_id in categories.split(","):
            if not cat_id.strip().isdigit():
                raise ValueError(f"Invalid category ID: {cat_id}")
                
    if tags:
        for tag_id in tags.split(","):
            if not tag_id.strip().isdigit():
                raise ValueError(f"Invalid tag ID: {tag_id}")
    
    # Build update data (only include provided fields)
    post_data = {}
    
    if title is not None:
        if title == "":
            raise ValueError("Title cannot be empty")
        post_data["title"] = title
        
    if content is not None:
        if content == "":
            raise ValueError("Content cannot be empty")
        post_data["content"] = content
        
    if status is not None:
        post_data["status"] = status
    if comment_status is not None:
        post_data["comment_status"] = comment_status
    if sticky is not None:
        post_data["sticky"] = sticky
    if password is not None:
        post_data["password"] = password
    if slug is not None:
        post_data["slug"] = slug
    if date is not None:
        post_data["date"] = date
    if featured_media is not None:
        post_data["featured_media"] = featured_media
    if format is not None:
        post_data["format"] = format
    if author_id is not None:
        post_data["author"] = author_id
    if excerpt is not None:
        post_data["excerpt"] = excerpt
    if ping_status is not None:
        post_data["ping_status"] = ping_status
    if categories is not None:
        post_data["categories"] = categories
    if tags is not None:
        post_data["tags"] = tags
    
    session = config.create_session()
    response = session.post(f"{config.api_url}/posts/{post_id}", json=post_data)
    
    if response.status_code == 200:
        return _format_posts_response(response.json())
    elif response.status_code == 401:
        raise Exception(f"Authentication failed: {response.text}")
    elif response.status_code == 403:
        raise Exception(f"Permission denied: {response.text}")
    else:
        raise Exception(f"Failed to update post: {response.status_code} - {response.text}")


@mcp.tool
def delete_post(
    post_id: Annotated[int, "ID of the post to delete"],
    force: Annotated[bool, "Whether to permanently delete (true) or move to trash (false) - default: false"] = False
) -> Dict[str, Any]:
    """Delete a post in WordPress site.
    
    If force=false, post is moved to trash. If force=true, post is permanently deleted.
    """
    params = {}
    if force:
        params["force"] = "true"
    
    session = config.create_session()
    response = session.delete(f"{config.api_url}/posts/{post_id}", params=params)
    
    if response.status_code == 200:
        return {
            "message": f"Post {post_id} deleted successfully. "
                      "Note: If this was a published post, it may still appear temporarily "
                      "due to caching. This is normal and should resolve shortly."
        }
    elif response.status_code == 401:
        raise Exception(f"Authentication failed: {response.text}")
    elif response.status_code == 403:
        raise Exception(f"Permission denied: {response.text}")
    else:
        raise Exception(f"Failed to delete post: {response.status_code} - {response.text}")