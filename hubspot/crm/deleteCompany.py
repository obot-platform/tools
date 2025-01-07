import os

from hubspot import HubSpot
from hubspot.crm.companies.exceptions import NotFoundException

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def delete_company(company_id):
    try:
        client.crm.companies.basic_api.archive(company_id)
    except NotFoundException:
        print("No company found with ID ", company_id)
        exit(0)
    except Exception as e:
        print("Exception when calling companies->delete: %s\n" % e)
        exit(1)

    print(f"Delete request for company ID {company_id} has been sent.")


if __name__ == "__main__":
    company_id = os.getenv("COMPANY_ID")

    delete_company(company_id)
