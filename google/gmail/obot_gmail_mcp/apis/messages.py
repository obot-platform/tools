import base64
from datetime import datetime
from email import encoders
from email.mime.base import MIMEBase
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText
from typing import Optional

from bs4 import BeautifulSoup
from filetype import guess_mime
from googleapiclient.errors import HttpError

from .helpers import (
    format_query_timestamp,
    extract_message_headers,
    prepend_base_path,
    setup_logger,
)

logger = setup_logger(__name__)

CATEGORY_IDS = {
    "CATEGORY_PERSONAL",
    "CATEGORY_SOCIAL",
    "CATEGORY_PROMOTIONS",
    "CATEGORY_UPDATES",
    "CATEGORY_FORUMS",
}


def modify_message_labels(
    service,
    message_id,
    add_labels: list[str] = None,
    remove_labels: list[str] = None,
    apply_action_to_thread: bool = False,
    archive: Optional[
        bool
    ] = None,  # True = remove 'INBOX', False = add 'INBOX', None = leave unchanged
    mark_as_read: Optional[
        bool
    ] = None,  # True = read, False = unread, None = leave unchanged
    mark_as_starred: Optional[
        bool
    ] = None,  # True = star, False = unstar, None = leave unchanged
    mark_as_important: Optional[
        bool
    ] = None,  # True = important, False = not important, None = leave unchanged
):
    add_labels = set(add_labels or [])
    remove_labels = set(remove_labels or [])

    # Validate UNREAD usage with mark_as_read
    if mark_as_read is not None and (
        "UNREAD" in add_labels or "UNREAD" in remove_labels
    ):
        raise ValueError(
            "Do not include 'UNREAD' in add_labels/remove_labels when using mark_as_read"
        )

    # Validate INBOX usage with archive
    if archive is not None and ("INBOX" in add_labels or "INBOX" in remove_labels):
        raise ValueError(
            "Do not include 'INBOX' in add_labels/remove_labels when using archive flag"
        )

    if mark_as_starred is not None and (
        "STARRED" in add_labels or "STARRED" in remove_labels
    ):
        raise ValueError(
            "Do not include 'STARRED' in add_labels/remove_labels when using mark_as_starred"
        )

    if mark_as_important is not None and (
        "IMPORTANT" in add_labels or "IMPORTANT" in remove_labels
    ):
        raise ValueError(
            "Do not include 'IMPORTANT' in add_labels/remove_labels when using mark_as_important"
        )

    # Apply archive behavior
    if archive is True:
        remove_labels.add("INBOX")
    elif archive is False:
        add_labels.add("INBOX")

    # Apply read/unread behavior
    if mark_as_read is True:
        remove_labels.add("UNREAD")
    elif mark_as_read is False:
        add_labels.add("UNREAD")

    if mark_as_starred is True:
        add_labels.add("STARRED")
    elif mark_as_starred is False:
        remove_labels.add("STARRED")

    if mark_as_important is True:
        add_labels.add("IMPORTANT")
    elif mark_as_important is False:
        remove_labels.add("IMPORTANT")

    # Optional: fail on conflicting labels
    conflict_labels = add_labels & remove_labels
    if conflict_labels:
        raise ValueError(f"Conflicting labelIds in add and remove: {conflict_labels}")

    if not add_labels and not remove_labels:
        raise ValueError(f"Warning: No labels to modify for message {message_id}")

    body = {
        "addLabelIds": list(add_labels),
        "removeLabelIds": list(remove_labels),
    }

    if apply_action_to_thread:
        try:
            thread = get_thread_with_message_id(service, message_id)
            thread_id = thread.get("id")
            message_ids = [msg["id"] for msg in thread.get("messages", [])]
            if message_ids:
                body["ids"] = message_ids
                service.users().messages().batchModify(
                    userId="me", body=body
                ).execute()  # this APIs returns empty response if successful
            else:
                raise ValueError(f"No messages found in thread {thread_id}")
        except Exception as e:
            raise ValueError(
                f"Error applying action to thread for message {message_id}: {e}"
            )

        applied_actions = (
            f"Added Labels: {add_labels} "
            if add_labels
            else "" + f"Removed Labels: {remove_labels}" if remove_labels else ""
        )
        response = f"Successfully applied actions:\n{applied_actions}\nto thread {thread_id} with {len(message_ids)} messages."
        return response
    try:
        response = {
            "Response": service.users()
            .messages()
            .modify(userId="me", id=message_id, body=body)
            .execute()
        }
        if "INBOX" in remove_labels:
            thread = get_thread_with_message_id(service, message_id)
            thread_messages = thread.get("messages", [])
            if len(thread_messages) > 1:
                response["Note"] = (
                    "This message is part of a thread with multiple emails. The thread will stay in your inbox unless you archive all messages in the conversation."
                )
    except HttpError as e:
        raise Exception(f"Error modifying message labels: {e}")
    return response


