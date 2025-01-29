import os

from auth import client
from update_doc import update_doc

def main():
    try:
        # Get the title from the DOC_TITLE environment variable or default to 'Untitled Document'
        title = os.getenv('DOC_TITLE', 'Untitled Document')
        doc_content = os.getenv('DOC_CONTENT')
        folder_path = os.getenv('DOC_DRIVE_DIR', '').strip()

        # Authenticate and build the Docs and Drive API services
        docs_service = client('docs', 'v1')

        # Create a new Google Doc with the specified title
        document = docs_service.documents().create(body={"title": title}).execute()

        # Get the document ID
        document_id = document.get('documentId')
        print(f"New document created with ID: {document_id}")

        # Add the content to the document and move it to the specified folder
        update_doc(document_id, doc_content, folder_path)

    except Exception as err:
        print(f"Error: {err}")
        exit(1)

if __name__ == "__main__":
    main()
