import sys
import os
import io

from googleapiclient.http import MediaIoBaseUpload

from auth import client
from id import extract_file_id
from move_doc import move_doc


def update_doc(file_id, doc_content, drive_dir):
    if doc_content:
        drive_service = client('drive', 'v3')

        # Convert Markdown content into an in-memory file
        markdown_file = io.BytesIO(doc_content.encode("utf-8"))

        # Use media upload for Drive import
        media = MediaIoBaseUpload(markdown_file, mimetype="text/markdown", resumable=True)

        # Overwrite the existing Google Doc with imported content
        updated_file = drive_service.files().update(
            fileId=file_id,
            media_body=media,
            body={'mimeType': 'application/vnd.google-apps.document'}
        ).execute()

        print(f"Document replaced successfully using import: https://docs.google.com/document/d/{file_id}")

    # Move the document to the specified folder
    move_doc(drive_service, file_id, drive_dir)


def main():
    try:
        doc_ref = os.getenv('DOC_REF')
        doc_content = os.getenv('DOC_CONTENT')
        drive_dir = os.getenv('DOC_DRIVE_DIR', '').strip()

        if not doc_ref:
            raise ValueError('DOC_REF environment variable is missing or empty')

        if not doc_content:
            raise ValueError('DOC_CONTENT environment variable is missing or empty')

        file_id = extract_file_id(doc_ref)
        update_doc(file_id, doc_content, drive_dir)

    except Exception as err:
        sys.stderr.write(f"Error: {err}\n")
        sys.exit(1)


if __name__ == "__main__":
    main()
