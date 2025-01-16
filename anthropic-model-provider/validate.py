import json
import os
import sys

from anthropic import Anthropic


def validate():
    client = Anthropic(api_key=os.environ.get("OBOT_ANTHROPIC_MODEL_PROVIDER_API_KEY", ""))
    try:
        _ = client.models.list(limit=1)
        return True
    except Exception as e:
        raise Exception(f"Anthropic API Key validation failed: {e}")

if __name__ == "__main__":
    try:
        validate()
        print("Anthropic API Key is valid.")
    except Exception as e:
        print(f"Anthropic API Key validation failed: {e}", file=sys.stderr)
        # The errors returned by the Anthropic SDK are not suitable to be returned to the user
        # (e.g. a not_found_error on invalid API key)
        # so we return a generic error message instead and log the actual error
        print(json.dumps({"error": "Anthropic API Key validation failed."}))
        exit(1)