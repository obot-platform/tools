import json
import sys

from helpers import configure, list_openai

if __name__ == "__main__":
    try:
        _, c, _, rgrp = configure()
        list_openai(c, rgrp)

        print("Successfully configured Azure OpenAI model provider")
    except Exception as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)
