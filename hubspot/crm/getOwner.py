import json
import os

from hubspot import HubSpot
from hubspot.crm.owners.exceptions import NotFoundException

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def get_owner(owner_id):
    try:
        result = client.crm.owners.owners_api.get_by_id(owner_id)
        json_result = json.dumps(result.to_dict(), default=str)
        print(json_result)
    except NotFoundException:
        print("No owner found with ID ", owner_id)
        exit(0)
    except Exception as e:
        print("Exception when calling owners->get: %s\n" % e)
        exit(1)


if __name__ == "__main__":
    owner_id = os.getenv("OWNER_ID")

    get_owner(owner_id)
