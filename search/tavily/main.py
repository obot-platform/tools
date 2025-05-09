import json
import logging
import os
import sys
from typing import List

from tavily import TavilyClient
from urllib3.util import parse_url

tool_name = os.getenv("TAVILY_TOOL_NAME", "Tavily")
if len(sys.argv) > 1:
    tool_name = sys.argv[1]


def main():
    if len(sys.argv) < 2:
        print("Usage: python main.py [search | extract]")
        sys.exit(1)

    command = sys.argv[1]

    match command:
        case "site-search-context":
           site_search_context()
           sys.exit(0)
        case "search" | "site-search":
            client = TavilyClient()  # env TAVILY_API_KEY required
            query = os.getenv("QUERY", "").strip()
            if not query:
                print("No search query provided")
                sys.exit(1)

            # site-search is a special case where we only allow certain domains
            # this is a different command so that we can use the same code for different tool implementations
            if command == "site-search":
                include_domains = get_allowed_domains_or_fail()
            else:
                domains_str = os.getenv("INCLUDE_DOMAINS", "")
                include_domains = [
                    domain.strip() for domain in domains_str.split(",") if domain.strip()
                ]

            max_results = 5  # broader search if general,
            if len(include_domains) > 0:
                max_results = 3 * len(
                    include_domains
                )  # more narrow  search if scoped to specific sites
            if os.getenv("MAX_RESULTS", "").strip():
                max_results = int(os.getenv("MAX_RESULTS"))  # override with env var

            time_range = os.getenv("TIME_RANGE", "").lower().strip()
            if time_range not in ["", "day", "week", "month", "year"]:
                print("Invalid time range: must be day, week, month, or year")
                sys.exit(1)

            search_params = {
                "query": query,
                "include_answer": os.getenv("INCLUDE_ANSWER", "").lower() == "true",  # no LLM-generated answer needed by default - we'll do that
                "include_raw_content": os.getenv("INCLUDE_RAW_CONTENT", "").lower() != "false",  # include raw content by default
                "max_results": max_results,
                "include_domains": include_domains,
            }

            if time_range:
                search_params["time_range"] = time_range

            response = client.search(**search_params)
        case "extract":
            client = TavilyClient()  # env TAVILY_API_KEY required
            url = parse_url(os.getenv("URL").strip())

            # default to https:// if no scheme is provided
            if not url.scheme:
                url = parse_url("https://" + url.url.removeprefix("://"))

            # Only http and https are supported
            if url.scheme not in ["http", "https"]:
                print("Invalid URL scheme: must be http or https")
                sys.exit(1)

            response = client.extract(url.url)
        case _:
            print(f"Unknown command: {command}")
            sys.exit(1)

    logging.basicConfig(stream=sys.stderr, level=logging.INFO)
    if not response:
        logging.error(f"Tavily - {command} - No results found")
        print("No results found")
        sys.exit(1)

    # print the response as a valid json object
    print(json.dumps(response))

def site_search_context():
    print(f"""WEBSITE KNOWLEDGE:
Use the {tool_name} website knowledge tool to search the following"
configured domains:
""")
    config = json.loads(os.getenv("OBOT_WEBSITE_KNOWLEDGE", "{}"))
    for site_def in config.get("sites", []):
        site = site_def.get("site", "")
        description = site_def.get("description", "")
        if site:
            print(f"DOMAIN: {site}\n")
            if description:
                print(f"DESCRIPTION: {description}\n")
    print(f"""END WEBSITE KNOWLEDGE
""")

def get_allowed_domains_or_fail() -> List[str]:
    result = []
    config = json.loads(os.getenv("OBOT_WEBSITE_KNOWLEDGE", "{}"))
    for site_def in config.get("sites", []):
        site = site_def.get("site", "")
        if site:
            result.append(site)
    if len(result) == 0:
        logging.error("No allowed domains found in OBOT_WEBSITE_KNOWLEDGE")
        sys.exit(1)
    return result


if __name__ == "__main__":
    main()
