import os

from hubspot import HubSpot
from hubspot.crm.tickets.exceptions import NotFoundException

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def delete_ticket(ticket_id):
    try:
        client.crm.tickets.basic_api.archive(ticket_id)
    except NotFoundException:
        print("No ticket found with ID ", ticket_id)
        exit(0)
    except Exception as e:
        print("Exception when calling tickets->delete: %s\n" % e)
        exit(1)

    print(f"Delete request for ticket ID {ticket_id} has been sent.")


if __name__ == "__main__":
    ticket_id = os.getenv("TICKET_ID")

    delete_ticket(ticket_id)
