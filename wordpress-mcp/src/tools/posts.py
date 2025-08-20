"""WordPress Posts management tools."""

from typing import Optional, List, Union, Dict, Any
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
    context: str = "view",
    page: int = 1,
    per_page: int = 10,
    author_ids: Optional[str] = None,
    search_query: Optional[str] = None,
    statuses: str = "publish",
    publish_after: Optional[str] = None,
    publish_before: Optional[str] = None,
    modified_after: Optional[str] = None,
    modified_before: Optional[str] = None,
    order: str = "desc",
    categories: Optional[str] = None,
    tags: Optional[str] = None
) -> Dict[str, Any]:
    """List posts in WordPress site and get basic information of each post.
    
    Args:
        context: The context of posts to list (view, embed, edit)
        page: Page number to list (default: 1)
        per_page: Number of posts per page (default: 10)
        author_ids: Comma-separated list of author IDs
        search_query: Limit results to those matching a string
        statuses: Comma-separated list of statuses (default: publish)
        publish_after: ISO 8601 date to filter posts published after
        publish_before: ISO 8601 date to filter posts published before  
        modified_after: ISO 8601 date to filter posts modified after
        modified_before: ISO 8601 date to filter posts modified before
        order: Sort order (asc, desc) (default: desc)
        categories: Comma-separated list of category IDs
        tags: Comma-separated list of tag IDs
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
    post_id: int,
    context: str = "view",
    password: Optional[str] = None
) -> Dict[str, Any]:
    """Retrieve all metadata of a post in WordPress site.
    
    Args:
        post_id: The ID of the post
        context: The context of the post (view, embed, edit)
        password: Password for protected posts
    """
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
    title: str = "",
    content: str = "",
    status: str = "draft",
    comment_status: str = "open",
    sticky: bool = False,
    password: Optional[str] = None,
    slug: Optional[str] = None,
    date: Optional[str] = None,
    featured_media: Optional[int] = None,
    format: str = "standard",
    author_id: Optional[int] = None,
    excerpt: Optional[str] = None,
    ping_status: str = "open",
    categories: Optional[str] = None,
    tags: Optional[str] = None
) -> Dict[str, Any]:
    """Create a post in WordPress site. By default, creates as draft.
    
    Args:
        title: The title of the post
        content: The content of the post (use HTML tags for formatting)
        status: Status of the post (publish, future, draft, pending, private)
        comment_status: Comment status (open, closed)
        sticky: Whether the post is sticky
        password: Password for the post
        slug: URL slug for the post
        date: ISO 8601 date string for publishing
        featured_media: ID of featured media file
        format: Post format (standard, aside, chat, gallery, link, image, quote, status, video, audio)
        author_id: ID of the author
        excerpt: Post excerpt
        ping_status: Ping status (open, closed)
        categories: Comma-separated list of category IDs
        tags: Comma-separated list of tag IDs
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
    post_id: int,
    title: Optional[str] = None,
    content: Optional[str] = None,
    status: Optional[str] = None,
    comment_status: Optional[str] = None,
    sticky: Optional[bool] = None,
    password: Optional[str] = None,
    slug: Optional[str] = None,
    date: Optional[str] = None,
    featured_media: Optional[int] = None,
    format: Optional[str] = None,
    author_id: Optional[int] = None,
    excerpt: Optional[str] = None,
    ping_status: Optional[str] = None,
    categories: Optional[str] = None,
    tags: Optional[str] = None
) -> Dict[str, Any]:
    """Update a post in WordPress site. Only provided fields will be updated.
    
    Args:
        post_id: ID of the post to update
        title: New title for the post
        content: New content for the post (use HTML tags for formatting)
        status: New status (publish, future, draft, pending, private)
        comment_status: New comment status (open, closed)
        sticky: Whether the post should be sticky
        password: New password for the post
        slug: New URL slug
        date: New publication date (ISO 8601 format)
        featured_media: New featured media ID
        format: New post format
        author_id: New author ID
        excerpt: New excerpt
        ping_status: New ping status (open, closed)
        categories: New comma-separated list of category IDs
        tags: New comma-separated list of tag IDs
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
    post_id: int,
    force: bool = False
) -> Dict[str, Any]:
    """Delete a post in WordPress site.
    
    Args:
        post_id: ID of the post to delete
        force: Whether to permanently delete (true) or move to trash (false)
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