import os
from apis.labels import delete_label
from apis.helpers import client


def delete_label_tool():
    service = client("gmail", "v1")
    label_id = os.getenv("LABEL_ID")
    if not label_id:
        print(f"Error: LABEL_ID is required and is not set")
        return
    label = delete_label(service, label_id)
    print(f"Label deleted: {label}")
