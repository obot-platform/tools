import googleapiclient.errors # Add import for potential API errors

def list_labels(service):
    try:
        response = service.users().labels().list(userId="me").execute()
        return response.get("labels", [])
    except googleapiclient.errors.HttpError as e:
        # Re-raise the specific API error or a custom exception
        raise Exception(f"An error occurred listing labels: {e}") from e
    except Exception as e:
        # Catch other potential exceptions
        raise Exception(f"An unexpected error occurred listing labels: {e}") from e


def get_label(service, label_id):
    try:
        return service.users().labels().get(userId="me", id=label_id).execute()
    except googleapiclient.errors.HttpError as e:
        raise Exception(f"An error occurred getting label {label_id}: {e}") from e
    except Exception as e:
        raise Exception(f"An unexpected error occurred getting label {label_id}: {e}") from e


def create_label(
    service, name, label_list_visibility="labelShow", message_list_visibility="show"
):
    label = {
        "name": name,
        "labelListVisibility": label_list_visibility,
        "messageListVisibility": message_list_visibility,
    }
    try:
        return service.users().labels().create(userId="me", body=label).execute()
    except googleapiclient.errors.HttpError as e:
        raise Exception(f"An error occurred creating label '{name}': {e}") from e
    except Exception as e:
        raise Exception(f"An unexpected error occurred creating label '{name}': {e}") from e


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

    try:
        return (
            service.users().labels().update(userId="me", id=label_id, body=label).execute()
        )
    except googleapiclient.errors.HttpError as e:
        raise Exception(f"An error occurred updating label {label_id}: {e}") from e
    except Exception as e:
        raise Exception(f"An unexpected error occurred updating label {label_id}: {e}") from e


def delete_label(service, label_id):
    try:
        service.users().labels().delete(userId="me", id=label_id).execute()
        return f"Label {label_id} deleted successfully."
    except googleapiclient.errors.HttpError as e:
        raise Exception(f"An error occurred deleting label {label_id}: {e}") from e
    except Exception as e:
        raise Exception(f"An unexpected error occurred deleting label {label_id}: {e}") from e
