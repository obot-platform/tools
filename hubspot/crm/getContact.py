import json
import os

from hubspot import HubSpot
from hubspot.crm.contacts.exceptions import NotFoundException

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def get_contact(contact_id):
    try:
        result = client.crm.contacts.basic_api.get_by_id(contact_id)
        json_result = json.dumps(result.to_dict(), default=str)
        print(json_result)
    except NotFoundException:
        print("No contact found with ID ", contact_id)
        exit(0)
    except Exception as e:
        print("Exception when calling contacts->get: %s\n" % e)
        exit(1)


if __name__ == "__main__":
    contact_id = os.getenv("CONTACT_ID")

    get_contact(contact_id)
