import os
from apis.labels import update_label
from apis.helpers import client


def update_label_tool():
    service = client("gmail", "v1")
    label_id = os.getenv("LABEL_ID")
    if not label_id:
        print(f"Error: LABEL_ID is required and is not set")
        return
    label_name = os.getenv("LABEL_NAME")
    label_list_visibility = os.getenv("LABEL_LIST_VISIBILITY")
    valid_label_list_visibilities = ["labelShow", "labelHide", "labelShowIfUnread"]
    if (
        label_list_visibility
        and label_list_visibility not in valid_label_list_visibilities
    ):
        print(
            f"Error: invalid value for LABEL_LIST_VISIBILITY: {label_list_visibility}. Must be one of {valid_label_list_visibilities}"
        )
        return
    message_list_visibility = os.getenv("MESSAGE_LIST_VISIBILITY")
    valid_message_list_visibilities = ["show", "hide"]
    if (
        message_list_visibility
        and message_list_visibility not in valid_message_list_visibilities
    ):
        print(
            f"Error: invalid value for MESSAGE_LIST_VISIBILITY: {message_list_visibility}. Must be one of {valid_message_list_visibilities}"
        )
        return
    label = update_label(
        service, label_id, label_name, label_list_visibility, message_list_visibility
    )
    print(f"Label updated: {label}")
