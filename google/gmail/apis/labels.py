def list_labels(service):
    response = service.users().labels().list(userId="me").execute()
    return response.get("labels", [])


def get_label(service, label_id):
    return service.users().labels().get(userId="me", id=label_id).execute()


def create_label(
    service, name, label_list_visibility="labelShow", message_list_visibility="show"
):
    label = {
        "name": name,
        "labelListVisibility": label_list_visibility,
        "messageListVisibility": message_list_visibility,
    }
    return service.users().labels().create(userId="me", body=label).execute()


def update_label(
    service,
    label_id,
    name=None,
    label_list_visibility=None,
    message_list_visibility=None,
):
    label = {"id": label_id}
    if name:
        label["name"] = name
    if label_list_visibility:
        label["labelListVisibility"] = label_list_visibility
    if message_list_visibility:
        label["messageListVisibility"] = message_list_visibility

    return (
        service.users().labels().update(userId="me", id=label_id, body=label).execute()
    )


def delete_label(service, label_id):
    service.users().labels().delete(userId="me", id=label_id).execute()
    return f"Label {label_id} deleted successfully."
