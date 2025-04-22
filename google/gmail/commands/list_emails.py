import os
import sys

from apis.helpers import client, parse_label_ids
from apis.messages import list_messages

async def list_emails_tool():
    max_results = os.getenv("MAX_RESULTS", "100")
    if max_results.isdigit():
        max_results = int(max_results)
    else:
        print(f"Invalid max_results: {max_results}. Must be a positive integer.")
        sys.exit(1)

    query = os.getenv("QUERY", "")
    if "after:" in query or "before:" in query:
        print(
            "Error: Please use the parameters `after` and `before` instead of having them in the `query`."
        )
        sys.exit(1)
    default_inbox = "INBOX"
    if query != "":
        default_inbox = ""  # if query is not empty, don't set inbox as default
    labels = os.getenv("LABEL_IDS", default_inbox)
    category = os.getenv("CATEGORY", "primary")
    valid_categories = ["primary", "social", "promotions", "updates", "forums"]
    if category not in valid_categories:
        print(f"Invalid category: {category}. Valid categories are: {valid_categories}")
        sys.exit(1)

    label_ids = parse_label_ids(labels)
    if "ALL" in label_ids:
        label_ids = []
    elif "INBOX" in label_ids:
        query = f"{query} category:{category.lower()}"

    after = os.getenv("AFTER", "")
    before = os.getenv("BEFORE", "")

    service = client("gmail", "v1")
    await list_messages(service, query, label_ids, max_results, after, before)
