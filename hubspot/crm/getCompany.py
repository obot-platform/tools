import json
import os

from hubspot import HubSpot
from hubspot.crm.companies.exceptions import NotFoundException

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def get_company(company_id):
    try:
        result = client.crm.companies.basic_api.get_by_id(company_id)
        json_result = json.dumps(result.to_dict(), default=str)
        print(json_result)
    except NotFoundException:
        print("No company found with ID ", company_id)
        exit(0)
    except Exception as e:
        print("Exception when calling companies->get: %s\n" % e)
        exit(1)


if __name__ == "__main__":
    company_id = os.getenv("COMPANY_ID")

    get_company(company_id)
