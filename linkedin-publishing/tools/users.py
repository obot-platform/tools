from tools.helper import ACCESS_TOKEN
from linkedin_api.clients.restli.client import RestliClient
import time
import sys


def get_user(client: RestliClient) -> dict:

    MAX_RETRIES = 3
    DELAY = 1  # seconds between retries

    for attempt in range(1, MAX_RETRIES + 1):
        try:
            response = client.get(resource_path="/userinfo", access_token=ACCESS_TOKEN)
            if response.status_code >= 200 and response.status_code < 300:
                break
        except Exception as e:
            print(f"Attempt {attempt}: Failed with error: {e}", file=sys.stderr)
            print(
                f"Response: {response.status_code} - {response.entity}", file=sys.stderr
            )

        if attempt < MAX_RETRIES:
            time.sleep(DELAY)  # Wait before retrying
    else:
        raise Exception(
            f"Failed to get user info, Error: {response.status_code} - {response.entity}"
        )

    return response.entity
