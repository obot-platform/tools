import json
import os

from hubspot import HubSpot
from hubspot.crm.associations.v4.exceptions import ApiException

token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
client = HubSpot(access_token=token)


def get_simple_properties(object_type):
        try:
            properties = client.crm.properties.core_api.get_all(object_type)
            properties = properties.to_dict()
            exclude_names = [
                "surveymonkeyeventlastupdated",
                "recent_deal_close_date",
                "hs_v2_latest_time",
                "hs_v2_date_",
                "hs_v2_cumulative",
                "hs_time_",
                "hs_last_",
                "hs_googleplusid",
            ]
            exclude_descriptions = [
                "Set automatically",
                "Automatically synced",
                "This property is no longer in use",
                "This is set by HubSpot",
                "This is automatically set by HubSpot",
                "This is set automatically by HubSpot",
                "It can be set in HubSpot",
                "This value is set automatically by HubSpot",
            ]

            exclude_enums_from_names = [
                "hs_language",
                "hs_timezone",
                "hs_sub_role",
            ]

            properties_json = []
            for p in properties['results']:
                if any(substring in p['name'] for substring in exclude_names):
                    continue
                if any(substring in p['description'] for substring in exclude_descriptions):
                    continue
                if p['archived'] is True:
                    continue
                if p['modification_metadata']['read_only_value'] is True:
                    continue
                serialize = {
                    "name": p['name'],
                    # "description": p['description'],
                    # "type": p['type']
                }
                if p['type'] == "enumeration" and p['name'] not in exclude_enums_from_names:
                    options = []
                    for option in p['options']:
                        options.append(option['label'])
                    serialize['options'] = options
                properties_json.append(json.dumps(serialize))
            print("Returning available properties for object type: " + object_type)
            print(properties_json)

            # print(token)

        except ApiException as e:
            print("Exception when calling properties->get_all: %s\n" % e)
            exit(1)


if __name__ == "__main__":
    object_type = os.getenv("OBJECT_TYPE")
    if object_type is None:
        print("Please provide an object type to get the properties for.")
        exit(1)
    get_simple_properties(object_type)
