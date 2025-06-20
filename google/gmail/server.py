from fastmcp import FastMCP
from pydantic import Field
from typing import Annotated, Literal, Union
import os
from apis.helpers import get_client, parse_label_ids, NON_PRIMARY_CATEGORIES_MAP
from apis.messages import list_messages, message_to_string, modify_message_labels, create_message_data, fetch_email_or_draft, get_email_body, has_attachment, format_message_metadata
from googleapiclient.errors import HttpError
from fastmcp.exceptions import ToolError
from apis.drafts import list_drafts, update_draft
from apis.labels import list_labels, get_label, create_label, update_label, delete_label

# Configure server-specific settings
PORT = os.getenv("PORT", 9000)
MCP_PATH = os.getenv("MCP_PATH", "/mcp/gmail")

mcp = FastMCP(
    name="GmailMCPServer",
    on_duplicate_tools="error",                  # Handle duplicate registrations
    on_duplicate_resources="warn",
    on_duplicate_prompts="replace",
)

@mcp.tool(
    exclude_args=["cred_token", "user_timezone"],
)
async def list_emails(
    max_results: Annotated[int, Field(description="Maximum number of emails to return.", ge=1, le=1000)] = 100,
    query: Annotated[str, Field(description="Query to search for emails.")] = "",
    label_ids: Annotated[str, Field(description="Comma-separated list of label IDs to filter emails by.")] = None,
    category: Annotated[Literal["primary", "social", "promotions", "updates", "forums"], Field(description="Category to filter emails by.")] = "primary",
    after: Annotated[str, Field(description="Date to search for emails after.")] = "",
    before: Annotated[str, Field(description="Date to search for emails before.")] = "",
    user_timezone: str = "UTC",
    cred_token: str = None) -> Union[list[str], str]:
    """
    List emails in the user's gmail account.
    If query is empty, list emails in the user's inbox.
    Otherwise, list emails matching the given query from all labels.
    Supports filtering by labels, category, and query.
    """

    if "after:" in query or "before:" in query:
        raise ValueError(
            "Error: Please use the parameters `after` and `before` instead of having them in the `query`."
        )
    default_inbox = "INBOX"
    if query != "":
        default_inbox = ""  # if query is not empty, don't set inbox as default
    labels = label_ids or default_inbox
    label_ids = parse_label_ids(labels)

    main_query = query
    is_primary_category = False
    if any(
        label.upper() == "ALL" for label in label_ids
    ):  # check if ALL is in the label_ids
        label_ids = []
    elif "INBOX" in label_ids:
        if category in NON_PRIMARY_CATEGORIES_MAP:
            label_ids.append(
                NON_PRIMARY_CATEGORIES_MAP[category]
            )  # we use the internal category names for non-primary categories
        else:  # handle primary categories separately. use query to mimic gmail UI behavior
            main_query = f"{query} category:{category.lower()}"
            is_primary_category = True

    service = get_client(cred_token)
    response = list_messages(service, main_query, label_ids, max_results, after, before)
    if len(response) > 0:
        formatted_response = []
        for message in response:
            formatted_response.append(message_to_string(service, message, user_timezone)[1])
        return formatted_response

    # If not primary category or no results found, we can exit early
    if not is_primary_category:
        return "No emails found"

    # For primary category, ESTIMATE if the feature is enabled
    estimate_response = list_messages(
        service, "category:primary", ["INBOX"], 10, "", ""
    )
    if len(estimate_response) > 0:
        return "No emails found"

    # If categories aren't enabled, try without category filter
    no_category_response = list_messages(
        service, query, label_ids, max_results, after, before
    )
    if len(no_category_response) > 0:
        formatted_response = []
        for message in no_category_response:
            formatted_response.append(message_to_string(service, message, user_timezone)[1])
        return formatted_response

    return "No emails found"

@mcp.tool(
    exclude_args=["cred_token"],
)
def list_drafts(
    max_results: Annotated[int, Field(description="Maximum number of drafts to return.", ge=1, le=1000)] = 100,
    cred_token: str = None
) -> list:
    """
    List drafts in the user's gmail account.
    """
    service = get_client(cred_token)
    drafts = list_drafts(service, max_results)
    return drafts

