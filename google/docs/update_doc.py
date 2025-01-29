import sys
import os

import markdown
from bs4 import BeautifulSoup

from auth import client
from id import extract_file_id
from move_doc import move_doc


def markdown_to_google_doc_requests(markdown_content):
    # Convert markdown content to HTML
    html_content = markdown.markdown(markdown_content)
    soup = BeautifulSoup(html_content, 'html.parser')

    requests = []
    current_index = 1

    def add_text_request(text, bold=False, italic=False, underline=False, link=None):
        nonlocal current_index
        # Skip completely empty or whitespace-only values, except for single newlines
        if not text.strip() and text != "\n":
            return

        text_style = {
            "bold": bold,
            "italic": italic,
            "underline": underline,
        }
        if link:
            text_style["link"] = {"url": link}

        text_length = len(text)
        requests.append({
            "insertText": {
                "location": {"index": current_index},
                "text": text
            }
        })

        if text_style or link:
            requests.append({
                "updateTextStyle": {
                    "range": {
                        "startIndex": current_index,
                        "endIndex": current_index + text_length
                    },
                    "textStyle": text_style,
                    "fields": ",".join(text_style.keys())
                }
            })

        current_index += text_length

        # Handle unstyled newlines
        if text.endswith("\n"):
            newline_length = 1
            requests.append({
                "updateTextStyle": {
                    "range": {
                        "startIndex": current_index - newline_length,
                        "endIndex": current_index
                    },
                    "textStyle": {},  # Explicitly remove styles
                    "fields": "bold,italic,underline,link"
                }
            })

    for element in soup.contents:
        if element.name in ['p']:
            add_text_request(element.get_text())
            add_text_request("\n")
        elif element.name in ['h1', 'h2', 'h3']:
            add_text_request(element.get_text(), bold=True)
            add_text_request("\n")
        elif element.name in ['ul']:
            for li in element.find_all('li'):
                add_text_request("\u2022 " + li.get_text())
                add_text_request("\n")
        elif element.name in ['ol']:
            for i, li in enumerate(element.find_all('li'), start=1):
                add_text_request(f"{i}. " + li.get_text())
                add_text_request("\n")
        elif element.name == 'a':
            add_text_request(element.get_text(), link=element['href'])
        elif element.name == 'table':
            for row in element.find_all('tr'):
                row_text = "\t".join([cell.get_text() for cell in row.find_all(['td', 'th'])]) + "\n"
                add_text_request(row_text)
        else:
            add_text_request(element.get_text())
            add_text_request("\n")

    return requests

def update_doc(file_id, doc_content, drive_dir):
    if doc_content:
        try:
            requests = markdown_to_google_doc_requests(doc_content)
        except Exception as e:
            raise ValueError(f"Failed to parse given doc content: {e}")

        docs_service = client('docs', 'v1')
        drive_service = client('drive', 'v3')

        # Retrieve the document to determine its length
        document = docs_service.documents().get(documentId=file_id).execute()
        content = document.get('body').get('content')
        document_length = content[-1].get('endIndex') if content and 'endIndex' in content[-1] else 1

        if document_length > 2:
            # Prepare requests to clear existing document content
            requests = [
                {
                    "deleteContentRange": {
                        "range": {
                            "startIndex": 1,
                            "endIndex": document_length - 1
                        }
                    }
                }
            ] + requests

        # Issue a batch update request to clear and apply new content
        response = docs_service.documents().batchUpdate(
            documentId=file_id,
            body={"requests": requests}
        ).execute()

        print(f"Document updated successfully: {file_id}")

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
