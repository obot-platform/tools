from typing import Optional, Annotated

from fastmcp import FastMCP
from pydantic import Field
from fastmcp.exceptions import ToolError

from .client import create_client, get_access_token
from .global_config import SCOPES
from .graph import (
    GroupThreadEmailInfo,
    list_groups,
    list_group_threads,
    delete_group_thread,
    list_group_thread_email_attachments,
    create_group_thread_email,
    list_group_thread_posts,
    get_user_type,
)
from .utils import post_to_string

# Create the group mailbox MCP server
group_mcp = FastMCP(
    name="OutlookGroupMCP",
    on_duplicate_tools="error",
    on_duplicate_resources="warn",
    on_duplicate_prompts="replace",
)


@group_mcp.tool(name="list_groups")
async def list_groups_tool() -> dict:
    """Lists all Microsoft 365 groups the user is a member of."""
    try:
        client = create_client(SCOPES, get_access_token())

        try:
            groups = await list_groups(client)
        except Exception as groups_error:
            # If listing groups fails, check user type to provide better error message
            try:
                user_type = await get_user_type(client)
                if user_type.lower() in ["guest", "personal"]:
                    return {
                        "message": f"User has type '{user_type}', which does not have enough permissions to list groups.",
                        "groups": [],
                    }
            except Exception:
                # If we can't get user type, just raise the original error
                pass

            raise ToolError(f"Failed to list groups: {groups_error}")

        if not groups:
            return {"message": "No groups found", "groups": []}

        # Return groups as a simple list
        group_list = []
        for group in groups:
            group_list.append(
                {
                    "id": group.id,
                    "display_name": group.display_name,
                    "description": group.description,
                    "mail": group.mail,
                    "visibility": group.visibility,
                }
            )

        return {"groups": group_list}
    except Exception as e:
        raise ToolError(f"Unexpected error when listing groups: {e}")


@group_mcp.tool(name="list_group_threads")
async def list_group_threads_tool(
    group_id: Annotated[
        str, Field(description="The ID of the group to list threads in.")
    ],
    start: Annotated[
        Optional[str],
        Field(
            description="The RFC3339 formatted start date and time of the time frame to list threads within."
        ),
    ] = None,
    end: Annotated[
        Optional[str],
        Field(
            description="The RFC3339 formatted end date and time of the time frame to list threads within."
        ),
    ] = None,
    limit: Annotated[
        int,
        Field(
            description="The maximum number of threads to return. If unset, returns up to 100 threads."
        ),
    ] = 100,
) -> dict:
    """Lists all group mailbox threads in a Microsoft 365 group. This will also return emails in the threads."""
    try:
        client = create_client(SCOPES, get_access_token())

        threads = await list_group_threads(
            client, group_id=group_id, start=start, end=end, limit=limit
        )

        # Return threads as a simple list
        thread_list = []
        for thread in threads:
            thread_list.append(
                {
                    "id": thread.id,
                    "topic": thread.topic,
                    "has_attachments": thread.has_attachments,
                    "last_delivered_date_time": (
                        thread.last_delivered_date_time.isoformat()
                        if thread.last_delivered_date_time
                        else None
                    ),
                    "unique_senders": thread.unique_senders,
                }
            )

        return {"threads": thread_list}
    except Exception as e:
        raise ToolError(f"Failed to list group threads: {e}")


@group_mcp.tool(name="create_group_thread_email")
async def create_group_thread_email_tool(
    group_id: Annotated[
        str, Field(description="The ID of the group to create the thread email in.")
    ],
    subject: Annotated[str, Field(description="The subject of the email.")],
    body: Annotated[
        str, Field(description="The body of the email in markdown format.")
    ],
    reply_to_thread_id: Annotated[
        Optional[str],
        Field(
            description="The ID of the thread to reply to. If unset, a new thread will be created."
        ),
    ] = None,
    recipients: Annotated[
        Optional[str],
        Field(
            description="The additional recipients to send the email to, must be a comma-separated list of email addresses."
        ),
    ] = None,
    # attachments: Annotated[Optional[str], Field(description="A comma separated list of workspace file paths to attach to the email.")] = None
) -> dict:
    """Compose a group thread email in Outlook that is always sent to the Microsoft 365 group email address."""
    try:
        client = create_client(SCOPES, get_access_token())

        # Parse additional recipients
        recipient_list = []
        if recipients:
            recipient_list = recipients.split(",")

        info = GroupThreadEmailInfo(
            group_id=group_id,
            subject=subject,
            body=body,
            reply_to_thread_id=reply_to_thread_id,
            recipients=recipient_list,
        )

        thread = await create_group_thread_email(client, info)

        if reply_to_thread_id:
            return {
                "message": f"Successfully replied to thread {reply_to_thread_id} in group {group_id}"
            }
        else:
            return {
                "message": f"Successfully created new thread '{subject}' in group {group_id}. Thread ID: {thread.id}"
            }
    except Exception as e:
        raise ToolError(f"Failed to create group thread email: {e}")


@group_mcp.tool(name="delete_group_thread")
async def delete_group_thread_tool(
    group_id: Annotated[
        str, Field(description="The ID of the group to delete the thread from.")
    ],
    thread_id: Annotated[str, Field(description="The ID of the thread to delete.")],
) -> dict:
    """Delete a group mailbox thread in Outlook."""
    try:
        client = create_client(SCOPES, get_access_token())
        await delete_group_thread(client, group_id, thread_id)
        return {"message": f"Group thread {thread_id} deleted successfully"}
    except Exception as e:
        raise ToolError(f"Failed to delete group thread: {e}")


@group_mcp.tool(name="list_group_thread_email_attachments")
async def list_group_thread_email_attachments_tool(
    group_id: Annotated[
        str, Field(description="The ID of the group containing the thread.")
    ],
    thread_id: Annotated[
        str, Field(description="The ID of the thread containing the email.")
    ],
    email_id: Annotated[
        str, Field(description="The ID of the email to list attachments for.")
    ],
) -> dict:
    """Lists all attachments in a specific Outlook group thread email."""
    try:
        client = create_client(SCOPES, get_access_token())
        attachments = await list_group_thread_email_attachments(
            client, group_id, thread_id, email_id
        )

        attachment_list = []
        for attachment in attachments:
            attachment_list.append(
                {
                    "id": attachment.id,
                    "name": attachment.name,
                    "content_type": attachment.content_type,
                    "size": attachment.size,
                    "is_inline": attachment.is_inline,
                }
            )

        return {"attachments": attachment_list}
    except Exception as e:
        raise ToolError(f"Failed to list group thread email attachments: {e}")


@group_mcp.tool(name="list_group_thread_posts")
async def list_group_thread_posts_tool(
    group_id: Annotated[
        str, Field(description="The ID of the group containing the thread.")
    ],
    thread_id: Annotated[
        str, Field(description="The ID of the thread to list posts from.")
    ],
) -> dict:
    """Lists all posts in a specific Outlook group thread."""
    try:
        client = create_client(SCOPES, get_access_token())
        posts = await list_group_thread_posts(client, group_id, thread_id)

        post_list = []
        for post in posts:
            post_str = await post_to_string(post, include_body=False)
            post_list.append(
                {
                    "id": post.id,
                    "details": post_str,
                    "created_date_time": (
                        post.created_date_time.isoformat()
                        if post.created_date_time
                        else None
                    ),
                    "has_attachments": post.has_attachments,
                }
            )

        return {"posts": post_list}
    except Exception as e:
        raise ToolError(f"Failed to list group thread posts: {e}")
