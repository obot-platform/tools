import json
import logging
import os
import sys

import voyageai.error
from voyageai import Client

if __name__ == "__main__":
    api_key = os.getenv("OBOT_VOYAGE_MODEL_PROVIDER_API_KEY")
    if api_key is None:
        print(json.dumps({"error": "VoyageAI API key not found"}))
        sys.exit(1)
    client = Client(api_key=api_key)
    try:
        _ = client.embed(texts=["obot"], model="voyage-3-lite")
    except voyageai.error.AuthenticationError as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)
    except Exception as e:
        logging.error(f"VoyageAI validation failed embedding a test text: {str(e)}")
        print(json.dumps({"error": "unknown error during VoyageAI embedding test"}))
        sys.exit(1)
    print("VoyageAI credential validation successful")
