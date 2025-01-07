import os

from hubspot import HubSpot
from hubspot.crm.contacts.exceptions import NotFoundException

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def delete_contact(contact_id):
    try:
        client.crm.contacts.basic_api.archive(contact_id)
    except NotFoundException:
        print("No contact found with ID ", contact_id)
        exit(0)
    except Exception as e:
        print("Exception when calling contacts->delete: %s\n" % e)
        exit(1)

    print(f"Delete request for contact ID {contact_id} has been sent.")


if __name__ == "__main__":
    contact_id = os.getenv("CONTACT_ID")

    delete_contact(contact_id)
