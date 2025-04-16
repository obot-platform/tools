#!/usr/bin/env python3

import os
import sys
import json
import base64
import requests

from helpers import client

class ObotClient:
    def __init__(self, server_url, token):
        self.server_url = server_url.rstrip('/') + "/api"
        self.token = token

    def upload_file(self, project_id, thread_id, assistant_id, filename, content):
        url = f"{self.server_url}/assistants/{assistant_id}/projects/{project_id}/threads/{thread_id}/files/{filename}"
        headers = {"Authorization": f"Bearer {self.token}"}
        response = requests.post(url, data=content, headers=headers)
        response.raise_for_status()
        return response

    def upload_knowledge_file(self, project_id, thread_id, assistant_id, filename, content):
        url = f"{self.server_url}/assistants/{assistant_id}/projects/{project_id}/threads/{thread_id}/knowledge-files/{filename}"
        headers = {"Authorization": f"Bearer {self.token}"}
        response = requests.post(url, data=content, headers=headers)
        response.raise_for_status()
        return response


def main():
    email_id = os.getenv('EMAIL_ID')
    attachment_id = os.getenv('ATTACHMENT_ID')
    filename = os.getenv('FILENAME')

    if not email_id or not attachment_id:
        print(json.dumps({
            'error': 'Both email_id and attachment_id are required parameters'
        }))
        sys.exit(1)

    service = client('gmail', 'v1')
    try:
        attachment = service.users().messages().attachments().get(
                userId='me',
                messageId=email_id,
                id=attachment_id
            ).execute()
        attachment_data = base64.urlsafe_b64decode(attachment['data'])
        project_id = "p1" + os.getenv("OBOT_PROJECT_ID").replace("t1", "", 1)

        obot_client = ObotClient(os.getenv('OBOT_SERVER_URL'), os.getenv('OBOT_TOKEN'))
        obot_client.upload_file(project_id, os.getenv('OBOT_THREAD_ID'), os.getenv('OBOT_AGENT_ID'), filename, attachment_data)
        obot_client.upload_knowledge_file(project_id, os.getenv('OBOT_THREAD_ID'), os.getenv('OBOT_AGENT_ID'), filename, attachment_data)
    except Exception as e:
        print(json.dumps({
            'error': f'Failed to retrieve attachment: {e}'
        }))

if __name__ == '__main__':
    main()