def get_thread_with_message_id(service, message_id: str):
    message = fetch_email_or_draft(service, message_id, format="metadata")
    thread_id = message.get("threadId")
    if thread_id:
        return (
            service.users()
            .threads()
            .get(userId="me", id=thread_id, format="minimal")
            .execute()
        )
    else:
        raise ValueError(f"No thread found for message {message_id}")


async def create_message_data(
    service,
    to,
    cc,
    bcc,
    subject,
    message_text,
    attachments,
    reply_to_email_id=None,
    reply_all=False,
):
    message = MIMEMultipart()
    message_text_html = message_text.replace("\n", "<br>")
    message.attach(MIMEText(message_text_html, "html"))

    thread_id = None
    if reply_to_email_id:
        # Get the original message and extract sender's email
        existing_message = (
            service.users()
            .messages()
            .get(userId="me", id=reply_to_email_id, format="full")
            .execute()
        )
        payload = existing_message["payload"]
        thread_id = existing_message["threadId"]
        headers = {header["name"]: header["value"] for header in payload["headers"]}

        original_from = headers.get("From", "")  # Sender of the original email
        subject = headers.get("Subject", "")  # Subject line
        original_date = headers.get("Date", "")  # Date of the original email
        original_body_html = extract_email_body(payload)
        reply_html = format_reply_gmail_style(
            original_from, original_date, original_body_html
        )
        final_reply_html = f"<br>{reply_html}"
        message["to"] = headers.get("From", "")
        if reply_all:
            message["cc"] = headers.get("CC", "")
        message.attach(MIMEText(final_reply_html, "html"))
        message["References"] = headers.get("References", "")
        message["In-Reply-To"] = headers.get("Message-ID", "")
    else:
        message["to"] = to
        if cc is not None:
            message["cc"] = cc
    message["subject"] = subject
    if bcc is not None:
        message["bcc"] = bcc

    # Read and attach any workspace files if provided
    # TODO: Commented out for now, will get this back in once we have a workspace solution
    # for filepath in attachments:
    #     try:
    #         # Get the file bytes from the workspace
    #         wksp_file_path = await prepend_base_path("files", filepath)
    #         file_content = await gptscript_client.read_file_in_workspace(wksp_file_path)

    #         # Determine the MIME type and subtype
    #         mime = guess_mime(file_content) or "application/octet-stream"
    #         main_type, sub_type = mime.split("/", 1)

    #         # Create the appropriate MIMEBase object for the attachment
    #         mime_base = MIMEBase(main_type, sub_type)
    #         mime_base.set_payload(file_content)
    #         encoders.encode_base64(mime_base)

    #         # Add header with the file name
    #         mime_base.add_header(
    #             "Content-Disposition",
    #             f'attachment; filename="{filepath.split("/")[-1]}"',
    #         )
    #         message.attach(mime_base)
    #     except Exception as e:
    #         # Raise a new exception with the problematic file path included
    #         raise Exception(f"Error attaching {filepath}: {e}")

    # Encode the message as a base64 string
    raw_message = base64.urlsafe_b64encode(message.as_bytes()).decode("utf-8")
    data = {"raw": raw_message}
    if thread_id:
        data["threadId"] = thread_id
    return data


def list_messages(
    service, query, labels, max_results=100, after=None, before=None
) -> list:
    all_messages = []
    next_page_token = None
    if after:
        query = f"after:{format_query_timestamp(after)} {query}"
    if before:
        query = f"before:{format_query_timestamp(before)} {query}"
    logger.info(f"Query: {query}\nlabels: {labels}")  # log query to server logs
    try:
        while True:
            if next_page_token:
                results = (
                    service.users()
                    .messages()
                    .list(
                        userId="me",
                        q=query,
                        labelIds=labels,
                        pageToken=next_page_token,
                        maxResults=10,
                    )
                    .execute()
                )
            else:
                results = (
                    service.users()
                    .messages()
                    .list(userId="me", q=query, labelIds=labels, maxResults=10)
                    .execute()
                )
            messages = results.get("messages", [])
            if not messages:
                break

            all_messages.extend(messages)
            if max_results is not None and len(all_messages) >= max_results:
                break

            next_page_token = results.get("nextPageToken")
            if not next_page_token:
                break
    except HttpError as err:
        raise Exception(f"Error listing messages: {err}")
    except Exception as e:
        logger.error(f"Error listing messages: {e}")
        raise Exception(f"Error listing messages: {e}")

    return all_messages


