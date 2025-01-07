import os

from hubspot import HubSpot
from hubspot.crm.deals.exceptions import NotFoundException

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def delete_deal(deal_id):
    try:
        client.crm.deals.basic_api.archive(deal_id)
    except NotFoundException:
        print("No deal found with ID ", deal_id)
        exit(0)
    except Exception as e:
        print("Exception when calling deals->delete: %s\n" % e)
        exit(1)

    print(f"Delete request for deal ID {deal_id} has been sent.")


if __name__ == "__main__":
    deal_id = os.getenv("DEAL_ID")

    delete_deal(deal_id)
