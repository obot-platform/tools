import asyncio
import os

from googleapiclient.errors import HttpError

from apis.helpers import client
from apis.messages import create_message


async def main():
    to_emails = os.getenv("TO_EMAILS")
    if to_emails is None:
        raise ValueError("At least one recipient must be specified with 'to_emails'")

    cc_emails = os.getenv("CC_EMAILS")
    bcc_emails = os.getenv("BCC_EMAILS")
    subject = os.getenv("SUBJECT")
    if subject is None:
        raise ValueError("Email subject must be set")

    message = os.getenv("MESSAGE")
    if message is None:
        raise ValueError("Email message must be set")

    reply_to_email_id = os.getenv("REPLY_TO_EMAIL_ID")
    reply_all = os.getenv("REPLY_ALL") == "true"
    # Filter out empty strings to clean up attachments
    attachments = os.getenv("ATTACHMENTS", "").split(",")
    attachments = [
        attachment.strip() for attachment in attachments if attachment.strip()
    ]

    service = client("gmail", "v1")
    try:
        await create_draft(
            service=service,
            to=to_emails,
            cc=cc_emails,
            bcc=bcc_emails,
            subject=subject,
            body=message,
            attachments=attachments,
            reply_to_email_id=reply_to_email_id,
            reply_all=reply_all,
        )
    except HttpError as err:
        print(err)
    except Exception as err:
        print(err)


async def create_draft(
    service,
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

        draft = {"message": message}

        draft_response = (
            service.users().drafts().create(userId="me", body=draft).execute()
        )
        print(f"Draft Id: {draft_response['id']} - Draft created successfully!")
    except HttpError as error:
        print(f"An error occurred: {error}")


if __name__ == "__main__":
    asyncio.run(main())
