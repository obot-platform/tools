import base64
import os

from googleapiclient.errors import HttpError

from apis.helpers import client
from apis.messages import (
    fetch_email_or_draft,
    get_email_body,
    has_attachment,
    format_message_metadata,
)

def main():
    email_id = os.getenv("EMAIL_ID")
    email_subject = os.getenv("EMAIL_SUBJECT")
    if email_id is None and email_subject is None:
        raise ValueError("Either email_id or email_subject must be set")

    service = client("gmail", "v1")
    try:
        if email_subject is not None:
            query = f'subject:"{email_subject}"'
            response = service.users().messages().list(userId="me", q=query).execute()
            if not response:
                raise ValueError(f"No emails found with subject: {email_subject}")
            email_id = response["messages"][0]["id"]

        msg = fetch_email_or_draft(service, email_id)
        body = get_email_body(msg)
        attachment = has_attachment(msg)

        _, metadata_str = format_message_metadata(msg)
        print(f"Email metadata:\n{metadata_str}")
        print(f"Email body:\n{body}")
        if attachment:
            print("Email has attachment(s)")
            link = "https://mail.google.com/mail/u/0/#inbox/" + email_id
            print(f"Link: {link}")

    except HttpError as err:
        print(err)


if __name__ == "__main__":
    main()
