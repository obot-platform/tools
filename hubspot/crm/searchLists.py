import json
import os

from hubspot import HubSpot
from hubspot.crm.associations.v4.exceptions import ApiException
from hubspot.crm.lists import ListSearchRequest

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def search_lists(search_terms):
    search: ListSearchRequest = ListSearchRequest(
        query=search_terms,
    )
    try:
        result = client.crm.lists.lists_api.do_search(search)
        if result.total == 0:
            print("No search results found")
            exit(0)
        json_result = json.dumps(result.to_dict(), default=str)
        print(json_result)
    except ApiException as e:
        print("Exception when calling lists->search: %s\n" % e)
        exit(1)


if __name__ == "__main__":
    search_terms = os.getenv("SEARCH_TERMS")

    search_lists(search_terms)
