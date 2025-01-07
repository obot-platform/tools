import json
import os

from hubspot import HubSpot
from hubspot.crm.associations.v4.exceptions import ApiException

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def search_tickets(filters):
    filters_map = json.loads(filters)
    hubspotFilters = []
    for k,v in filters_map.items():
        hubspotFilters.append({
            "propertyName": k,
            "operator": "CONTAINS_TOKEN",
            "value": v,
        })
    filterGroups = {
        "filterGroups": [{"filters": hubspotFilters}],
    }

    try:
        results = client.crm.tickets.search_api.do_search(filterGroups)
        if len(results.results) == 0:
            print("No search results found")
            exit(0)
        for result in results.results:
            print(f"Ticket ID: {result.id}")
            print(f"    Properties: {result.properties}")
    except ApiException as e:
        print("Exception when calling tickets->search: %s\n" % e)
        exit(1)


if __name__ == "__main__":
    filters = os.getenv("FILTERS")

    search_tickets(filters)
