import os

from hubspot import HubSpot
from hubspot.crm.associations.v4.exceptions import ApiException

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def get_simple_properties(object_type, group_name):
        try:
            result = client.crm.properties.groups_api.get_by_name(object_type, group_name)
            print(result.to_dict())
        except ApiException as e:
            print("Exception when calling properties->get_group_by_name: %s\n" % e)
            exit(1)


if __name__ == "__main__":
    object_type = os.getenv("OBJECT_TYPE")
    if object_type is None:
        print("Please provide an object type to get the properties for.")
        exit(1)
    group_name = os.getenv("GROUP_NAME")
    if group_name is None:
        print("Please provide a group name to get the properties for.")
        exit(1)
    get_simple_properties(object_type, group_name)
