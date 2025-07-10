from googleapiclient.errors import HttpError


def list_labels(service) -> list[dict]:
    try:
        response = service.users().labels().list(userId="me").execute()
        return response.get("labels", [])
    except HttpError as e:
        raise Exception(f"HttpError: An error occurred listing labels: {e}")
    except Exception as e:
        raise Exception(f"An unexpected error occurred listing labels: {e}")


def get_label(service, label_id) -> dict:
    try:
        return service.users().labels().get(userId="me", id=label_id).execute()
    except HttpError as e:
        raise Exception(f"HttpError: An error occurred getting label {label_id}: {e}")
    except Exception as e:
        raise Exception(f"An unexpected error occurred getting label {label_id}: {e}")


def create_label(
    service, name, label_list_visibility="labelShow", message_list_visibility="show"
) -> dict:
    label = {
        "name": name,
        "labelListVisibility": label_list_visibility,
        "messageListVisibility": message_list_visibility,
    }
    try:
        return service.users().labels().create(userId="me", body=label).execute()
    except HttpError as e:
        raise Exception(f"HttpError: An error occurred creating label '{name}': {e}")
    except Exception as e:
        raise Exception(f"An unexpected error occurred creating label '{name}': {e}")


def update_label(
    service,
    label_id,
    name=None,
    label_list_visibility=None,
    message_list_visibility=None,
) -> dict:
    label = {"id": label_id}
    if name:
        label["name"] = name
    if label_list_visibility:
        label["labelListVisibility"] = label_list_visibility
    if message_list_visibility:
        label["messageListVisibility"] = message_list_visibility

    try:
        return (
            service.users()
            .labels()
            .update(userId="me", id=label_id, body=label)
            .execute()
        )
    except HttpError as e:
        raise Exception(f"HttpError: An error occurred updating label {label_id}: {e}")
    except Exception as e:
        raise Exception(f"An unexpected error occurred updating label {label_id}: {e}")


def delete_label(service, label_id) -> str:
    try:
        service.users().labels().delete(userId="me", id=label_id).execute()
        return f"Label {label_id} deleted successfully."
    except HttpError as e:
        raise Exception(f"HttpError: An error occurred deleting label {label_id}: {e}")
    except Exception as e:
        raise Exception(f"An unexpected error occurred deleting label {label_id}: {e}")
