import os

from auth import client
from move_doc import move_doc

def main():
    try:
        # Get the title from the DOC_TITLE environment variable or default to 'Untitled Document'
        title = os.getenv('DOC_TITLE', 'Untitled Document')


        # Authenticate and build the Docs and Drive API services
        docs_service = client('docs', 'v1')
        drive_service = client('drive', 'v3')

        # Create a new Google Doc with the specified title
        document = docs_service.documents().create(body={"title": title}).execute()

        # Get the document ID
        document_id = document.get('documentId')
        print(f"New document created with ID: {document_id}")

        # Move the document to the specified folder path
        folder_path = os.getenv('DOC_DRIVE_DIR', '').strip()
        move_doc(drive_service, document_id, folder_path)

    except Exception as err:
        print(f"Error: {err}")
        exit(1)

if __name__ == "__main__":
    main()
