import json
import os

from hubspot import HubSpot
from hubspot.crm.tickets.exceptions import NotFoundException

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def get_ticket(ticket_id):
    try:
        result = client.crm.tickets.basic_api.get_by_id(ticket_id)
        json_result = json.dumps(result.to_dict(), default=str)
        print(json_result)
    except NotFoundException:
        print("No ticket found with ID ", ticket_id)
        exit(0)
    except Exception as e:
        print("Exception when calling tickets->get: %s\n" % e)
        exit(1)


if __name__ == "__main__":
    ticket_id = os.getenv("TICKET_ID")

    get_ticket(ticket_id)
