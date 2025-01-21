import sys
import os
import json
from googleapiclient.discovery import build
from auth import client

def main():
    try:
        # Get the title from the DOC_TITLE environment variable or default to 'Untitled Document'
        title = os.getenv('DOC_TITLE', 'Untitled Document')

        # Authenticate and build the Docs API service
        service = client('docs', 'v1')

        # Create a new Google Doc with the specified title
        document = service.documents().create(body={"title": title}).execute()

        # Get the document ID
        document_id = document.get('documentId')

        print(f"New document created with ID: {document_id}")
    except Exception as err:
        sys.stderr.write(f"Error: {err}\n")
        sys.exit(1)

if __name__ == "__main__":
    main()