@mcp.tool(
        annotations={
        "readOnlyHint": True,
        "destructiveHint": False,
    },
    exclude_args=["cred_token"],
)
def list_labels(
    label_id: Annotated[str, Field(description="Label ID to fetch (optional)")] = None,
    cred_token: str = None
) -> list[dict]:
    """
    Fetch a specific label by ID if provided, otherwise list all labels.
    """
    service = get_client(cred_token)
    if label_id:
        label = get_label(service, label_id)
        return [label]
    else:
        labels = list_labels(service)
        custom_labels = [l for l in labels if l.get("type") == "user"]
        return custom_labels

@mcp.tool(
    exclude_args=["cred_token"],
)
def create_label(
    label_name: Annotated[str, Field(description="Name of the label to create.")],
    label_list_visibility: Annotated[Literal["labelShow", "labelHide", "labelShowIfUnread"], Field(description="Label list visibility")] = "labelShow",
    message_list_visibility: Annotated[Literal["show", "hide"], Field(description="Message list visibility")] = "show",
    cred_token: str = None
) -> dict:
    """
    Create a new label in the user's gmail account.
    """
    service = get_client(cred_token)
    label = create_label(service, label_name, label_list_visibility, message_list_visibility)
    return label

@mcp.tool(
    exclude_args=["cred_token"],
)
def update_label(
    label_id: Annotated[str, Field(description="ID of the label to update.")],
    label_name: Annotated[str, Field(description="New name for the label")] = None,
    label_list_visibility: Annotated[Literal["labelShow", "labelHide", "labelShowIfUnread"], Field(description="Label list visibility")] = None,
    message_list_visibility: Annotated[Literal["show", "hide"], Field(description="Message list visibility")] = None,
    cred_token: str = None
) -> dict:
    """
    Update an existing label in the user's gmail account.
    """
    service = get_client(cred_token)
    label = update_label(service, label_id, label_name, label_list_visibility, message_list_visibility)
    return label

@mcp.tool(
    exclude_args=["cred_token"],
)
def delete_label(
    label_id: Annotated[str, Field(description="ID of the label to delete.")],
    cred_token: str = None
) -> str:
    """
    Delete a label in the user's gmail account.
    """
    service = get_client(cred_token)
    result = delete_label(service, label_id)
    return result

@mcp.tool(
    exclude_args=["cred_token"],
)
def modify_message_labels(
    email_id: Annotated[str, Field(description="ID of the email message to modify labels for.")],
    add_label_ids: Annotated[list[str], Field(description="List of label IDs to add")] = None,
    remove_label_ids: Annotated[list[str], Field(description="List of label IDs to remove")] = None,
    archive: Annotated[bool, Field(description="Whether to archive the message")] = None,
    mark_as_read: Annotated[bool, Field(description="Whether to mark the message as read")] = None,
    mark_as_starred: Annotated[bool, Field(description="Whether to mark the message as starred")] = None,
    mark_as_important: Annotated[bool, Field(description="Whether to mark the message as important")] = None,
    apply_action_to_thread: Annotated[bool, Field(description="Whether to apply action to the whole thread")] = False,
    cred_token: str = None
) -> dict:
    """
    Modify labels on a Gmail email or on all messages within the same thread. Supports marking an email or the entire thread as read or unread, archiving or unarchiving, starring or unstarring, marking as important or not important, and adding or removing custom labels.
    """
    service = get_client(cred_token)
    add_labels = parse_label_ids(add_label_ids) if add_label_ids else None
    remove_labels = parse_label_ids(remove_label_ids) if remove_label_ids else None
    res = modify_message_labels(
        service,
        email_id,
        add_labels,
        remove_labels,
        apply_action_to_thread,
        archive,
        mark_as_read,
        mark_as_starred,
        mark_as_important,
    )
    return res

@mcp.tool(
    annotations={"readOnlyHint": True, "destructiveHint": False},
    exclude_args=["cred_token"],
)
async def get_current_email_address(
    cred_token: str = None
) -> str:
    """
    Gets the email address of the currently signed in user.
    """
    service = get_client(cred_token)
    try:
        profile = service.users().getProfile(userId="me").execute()
        return profile["emailAddress"]
    except Exception as e:
        raise ToolError(f"Error getting email address: {e}")

