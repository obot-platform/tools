import gptscript
from googleapiclient.errors import HttpError
from gptscript.datasets import DatasetElement

from apis.helpers import extract_message_headers
from apis.messages import create_message


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

        try:
            gptscript_client = gptscript.GPTScript()

            elements = []
            if len(all_drafts) == 0:
                print("No drafts found")
                return

            for draft in all_drafts:
                draft_id, draft_str = draft_to_string(service, draft)
                elements.append(
                    DatasetElement(name=draft_id, description="", contents=draft_str)
                )

            dataset_id = await gptscript_client.add_dataset_elements(
                elements, name=f"gmail_drafts"
            )

            print(f"Created dataset with ID {dataset_id} with {len(elements)} drafts")
        except Exception as e:
            print("An error occurred while creating the dataset:", e)

    except HttpError as err:
        print(f"An error occurred: {err}")


def draft_to_string(service, draft):
    draft_id = draft["id"]
    draft_msg = service.users().drafts().get(userId="me", id=draft_id).execute()
    msg = draft_msg["message"]
    subject, sender, to, cc, bcc, date = extract_message_headers(msg)
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
        message = await create_message(
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
        print(f"Draft Id: {draft_response['id']} - Draft updated successfully!")
    except HttpError as error:
        print(f"An error occurred: {error}")
