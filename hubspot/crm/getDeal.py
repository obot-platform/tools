import json
import os

from hubspot import HubSpot
from hubspot.crm.deals.exceptions import NotFoundException

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def get_deal(deal_id):
    try:
        result = client.crm.deals.basic_api.get_by_id(deal_id)
        json_result = json.dumps(result.to_dict(), default=str)
        print(json_result)
    except NotFoundException:
        print("No deal found with ID ", deal_id)
        exit(0)
    except Exception as e:
        print("Exception when calling deals->get: %s\n" % e)
        exit(1)


if __name__ == "__main__":
    deal_id = os.getenv("DEAL_ID")

    get_deal(deal_id)