@mcp.tool(
    exclude_args=["cred_token"],
)
async def create_draft(
    to_emails: Annotated[str, Field(description="Comma-separated list of email addresses to send the email to.")],
    subject: Annotated[str, Field(description="Subject of the email.")],
    message: Annotated[str, Field(description="Message body of the email.")],
    cc_emails: Annotated[str, Field(description="Comma-separated list of email addresses to cc the email to (Optional)")] = None,
    bcc_emails: Annotated[str, Field(description="Comma-separated list of email addresses to bcc the email to (Optional)")] = None,
    reply_to_email_id: Annotated[str, Field(description="The ID of the email to reply to (Optional)")] = None,
    reply_all: Annotated[bool, Field(description="Whether to reply to all (Optional: Default is false)")] = False,
    # attachments: Annotated[list[str], Field(description="List of workspace file paths to attach to the email (Optional)")] = None, # not supported yet till workspace is implemented
    cred_token: str = None
) -> str:
    """
    Create a draft email in the user's Gmail account.
    """
    service = get_client(cred_token)
    # att_list = [a.strip() for a in attachments if a.strip()]
    try:
        draft_obj = await create_message_data(
            service=service,
            to=to_emails,
            cc=cc_emails,
            bcc=bcc_emails,
            subject=subject,
            message_text=message,
            # attachments=att_list,
            attachments=[],
            reply_to_email_id=reply_to_email_id,
            reply_all=reply_all,
        )
        return f"Draft Id: {draft_obj['id']} - Draft created successfully!"
    except Exception as e:
        raise ToolError(f"Error creating draft: {e}")

@mcp.tool(
    exclude_args=["cred_token"],
)
def delete_draft(
    draft_id: Annotated[str, Field(description="The ID of the draft to delete.")],
    cred_token: str = None
) -> str:
    """
    Delete a draft email in the user's Gmail account.
    """
    service = get_client(cred_token)
    try:
        service.users().drafts().delete(userId="me", id=draft_id).execute()
        return f"Draft Id: {draft_id} deleted successfully!"
    except HttpError as err:
        raise ToolError(str(err))
    except Exception as err:
        raise ToolError(str(err))

@mcp.tool(
    exclude_args=["cred_token"],
)
def delete_email(
    email_id: Annotated[str, Field(description="The ID of the email to delete.")],
    cred_token: str = None
) -> str:
    """
    Delete an email in the user's Gmail account (moves to trash).
    """
    service = get_client(cred_token)
    try:
        service.users().messages().trash(userId="me", id=email_id).execute()
        return f"Email Id: {email_id} deleted successfully!"
    except HttpError as err:
        raise ToolError(str(err))
    except Exception as err:
        raise ToolError(str(err))

@mcp.tool(
    exclude_args=["cred_token"],
)
def read_email(
    email_id: Annotated[str, Field(description="Email or Draft ID to read (Optional: If not provided, email_subject is required)")] = None,
    email_subject: Annotated[str, Field(description="Email subject to read (Optional: If not provided, email_id is required)")] = None,
    user_timezone: str = "UTC",
    cred_token: str = None
) -> dict:
    """
    Read an email or draft from the user's Gmail account.
    """
    service = get_client(cred_token)
    if not email_id and not email_subject:
        raise ToolError("Either email_id or email_subject must be set")
    try:
        if email_subject:
            query = f'subject:"{email_subject}"'
            response = service.users().messages().list(userId="me", q=query).execute()
            if not response or not response.get("messages"):
                raise ToolError(f"No emails found with subject: {email_subject}")
            email_id = response["messages"][0]["id"]
        msg = fetch_email_or_draft(service, email_id)
        body = get_email_body(msg)
        attachment = has_attachment(msg)
        _, metadata_str = format_message_metadata(msg, user_timezone)
        result = {
            "metadata": metadata_str,
            "body": body,
            "has_attachment": bool(attachment),
        }
        if attachment:
            result["link"] = f"https://mail.google.com/mail/u/0/#inbox/{email_id}"
        return result
    except HttpError as err:
        raise ToolError(str(err))
    except Exception as err:
        raise ToolError(str(err))

@mcp.tool(
    exclude_args=["cred_token"],
)
def send_draft(
    draft_id: Annotated[str, Field(description="The ID of the draft email to send.")],
    cred_token: str = None
) -> str:
    """
    Send a draft email in the user's Gmail account.
    """
    service = get_client(cred_token)
    try:
        sent_message = service.users().drafts().send(userId="me", body={"id": draft_id}).execute()
        return f"Draft Id: {draft_id} sent successfully! Message Id: {sent_message['id']}"
    except HttpError as err:
        raise ToolError(str(err))
    except Exception as err:
        raise ToolError(str(err))

