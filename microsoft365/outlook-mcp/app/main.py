import asyncio
from typing import Annotated, Literal, Optional

from fastmcp import FastMCP
from fastmcp.exceptions import ToolError
from pydantic import Field
from starlette.requests import Request
from starlette.responses import JSONResponse

from .client import create_client, get_access_token
from .global_config import SCOPES
from .graph import (
    DraftInfo,
    create_draft,
    create_draft_reply,
    delete_message,
    get_attachment_content,
    get_group_thread_post,
    get_me,
    get_message_details,
    list_attachments,
    list_mail_folders,
    list_messages,
    move_message,
    search_messages,
    send_draft,
)
from .group_mcp import group_mcp
from .utils import message_to_dict, message_to_string, post_to_string

mcp = FastMCP(
    name="OutlookMailMCP",
    on_duplicate_tools="error",
    on_duplicate_resources="warn",
    on_duplicate_prompts="replace",
)


@mcp.custom_route("/health", methods=["GET"])
async def health_check(request: Request):
    return JSONResponse({"status": "healthy"})


# Server composition - import group tools
async def setup_server():
    """Setup server composition by importing group tools."""
    await mcp.import_server(group_mcp)


@mcp.tool(name="list_mail_folders")
async def list_mail_folders_tool() -> dict:
    """Lists all available Outlook mail folders."""
    try:
        client = create_client(SCOPES, get_access_token())
        folders = await list_mail_folders(client)

        # Return folders as a simple list
        folder_list = []
        for folder in folders:
            folder_list.append(
                {
                    "id": folder.id,
                    "display_name": folder.display_name,
                    "parent_folder_id": folder.parent_folder_id,
                    "child_folder_count": folder.child_folder_count,
                    "unread_item_count": folder.unread_item_count,
                    "total_item_count": folder.total_item_count,
                }
            )

        return {"folders": folder_list}
    except Exception as e:
        raise ToolError(f"Failed to list mail folders: {e}")


@mcp.tool(name="list_emails")
async def list_emails_tool(
    folder_id: Annotated[
        Optional[str],
        Field(
            description="The ID of the folder to list emails in. If unset, lists emails from all folders."
        ),
    ] = None,
    start: Annotated[
        Optional[str],
        Field(
            description="The RFC3339 formatted start date and time of the time frame to list emails within."
        ),
    ] = None,
    end: Annotated[
        Optional[str],
        Field(
            description="The RFC3339 formatted end date and time of the time frame to list emails within."
        ),
    ] = None,
    limit: Annotated[
        int,
        Field(description="The maximum number of emails to return. Default is 10."),
    ] = 10,
    read_status: Annotated[
        Optional[Literal["read", "unread"]],
        Field(
            description="The optional read status of the emails to list, valid values are 'read' or 'unread'. If unset, lists all emails."
        ),
    ] = None,
) -> dict:
    """Lists emails in an Outlook folder."""
    try:
        client = create_client(SCOPES, get_access_token())

        # Parse parameters

        # Parse read status
        is_read = None
        if read_status:
            if read_status.lower() == "read":
                is_read = True
            elif read_status.lower() == "unread":
                is_read = False
            else:
                raise ToolError("read_status must be 'read', 'unread', or empty")

        # Get messages
        messages = await list_messages(
            client,
            folder_id=folder_id,
            start=start,
            end=end,
            limit=limit,
            is_read=is_read,
        )

        # Convert messages to dictionaries for JSON serialization
        message_dicts = [message_to_dict(msg) for msg in messages]

        return {"messages": message_dicts}
    except Exception as e:
        raise ToolError(f"Failed to list emails: {e}")


