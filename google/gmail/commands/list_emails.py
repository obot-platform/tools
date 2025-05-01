import os
import sys

from apis.helpers import client, parse_label_ids
from apis.messages import list_messages, create_gptscript_dataset

NON_PRIMARY_CATEGORIES_MAP = {
    "social": "CATEGORY_SOCIAL",
    "promotions": "CATEGORY_PROMOTIONS",
    "updates": "CATEGORY_UPDATES",
    "forums": "CATEGORY_FORUMS",
}


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
    main_query = query
    is_primary_category = False
    if any(
        label.upper() == "ALL" for label in label_ids
    ):  # check if ALL is in the label_ids
        label_ids = []
    elif "INBOX" in label_ids:
        if category in NON_PRIMARY_CATEGORIES_MAP:
            label_ids.append(
                NON_PRIMARY_CATEGORIES_MAP[category]
            )  # we use the internal category names for non-primary categories
        else:  # handle primary categories separately. use query to mimic gmail UI behavior
            main_query = f"{query} category:{category.lower()}"
            is_primary_category = True
    after = os.getenv("AFTER", "")
    before = os.getenv("BEFORE", "")

    service = client("gmail", "v1")
    response = list_messages(service, main_query, label_ids, max_results, after, before)
    if len(response) > 0:
        output_str = await create_gptscript_dataset(
            service, response, main_query, label_ids
        )
        print(output_str)
        return

    # If not primary category or no results found, we can exit early
    if not is_primary_category:
        print("No emails found")
        return

    # For primary category, ESTIMATE if the feature is enabled
    estimate_response = list_messages(
        service, "category:primary", ["INBOX"], 10, "", ""
    )
    if len(estimate_response) > 0:
        print("No emails found")
        return

    # If categories aren't enabled, try without category filter
    no_category_response = list_messages(
        service, query, label_ids, max_results, after, before
    )
    if len(no_category_response) > 0:
        output_str = await create_gptscript_dataset(
            service, no_category_response, query, label_ids
        )
        print(output_str)
        return

    print("No emails found")
    return