@mcp.tool(
    exclude_args=["cred_token"],
)
async def send_email(
    to_emails: Annotated[str, Field(description="Comma-separated list of email addresses to send the email to.")],
    subject: Annotated[str, Field(description="Subject of the email.")],
    message: Annotated[str, Field(description="Message body of the email.")],
    cc_emails: Annotated[str, Field(description="Comma-separated list of email addresses to cc the email to (Optional)")] = None,
    bcc_emails: Annotated[str, Field(description="Comma-separated list of email addresses to bcc the email to (Optional)")] = None,
    # attachments: Annotated[list[str], Field(description="List of workspace file paths to attach to the email (Optional)")] = None, # not supported yet till workspace is implemented
    cred_token: str = None
) -> str:
    """
    Send an email from the user's Gmail account.
    Do not attempt to forward or reply to emails using this tool.
    """
    service = get_client(cred_token)
    # att_list = [a.strip() for a in attachments if a.strip()] if attachments else []
    try:
        message_obj = await create_message_data(
            service=service,
            to=to_emails,
            cc=cc_emails,
            bcc=bcc_emails,
            subject=subject,
            message_text=message,
            # attachments=att_list,
            attachments=[],
        )
        sent_message = service.users().messages().send(userId="me", body=message_obj).execute()
        return f"Message Id: {sent_message['id']} - Message sent successfully!"
    except HttpError as err:
        raise ToolError(str(err))
    except Exception as err:
        raise ToolError(str(err))

@mcp.tool(
    exclude_args=["cred_token"],
)
async def update_draft(
    draft_id: Annotated[str, Field(description="The ID of the draft email to update.")],
    to_emails: Annotated[str, Field(description="Comma-separated list of email addresses to send the email to.")],
    subject: Annotated[str, Field(description="Subject of the email.")],
    message: Annotated[str, Field(description="Message body of the email.")],
    cc_emails: Annotated[str, Field(description="Comma-separated list of email addresses to cc the email to (Optional)")] = None,
    bcc_emails: Annotated[str, Field(description="Comma-separated list of email addresses to bcc the email to (Optional)")] = None,
    reply_to_email_id: Annotated[str, Field(description="The ID of the email to reply to (Optional)")] = None,
    reply_all: Annotated[bool, Field(description="Whether to reply to all (Optional: Default is false)")] = False,
    # attachments: Annotated[list[str], Field(description="List of workspace file paths to attach to the email (Optional)")] = None, # not supported yet till workspace is implemented
    cred_token: str = None
) -> str:
    """
    Update a draft email in the user's Gmail account.
    """
    service = get_client(cred_token)
    # att_list = [a.strip() for a in attachments if a.strip()] if attachments else []
    try:
        update_draft(
            service=service,
            draft_id=draft_id,
            to=to_emails,
            cc=cc_emails,
            bcc=bcc_emails,
            subject=subject,
            body=message,
            # attachments=att_list,
            attachments=[],
            reply_to_email_id=reply_to_email_id,
            reply_all=reply_all,
        )
        return f"Draft Id: {draft_id} updated successfully!"
    except HttpError as err:
        raise ToolError(str(err))
    except Exception as err:
        raise ToolError(str(err))

@mcp.tool(
    annotations={"readOnlyHint": True, "destructiveHint": False},
    exclude_args=["cred_token"],
)
def list_attachments(
    email_id: Annotated[str, Field(description="The ID of the email to list attachments from.")],
    cred_token: str = None
) -> list:
    """
    List attachments in an email from a user's Gmail account.
    """
    service = get_client(cred_token)
    try:
        msg = fetch_email_or_draft(service, email_id)
        if "payload" not in msg:
            return []
        attachments = []
        if "parts" in msg["payload"]:
            for part in msg["payload"]["parts"]:
                if part.get("filename") and part.get("body", {}).get("attachmentId"):
                    attachments.append({
                        "id": part["body"]["attachmentId"],
                        "filename": part["filename"]
                    })
        return attachments
    except HttpError as error:
        raise ToolError(str(error))
    except Exception as error:
        raise ToolError(str(error))

# TODO: tools missing:
# - read_attachment: need supports of something like a gptscript knowledge tool
# - download_attachment: need to support downloading attachments to the workspace

def streamable_http_server():
    """Main entry point for the Gmail MCP server."""
    mcp.run(
        transport="streamable-http", # fixed to streamable-http
        host="0.0.0.0",
        port=PORT,
        path=MCP_PATH,
    )

def stdio_server():
    """Main entry point for the Gmail MCP server."""
    mcp.run()


if __name__ == "__main__":
    streamable_http_server()