#!/usr/bin/env python3

import os
import sys
import json
import base64
from googleapiclient.errors import HttpError
from gptscript import gptscript
import tiktoken
import subprocess

from helpers import client

TIKTOKEN_MODEL = "gpt-4o"
enc = tiktoken.encoding_for_model(TIKTOKEN_MODEL)
TOKEN_THRESHOLD = 10000

def main():
    email_id = os.getenv('EMAIL_ID')
    attachment_id = os.getenv('ATTACHMENT_ID')

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
        process = subprocess.Popen([os.getenv('GPTSCRIPT_TOOL_DIR') + '/../../knowledge/bin/gptscript-go-tool', 'load', '--format', 'markdown', '-', '-'], stdin=subprocess.PIPE)
        process.communicate(input=attachment_data)
        if process.returncode != 0:
            raise Exception(f"gptscript-go-tool failed with return code {process.returncode}")
        stdout, _ = process.communicate()

        # Check token count using tiktoken
        tokens = enc.encode(stdout.decode('utf-8', errors='ignore'))
        token_count = len(tokens)
        if token_count > TOKEN_THRESHOLD:
            print(json.dumps({
                'error': f'Attachment content exceeds maximum token limit of {TOKEN_THRESHOLD} (got {token_count} tokens)'
            }))
            sys.exit(1)
        print(stdout.decode())
        
    except Exception as e:
        print(json.dumps({
            'error': f'Failed to retrieve attachment: {e}'
        }))

if __name__ == '__main__':
    main()
