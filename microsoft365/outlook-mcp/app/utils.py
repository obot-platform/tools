"""Utility functions for Outlook Mail operations."""

from typing import List, Dict, Any, Optional
from datetime import datetime
from msgraph.generated.models.message import Message
from msgraph.generated.models.post import Post


def message_to_dict(message: Message) -> Dict[str, Any]:
    """Convert a Message object to a dictionary for JSON serialization."""
    result = {
        "id": message.id,
        "subject": message.subject,
        "is_read": message.is_read,
        "has_attachments": message.has_attachments,
        "parent_folder_id": message.parent_folder_id,
        "importance": str(message.importance) if message.importance else None,
        "received_date_time": (
            message.received_date_time.isoformat()
            if message.received_date_time
            else None
        ),
        "sent_date_time": (
            message.sent_date_time.isoformat() if message.sent_date_time else None
        ),
        "created_date_time": (
            message.created_date_time.isoformat() if message.created_date_time else None
        ),
        "last_modified_date_time": (
            message.last_modified_date_time.isoformat()
            if message.last_modified_date_time
            else None
        ),
        "web_link": message.web_link,
        "conversation_id": message.conversation_id,
        "conversation_index": message.conversation_index,
        "is_draft": message.is_draft,
        "categories": list(message.categories) if message.categories else [],
    }

    # Sender info
    if message.sender and message.sender.email_address:
        result["sender"] = {
            "name": message.sender.email_address.name,
            "address": message.sender.email_address.address,
        }

    # From info (might be different from sender)
    if message.from_ and message.from_.email_address:
        result["from"] = {
            "name": message.from_.email_address.name,
            "address": message.from_.email_address.address,
        }

    # To recipients
    if message.to_recipients:
        result["to_recipients"] = []
        for recipient in message.to_recipients:
            if recipient.email_address:
                result["to_recipients"].append(
                    {
                        "name": recipient.email_address.name,
                        "address": recipient.email_address.address,
                    }
                )

    # CC recipients
    if message.cc_recipients:
        result["cc_recipients"] = []
        for recipient in message.cc_recipients:
            if recipient.email_address:
                result["cc_recipients"].append(
                    {
                        "name": recipient.email_address.name,
                        "address": recipient.email_address.address,
                    }
                )

    # BCC recipients
    if message.bcc_recipients:
        result["bcc_recipients"] = []
        for recipient in message.bcc_recipients:
            if recipient.email_address:
                result["bcc_recipients"].append(
                    {
                        "name": recipient.email_address.name,
                        "address": recipient.email_address.address,
                    }
                )

    # Body content
    if message.body:
        result["body"] = {
            "content_type": (
                str(message.body.content_type) if message.body.content_type else None
            ),
            "content": message.body.content,
        }

    # Unique body (preview)
    if message.unique_body:
        result["unique_body"] = {
            "content_type": (
                str(message.unique_body.content_type)
                if message.unique_body.content_type
                else None
            ),
            "content": message.unique_body.content,
        }

    # Flag info
    if message.flag:
        result["flag"] = {
            "flag_status": (
                str(message.flag.flag_status) if message.flag.flag_status else None
            ),
            "start_date_time": (
                message.flag.start_date_time.isoformat()
                if message.flag.start_date_time
                else None
            ),
            "due_date_time": (
                message.flag.due_date_time.isoformat()
                if message.flag.due_date_time
                else None
            ),
            "completed_date_time": (
                message.flag.completed_date_time.isoformat()
                if message.flag.completed_date_time
                else None
            ),
        }

    return result


async def message_to_string(message: Message, include_body: bool = False) -> str:
    """Convert a message to a string representation."""
    lines = []

    # Basic message info
    lines.append(f"ID: {message.id}")
    lines.append(f"Subject: {message.subject or 'No subject'}")

    # Sender info
    if message.sender and message.sender.email_address:
        sender_name = message.sender.email_address.name or "Unknown"
        sender_email = message.sender.email_address.address or "Unknown"
        lines.append(f"From: {sender_name} <{sender_email}>")

    # Recipients
    if message.to_recipients:
        to_list = []
        for recipient in message.to_recipients:
            if recipient.email_address:
                name = recipient.email_address.name or ""
                email = recipient.email_address.address or ""
                if name:
                    to_list.append(f"{name} <{email}>")
                else:
                    to_list.append(email)
        if to_list:
            lines.append(f"To: {', '.join(to_list)}")

    # CC recipients
    if message.cc_recipients:
        cc_list = []
        for recipient in message.cc_recipients:
            if recipient.email_address:
                name = recipient.email_address.name or ""
                email = recipient.email_address.address or ""
                if name:
                    cc_list.append(f"{name} <{email}>")
                else:
                    cc_list.append(email)
        if cc_list:
            lines.append(f"CC: {', '.join(cc_list)}")

    # Date info
    if message.received_date_time:
        lines.append(f"Received: {message.received_date_time.isoformat()}")

    # Status info
    lines.append(f"Read: {'Yes' if message.is_read else 'No'}")
    lines.append(f"Has Attachments: {'Yes' if message.has_attachments else 'No'}")

    # Folder info
    if message.parent_folder_id:
        lines.append(f"Folder ID: {message.parent_folder_id}")

    # Importance and priority
    if message.importance:
        lines.append(f"Importance: {message.importance}")

    # Body content
    if include_body and message.body:
        lines.append("")
        lines.append("Body:")
        lines.append("-" * 40)
        if message.body.content:
            # Strip HTML tags for plain text display
            body_content = strip_html_tags(message.body.content)
            lines.append(body_content)
        lines.append("-" * 40)

    return "\n".join(lines)


