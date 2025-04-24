import os
import sys
from apis.messages import modify_message_labels
from apis.helpers import client, str_to_bool, parse_label_ids

def modify_message_labels_tool():
    email_id = os.getenv("EMAIL_ID")
    if not email_id:
        print(f"required environment variable EMAIL_ID not set")
        sys.exit(1)

    add_labels = os.getenv("ADD_LABEL_IDS", None)
    if add_labels:
        add_labels = parse_label_ids(add_labels)
    remove_labels = os.getenv("REMOVE_LABEL_IDS", None)
    if remove_labels:
        remove_labels = parse_label_ids(remove_labels)

    env_flags = {
        "archive": "ARCHIVE",
        "mark_as_read": "MARK_AS_READ",
        "mark_as_starred": "MARK_AS_STARRED",
        "mark_as_important": "MARK_AS_IMPORTANT"
    }

    parsed_flags = {
        key: str_to_bool(os.getenv(env_key)) if os.getenv(env_key) is not None else None
        for key, env_key in env_flags.items()
    }

    # Unpack if needed
    archive = parsed_flags["archive"]
    mark_as_read = parsed_flags["mark_as_read"]
    mark_as_starred = parsed_flags["mark_as_starred"]
    mark_as_important = parsed_flags["mark_as_important"]

    service = client()
    res = modify_message_labels(
        service,
        email_id,
        add_labels,
        remove_labels,
        archive,
        mark_as_read,
        mark_as_starred,
        mark_as_important,
    )
    print(res)
