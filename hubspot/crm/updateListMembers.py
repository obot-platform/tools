import os

from hubspot import HubSpot
from hubspot.crm.associations.v4.exceptions import ApiException
from hubspot.crm.lists import MembershipChangeRequest

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def update_list_members(list_id, add, remove):
    change_request: MembershipChangeRequest = MembershipChangeRequest(
        record_ids_to_add=add,
        record_ids_to_remove=remove
    )
    try:
        client.crm.lists.memberships_api.add_and_remove(list_id, change_request)
        print (f"Updated list with ID: {list_id}")
    except ApiException as e:
        print("Exception when calling lists->update_members: %s\n" % e)
        exit(1)


if __name__ == "__main__":
    add_from_env = os.getenv("ADD_CONTACTS",[])
    ids_to_add, ids_to_remove = None, None

    if isinstance(add_from_env, str):
        add_from_env = add_from_env.split(",")
        ids_to_add = [item.strip() for item in add_from_env]

    remove_from_env = os.getenv("REMOVE_CONTACTS",[])
    if isinstance(remove_from_env, str):
        remove_from_env = remove_from_env.split(",")
        ids_to_remove = [item.strip() for item in remove_from_env]

    list_id = os.getenv("LIST_ID")

    update_list_members(list_id, ids_to_add, ids_to_remove)