async def post_to_string(post: Post, include_body: bool = False) -> str:
    """Convert a post to a string representation."""
    lines = []

    # Basic post info
    lines.append(f"Post ID: {post.id}")

    # Sender info
    if post.sender and post.sender.email_address:
        sender_name = post.sender.email_address.name or "Unknown"
        sender_email = post.sender.email_address.address or "Unknown"
        lines.append(f"From: {sender_name} <{sender_email}>")

    # Recipients (if any)
    if post.new_participants:
        to_list = []
        for recipient in post.new_participants:
            if recipient.email_address:
                name = recipient.email_address.name or ""
                email = recipient.email_address.address or ""
                if name:
                    to_list.append(f"{name} <{email}>")
                else:
                    to_list.append(email)
        if to_list:
            lines.append(f"Additional Recipients: {', '.join(to_list)}")

    # Date info
    if post.created_date_time:
        lines.append(f"Created: {post.created_date_time.isoformat()}")
    if post.last_modified_date_time:
        lines.append(f"Modified: {post.last_modified_date_time.isoformat()}")

    # Attachments
    lines.append(f"Has Attachments: {'Yes' if post.has_attachments else 'No'}")

    # Body content
    if include_body and post.body:
        lines.append("")
        lines.append("Body:")
        lines.append("-" * 40)
        if post.body.content:
            # Strip HTML tags for plain text display
            body_content = strip_html_tags(post.body.content)
            lines.append(body_content)
        lines.append("-" * 40)

    return "\n".join(lines)


def strip_html_tags(html_content: str) -> str:
    """Strip HTML tags from content for plain text display."""
    import re

    # Simple HTML tag removal
    clean = re.compile("<.*?>")
    return re.sub(clean, "", html_content)


async def create_dataset_from_messages(
    messages: List[Message], name: str, description: str
) -> str:
    """Create a dataset from messages."""
    elements = []

    for message in messages:
        message_str = await message_to_string(message, include_body=False)
        elements.append(
            {
                "name": message.id or "",
                "description": message.subject or "No subject",
                "contents": message_str,
            }
        )

    # In a full implementation, this would use the GPTScript client
    # For now, we'll just return a mock dataset ID
    dataset_id = f"dataset_{hash(name)}_{len(elements)}"
    return dataset_id


def format_datetime(dt: Optional[datetime]) -> str:
    """Format datetime for display."""
    if dt is None:
        return "Unknown"
    return dt.isoformat()


def safe_get_attr(obj: Any, attr: str, default: Any = None) -> Any:
    """Safely get attribute from object."""
    try:
        return getattr(obj, attr, default)
    except (AttributeError, TypeError):
        return default


def truncate_text(text: str, max_length: int = 100) -> str:
    """Truncate text to a maximum length."""
    if len(text) <= max_length:
        return text
    return text[: max_length - 3] + "..."


def format_file_size(size_bytes: Optional[int]) -> str:
    """Format file size in human readable format."""
    if size_bytes is None:
        return "Unknown size"

    if size_bytes == 0:
        return "0 bytes"

    size_names = ["bytes", "KB", "MB", "GB", "TB"]
    i = 0
    while size_bytes >= 1024 and i < len(size_names) - 1:
        size_bytes /= 1024.0
        i += 1

    return f"{size_bytes:.1f} {size_names[i]}"


def extract_email_address(email_str: str) -> str:
    """Extract email address from a string that might contain name and email."""
    import re

    # Look for email pattern
    email_pattern = r"\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b"
    match = re.search(email_pattern, email_str)

    if match:
        return match.group()

    # If no email pattern found, return the original string
    return email_str.strip()


def validate_email_address(email: str) -> bool:
    """Validate email address format."""
    import re

    pattern = r"^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$"
    return re.match(pattern, email) is not None


def clean_subject(subject: str) -> str:
    """Clean email subject by removing common prefixes."""
    if not subject:
        return ""

    # Remove common prefixes
    prefixes = ["RE:", "FW:", "FWD:", "RE :", "FW :", "FWD :"]

    cleaned = subject.strip()
    for prefix in prefixes:
        if cleaned.upper().startswith(prefix):
            cleaned = cleaned[len(prefix) :].strip()

    return cleaned


def parse_attachment_name(attachment_name: str) -> tuple[str, str]:
    """Parse attachment name to get filename and extension."""
    if not attachment_name:
        return "unknown", ""

    parts = attachment_name.rsplit(".", 1)
    if len(parts) == 2:
        return parts[0], parts[1].lower()
    else:
        return attachment_name, ""
