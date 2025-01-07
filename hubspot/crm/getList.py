import json
import os

from hubspot import HubSpot
from hubspot.crm.owners.exceptions import NotFoundException

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def get_owner(list_id):
    try:
        result = client.crm.lists.lists_api.get_by_id(list_id)
        json_result = json.dumps(result.to_dict(), default=str)
        print(json_result)
    except NotFoundException:
        print("No list found with ID ", list_id)
        exit(0)
    except Exception as e:
        print("Exception when calling lists->get: %s\n" % e)
        exit(1)


if __name__ == "__main__":
    list_id = os.getenv("LIST_ID")

    get_owner(list_id)