@mcp.tool(name="get_email_details")
async def get_email_details_tool(
    email_id: Annotated[
        str,
        Field(
            description="The ID of the email to get details for, or the post_id in a group thread."
        ),
    ],
    group_id: Annotated[
        Optional[str],
        Field(
            description="If the email is a post in a group mailbox, the ID of the group mailbox is also required."
        ),
    ] = None,
    thread_id: Annotated[
        Optional[str],
        Field(
            description="If the email is a post in a group mailbox, the ID of the thread is also required."
        ),
    ] = None,
) -> dict:
    """Get the details of an Outlook email, or a post in a group thread."""
    try:
        client = create_client(SCOPES, get_access_token())

        if group_id and thread_id:
            # Handle group mailbox email details (posts in conversation threads)
            post = await get_group_thread_post(client, group_id, thread_id, email_id)
            post_str = await post_to_string(post, include_body=True)
            return {"email_details": post_str}
        else:
            # Handle regular email details
            message = await get_message_details(client, email_id)
            message_str = await message_to_string(message, include_body=True)
            return {"email_details": message_str}
    except Exception as e:
        raise ToolError(f"Failed to get email details: {e}")


@mcp.tool(name="search_emails")
async def search_emails_tool(
    subject: Annotated[
        Optional[str], Field(description="Search query for the subject of the email.")
    ] = None,
    from_address: Annotated[
        Optional[str],
        Field(description="Search query for the email address of the sender."),
    ] = None,
    from_name: Annotated[
        Optional[str], Field(description="Search query for the name of the sender.")
    ] = None,
    folder_id: Annotated[
        Optional[str],
        Field(
            description="The ID of the folder to search in. If unset, will search all folders."
        ),
    ] = None,
    start: Annotated[
        Optional[str],
        Field(
            description="The start date and time of the time frame to search within, in RFC 3339 format."
        ),
    ] = None,
    end: Annotated[
        Optional[str],
        Field(
            description="The end date and time of the time frame to search within, in RFC 3339 format."
        ),
    ] = None,
    limit: Annotated[
        int,
        Field(description="The maximum number of emails to return. Default is 10."),
    ] = 10,
) -> dict:
    """Search for emails in Outlook. At least one of subject, from_address, or from_name must be specified."""
    try:
        client = create_client(SCOPES, get_access_token())

        # Search messages
        messages = await search_messages(
            client,
            subject=subject,
            from_address=from_address,
            from_name=from_name,
            folder_id=folder_id,
            start=start,
            end=end,
            limit=limit,
        )

        # Convert messages to dictionaries for JSON serialization
        message_dicts = [message_to_dict(msg) for msg in messages]

        return {"messages": message_dicts}
    except Exception as e:
        raise ToolError(f"Failed to search emails: {e}")


@mcp.tool(name="create_draft")
async def create_draft_tool(
    subject: Annotated[str, Field(description="The subject of the email.")],
    body: Annotated[
        str, Field(description="The body of the email in markdown format.")
    ],
    recipients: Annotated[
        str,
        Field(
            description="A comma-separated list of email addresses to send the email to. No spaces. Example: person1@example.com,person2@example.com"
        ),
    ],
    cc: Annotated[
        Optional[str],
        Field(
            description="A comma-separated list of email addresses to CC on the email. No spaces. Example: person1@example.com,person2@example.com"
        ),
    ] = None,
    bcc: Annotated[
        Optional[str],
        Field(
            description="A comma-separated list of email addresses to BCC on the email. No spaces. Example: person1@example.com,person2@example.com"
        ),
    ] = None,
    reply_email_id: Annotated[
        Optional[str], Field(description="The ID of the email to reply to.")
    ] = None,
    reply_all: Annotated[
        Optional[bool],
        Field(
            description="Whether to reply to all. If true, CC will be the original email's CC."
        ),
    ] = False,
) -> dict:
    """Create (but do not send) a draft individual Outlook email."""
    try:
        client = create_client(SCOPES, get_access_token())

        info = DraftInfo(
            subject=subject,
            body=body,
            recipients=recipients.split(","),
            cc=cc.split(",") if cc else [],
            bcc=bcc.split(",") if bcc else [],
            attachments=[],
            reply_all=reply_all,
            reply_to_email_id=reply_email_id or "",
        )

        if info.reply_to_email_id:
            draft = await create_draft_reply(client, info)
        else:
            draft = await create_draft(client, info)

        return {"message": f"Draft created successfully. Draft ID: {draft.id}"}
    except Exception as e:
        raise ToolError(f"Failed to create draft: {e}")