def message_to_string(service, message, user_tz: str) -> tuple[str, str]:
    msg = (
        service.users()
        .messages()
        .get(
            userId="me",
            id=message["id"],
            format="metadata",
            metadataHeaders=["From", "Subject"],
        )
        .execute()
    )
    return format_message_metadata(msg, user_tz)


def format_message_metadata(msg, user_tz: str) -> tuple[str, str]:
    msg_id = msg["id"]
    subject, sender, to, cc, bcc, date, label_ids = extract_message_headers(
        msg, user_tz
    )
    read_status = "Read" if "UNREAD" not in label_ids else "Unread"
    category = []
    label_ids_set = set(label_ids)
    if "INBOX" in label_ids_set:  # Category only applies when message is in inbox
        for category_id in CATEGORY_IDS:
            if category_id in label_ids_set:
                category.append(category_id)

    msg_str = f"ID: {msg_id} From: {sender}, Subject: {subject}, To: {to}, CC: {cc}, Bcc: {bcc}, Received: {date}, Read_status: {read_status}"
    if category:
        msg_str += f", Built-in Categories: [{', '.join(category)}]"
    return (
        msg_id,
        msg_str,
    )


def display_list_messages(service, messages: list):
    print("Messages:")
    for message in messages:
        _, msg_str = message_to_string(service, message)
        print(msg_str)


def fetch_email_or_draft(service, obj_id, format="full"):
    try:
        # Try fetching as an email first
        return (
            service.users()
            .messages()
            .get(userId="me", id=obj_id, format=format)
            .execute()
        )
    except HttpError as email_err:
        if email_err.resp.status == 404 or email_err.resp.status == 400:
            # If email not found, try fetching as a draft
            try:
                draft_msg = (
                    service.users().drafts().get(userId="me", id=obj_id).execute()
                )
                return draft_msg["message"]
            except HttpError as draft_err:
                raise Exception(f"Error fetching draft: {draft_err}")
        else:
            raise email_err  # Reraise the error if it's not a 404 (not found)


def extract_email_body(payload):
    """Extracts the email body and formats it as HTML."""
    body_html = None

    if "parts" in payload:  # Multipart email
        for part in payload["parts"]:
            if part["mimeType"] == "text/html":
                body_html = base64.urlsafe_b64decode(part["body"]["data"]).decode(
                    "utf-8"
                )
                break
    else:  # Single-part email
        if payload["mimeType"] == "text/html":
            body_html = base64.urlsafe_b64decode(payload["body"]["data"]).decode(
                "utf-8"
            )

    return body_html


def format_reply_gmail_style(original_from, original_date, original_body_html):
    """Formats the reply in Gmail-style with 'On [date], [sender] wrote:'."""

    if original_body_html is None:
        original_body_html = ""

    # Format date as "Mon, Mar 18, 2024 at 10:30 AM"
    formatted_date = original_date
    try:
        parsed_date = datetime.strptime(original_date, "%a, %d %b %Y %H:%M:%S %z")
        formatted_date = parsed_date.strftime("%a, %b %d, %Y at %I:%M %p")
    except:
        pass
    # Plain text reply formatting

    soup = BeautifulSoup(original_body_html, "html.parser")
    quoted_body_html = f'<blockquote style="border-left:2px solid gray;margin-left:10px;padding-left:10px;">{soup.prettify()}</blockquote>'

    reply_html = f"<br><br>On {formatted_date}, <b>{original_from}</b> wrote:<br>{quoted_body_html}"

    return reply_html


def has_attachment(message):
    def parse_parts(parts):
        for part in parts:
            if part["filename"] and part["body"].get("attachmentId"):
                return True
        return False

    parts = message["payload"].get("parts", [])
    if parts:
        return parse_parts(parts)
    else:
        return False


def get_email_body(message):
    def parse_parts(parts):
        for part in parts:
            mime_type = part["mimeType"]
            if mime_type == "text/plain" or mime_type == "text/html":
                body_data = part["body"]["data"]
                decoded_body = base64.urlsafe_b64decode(body_data).decode("utf-8")
                return decoded_body
            if mime_type == "multipart/alternative" or mime_type == "multipart/mixed":
                return parse_parts(part["parts"])
        return None

    try:
        parts = message["payload"].get("parts", [])
        if parts:
            return parse_parts(parts)
        else:
            body_data = message["payload"]["body"]["data"]
            decoded_body = base64.urlsafe_b64decode(body_data).decode("utf-8")
            return decoded_body
    except Exception as e:
        raise Exception(f"Error while decoding the email body: {e}")
