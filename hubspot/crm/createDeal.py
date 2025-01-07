import json
import os

from hubspot import HubSpot
from hubspot.crm.associations.v4.exceptions import ApiException

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def create_deal(properties):
    json_properties = json.loads(properties)
    if "properties" not in json_properties.keys():
        json_properties = {"properties": json_properties}
    try:
        result = client.crm.deals.basic_api.create(json_properties)
        print (f" Created deal with ID: {result.id}")
    except ApiException as e:
        print("Exception when calling deals->create: %s\n" % e)
        exit(1)


if __name__ == "__main__":
    properties = os.getenv("PROPERTIES")

    create_deal(properties)
