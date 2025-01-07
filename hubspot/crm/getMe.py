import os

import hubspot
from hubspot import HubSpot


def get_me(client: hubspot.Client):
    try:
        user = client.oauth.access_tokens_api.get(token)
        print(f"Current user is {user.user}. UserId is {user.user_id}.")
        owners = client.crm.owners.owners_api.get_page(email=user.user)
        print(f"OwnerId is {owners.results[0].id}. Use this to look up ownership associations.")
    except Exception as e:
        print(f"Error getting current user information: {e}")



if __name__ == "__main__":
    token = os.getenv("GPTSCRIPT_API_HUBAPI_COM_BEARER_TOKEN")
    client = HubSpot(access_token=token)
    get_me(client)
