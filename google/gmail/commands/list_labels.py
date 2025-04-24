import os
from apis.labels import list_labels, get_label
from apis.helpers import client


def list_labels_tool():
    label_id = os.getenv("LABEL_ID", None)
    service = client("gmail", "v1")

    if label_id:  # get a specific label if label_id is provided
        label = get_label(service, label_id)
        print(f"Label: {label}")
    else:
        labels = list_labels(service)
        custom_labels = [
            l for l in labels if l.get("type") == "user"
        ]  # only show custom labels
        print(f"Custom labels: {custom_labels}")
