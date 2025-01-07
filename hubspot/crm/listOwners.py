import json
import os

from hubspot import HubSpot
from hubspot.crm.owners.exceptions import NotFoundException

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def get_all_owners():
    try:
        result = client.crm.owners.get_all()
        json_result = json.dumps(result, default=str)
        print(json_result)
    except NotFoundException:
        print("No Owners found")
        exit(0)
    except Exception as e:
        print("Exception when calling owners->get_all: %s\n" % e)
        exit(1)


if __name__ == "__main__":
    get_all_owners()