@mcp.tool(name="send_draft")
async def send_draft_tool(
    draft_id: Annotated[str, Field(description="The ID of the draft to send.")],
) -> dict:
    """Send an existing draft email in Outlook."""
    try:
        client = create_client(SCOPES, get_access_token())
        await send_draft(client, draft_id)
        return {"message": f"Draft {draft_id} sent successfully"}
    except Exception as e:
        raise ToolError(f"Failed to send draft: {e}")


@mcp.tool(name="delete_email")
async def delete_email_tool(
    email_id: Annotated[
        str,
        Field(
            description="The ID of the email to delete. This is NOT a mail folder ID."
        ),
    ],
) -> dict:
    """Delete an Outlook email."""
    try:
        client = create_client(SCOPES, get_access_token())
        await delete_message(client, email_id)
        return {"message": f"Email {email_id} deleted successfully"}
    except Exception as e:
        raise ToolError(f"Failed to delete email: {e}")


@mcp.tool(name="move_email")
async def move_email_tool(
    email_id: Annotated[str, Field(description="The ID of the email to move.")],
    destination_folder_id: Annotated[
        str, Field(description="The ID of the folder to move the email into.")
    ],
) -> dict:
    """Moves an email to a different Outlook folder."""
    try:
        client = create_client(SCOPES, get_access_token())
        _ = await move_message(client, email_id, destination_folder_id)
        return {"message": f"Email {email_id} moved to folder {destination_folder_id}"}
    except Exception as e:
        raise ToolError(f"Failed to move email: {e}")


@mcp.tool(name="get_my_email_address")
async def get_my_email_address_tool() -> dict:
    """Get the email address of the currently authenticated Outlook user."""
    try:
        client = create_client(SCOPES, get_access_token())
        user = await get_me(client)
        return {"email": user.mail or user.user_principal_name}
    except Exception as e:
        raise ToolError(f"Failed to get email address: {e}")


@mcp.tool(name="list_attachments")
async def list_attachments_tool(
    email_id: Annotated[
        str, Field(description="The ID of the email to list attachments for.")
    ],
) -> dict:
    """List the attachments of an Outlook email."""
    try:
        client = create_client(SCOPES, get_access_token())
        attachments = await list_attachments(client, email_id)

        # Return attachments as a simple list
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
        raise ToolError(f"Failed to list attachments: {e}")


@mcp.tool(name="download_attachment", enabled=False)
async def download_attachment_tool(
    email_id: Annotated[
        str, Field(description="The ID of the email to get the attachment from.")
    ],
    attachment_id: Annotated[
        str, Field(description="The ID of the attachment to get.")
    ],
) -> dict:
    """Download an attachment from an Outlook email into workspace."""
    try:
        client = create_client(SCOPES, get_access_token())

        # Get attachment content
        content = await get_attachment_content(client, email_id, attachment_id)

        # Get attachment details for filename
        attachments = await list_attachments(client, email_id)
        attachment_name = None
        for att in attachments:
            if att.id == attachment_id:
                attachment_name = att.name
                break

        if not attachment_name:
            attachment_name = f"attachment_{attachment_id}"

        # Save to workspace
        with open(attachment_name, "wb") as f:
            f.write(content)

        return {"message": f"Downloaded attachment {attachment_name} to workspace"}
    except Exception as e:
        raise ToolError(f"Failed to download attachment: {e}")


@mcp.tool(name="read_attachment")
async def read_attachment_tool(
    email_id: Annotated[
        str, Field(description="The ID of the email to get the attachment from.")
    ],
    attachment_id: Annotated[
        str, Field(description="The ID of the attachment to get.")
    ],
) -> dict:
    """Get the markdown converted contents of an attachment from an Outlook email."""
    try:
        client = create_client(SCOPES, get_access_token())

        # Get attachment content
        content = await get_attachment_content(client, email_id, attachment_id)
        return {"content": content}
    except Exception as e:
        raise ToolError(f"Failed to read attachment: {e}")


def streamable_http_server():
    """Main entry point for the Gmail MCP server."""
    asyncio.run(setup_server())
    mcp.run(
        transport="streamable-http",  # fixed to streamable-http
        host="0.0.0.0",
        port=9000,
        path="/mcp/outlook",
    )


if __name__ == "__main__":
    streamable_http_server()
