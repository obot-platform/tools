from googleapiclient.errors import HttpError

from .helpers import extract_message_headers
from .messages import create_message_data


async def list_drafts(service, max_results=None):
    all_drafts = []
    next_page_token = None
    try:
        while True:
            if next_page_token:
                results = (
                    service.users()
                    .drafts()
                    .list(userId="me", pageToken=next_page_token, maxResults=10)
                    .execute()
                )
            else:
                results = (
                    service.users().drafts().list(userId="me", maxResults=10).execute()
                )

            drafts = results.get("drafts", [])
            if not drafts:
                break

            all_drafts.extend(drafts)
            if max_results is not None and len(all_drafts) >= max_results:
                break

            next_page_token = results.get("nextPageToken")
            if not next_page_token:
                break

        return all_drafts

    except HttpError as err:
        raise HttpError(f"An error occurred: {err}")


def draft_to_string(service, draft, user_tz: str = "UTC"):
    draft_id = draft["id"]
    draft_msg = service.users().drafts().get(userId="me", id=draft_id).execute()
    msg = draft_msg["message"]
    subject, sender, to, cc, bcc, date, _ = extract_message_headers(msg, user_tz)
    return (
        draft_id,
        f"Draft ID: {draft_id}, From: {sender}, Subject: {subject}, To: {to}, CC: {cc}, Bcc: {bcc}, Saved: {date}",
    )


def display_list_drafts(service, drafts: list):
    print("Drafts:")
    for draft in drafts:
        _, draft_str = draft_to_string(service, draft)
        print(draft_str)


async def update_draft(
    service,
    draft_id,
    to,
    cc,
    bcc,
    subject,
    body,
    attachments,
    reply_to_email_id=None,
    reply_all=False,
):
    try:
        message = await create_message_data(
            service=service,
            to=to,
            cc=cc,
            bcc=bcc,
            subject=subject,
            message_text=body,
            attachments=attachments,
            reply_to_email_id=reply_to_email_id,
            reply_all=reply_all,
        )
        updated_draft = {
            "id": draft_id,
            "message": message,
        }

        draft_response = (
            service.users()
            .drafts()
            .update(userId="me", id=draft_id, body=updated_draft)
            .execute()
        )
        return draft_response
    except Exception as e:
        raise Exception(f"Error: An error occurred: {e}")
