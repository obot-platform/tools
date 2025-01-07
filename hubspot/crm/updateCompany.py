import json
import os

from hubspot import HubSpot
from hubspot.crm.associations.v4.exceptions import ApiException

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def update_company(company_id, properties):
    json_properties = json.loads(properties)
    if "properties" not in json_properties.keys():
        json_properties = {"properties": json_properties}

    try:
        result = client.crm.companies.basic_api.update(company_id, json_properties)
        print (f" Created company with ID: {result.id}")
    except ApiException as e:
        print("Exception when calling companies->update: %s\n" % e)
        exit(1)


if __name__ == "__main__":
    properties = os.getenv("PROPERTIES")
    company_id = os.getenv("COMPANY_ID")

    update_company(company_id, properties)
