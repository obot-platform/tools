import os
from apis.permissions import create_permission
from apis.helper import get_client


def create_permission_tool() -> None:
    client = get_client("drive", "v3")
    file_id = os.getenv("FILE_ID")
    role = os.getenv("ROLE")
    type = os.getenv("TYPE")
    email_address = os.getenv("EMAIL_ADDRESS")
    domain = os.getenv("DOMAIN")

    if not file_id:
        print("Error: FILE_ID environment variable is required but not set")
        return
    if not role:
        print("Error: ROLE environment variable is required but not set")
        return
    if not type:
        print("Error: TYPE environment variable is required but not set")
        return

    valid_roles = [
        "owner",
        "organizer",
        "fileOrganizer",
        "writer",
        "commenter",
        "reader",
    ]
    if role not in valid_roles:
        print(f"Error: Invalid role '{role}'")
        print(f"Valid roles are: {', '.join(valid_roles)}")
        print("Note:")
        print("- 'owner' can only be used for My Drive files")
        print("- 'organizer' and 'fileOrganizer' can only be used for shared drives")
        return

    valid_types = ["user", "group", "domain", "anyone"]
    if type not in valid_types:
        print(f"Error: TYPE must be one of {valid_types}, but got {type}")
        return

    if type in ["user", "group"] and not email_address:
        print(
            "Error: EMAIL_ADDRESS environment variable is required for user/group permission"
        )
        return
    if type == "domain" and not domain:
        print("Error: DOMAIN environment variable is required for domain permission")
        return

    permission = create_permission(
        client,
        file_id=file_id,
        role=role,
        type=type,
        email_address=email_address,
        domain=domain,
    )
    print(permission)
