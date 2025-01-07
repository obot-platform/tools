import os

from hubspot import HubSpot
from hubspot.crm.associations.v4.exceptions import ApiException


def get_object_associations_to_type(client, from_object_type, from_object_id, to_object_type):
    try:
        # resp = client.crm.associations.v4.schema.definitions_api.get_all(from_object_type=from_object_type,
        #                                                                  to_object_type=to_object_type)
        resp = client.crm.associations.v4.basic_api.get_page(object_type=from_object_type, object_id=from_object_id, to_object_type=to_object_type)
    except ApiException as e:
        print("Exception when calling associations->get_page: %s\n" % e)
        print("This should not happen, there is something wrong with the request.")
        exit(1)

    print(resp)
    # print(
    #     f"The association type ID from {from_object_type} to {to_object_type} is {resp.results[0].type_id}. This is a {resp.results[0].category} association.")
    return resp


if __name__ == "__main__":
    token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
    client = HubSpot(access_token=token)
    from_object_type = os.getenv("FROM_OBJECT_TYPE")
    from_object_id = os.getenv("FROM_OBJECT_ID")
    to_object_type = os.getenv("TO_OBJECT_TYPE")
    get_object_associations_to_type(client, from_object_type, from_object_id, to_object_type)
