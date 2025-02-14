import logging
import os
import sys
from typing import List

from tavily import TavilyClient
from urllib3.util import parse_url


def main():
    if len(sys.argv) < 2:
        print("Usage: python main.py [search | extract]")
        sys.exit(1)

    command = sys.argv[1]
    client = TavilyClient()  # env TAVILY_API_KEY required

    match command:
        case "search" | "safe-search":
            query = os.getenv("QUERY", "").strip()
            if not query:
                print("No search query provided")
                sys.exit(1)

            domains_str = os.getenv("INCLUDE_DOMAINS", "")
            include_domains = [
                domain.strip() for domain in domains_str.split(",") if domain.strip()
            ]

            # safe-search is a special case where we only allow certain domains
            # this is a different command so that we can use the same code for different tool implementations
            if command == "safe-search":
                include_domains = check_allowed_include_domains(include_domains)

            max_results = 20  # broader search if general,
            if len(include_domains) > 0:
                max_results = 5 * len(
                    include_domains
                )  # more narrow  search if scoped to specific sites
            if os.getenv("MAX_RESULTS", "").strip():
                max_results = int(os.getenv("MAX_RESULTS"))  # override with env var

            response = client.search(
                query=query,
                include_answer=os.getenv("INCLUDE_ANSWER", "").lower()
                == "true",  # no LLM-generated answer needed by default - we'll do that
                include_raw_content=os.getenv("INCLUDE_RAW_CONTENT", "").lower()
                != "false",  # include raw content by default
                max_results=max_results,
                include_domains=include_domains,
            )
        case "extract":
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
    logging.info(f"Tavily - response:\n{response}")
    print(response)


def check_allowed_include_domains(include_domains: List[str]) -> List[str]:
    # TAVILY_ALLOWED_DOMAINS has the TAVILY_ prefix as it will be set by Obot directly in the env,
    # while e.g. INCLUDE_DOMAINS is a tool parameter
    allowed_domains_str = os.getenv("TAVILY_ALLOWED_DOMAINS", "")
    allowed_domains = [
        domain.strip() for domain in allowed_domains_str.split(",") if domain.strip()
    ]

    if len(allowed_domains) == 0:
        print("No allowed domains provided")
        sys.exit(1)

    # allow not setting INCLUDE_DOMAINS -  fallback to all allowed domains
    if len(include_domains) == 0:
        return allowed_domains

    allowed_include_domains = []
    disallowed_include_domains = []

    for domain in include_domains:
        if domain in allowed_domains:
            allowed_include_domains.append(domain)
        else:
            disallowed_include_domains.append(domain)

    if len(disallowed_include_domains) > 0:
        if os.getenv("TAVILY_ALLOWED_DOMAINS_STRICT", "").lower() == "true":
            print(
                f"Tried to access domains {disallowed_include_domains} which are not listed in allowed domains {allowed_domains}"
            )
            sys.exit(1)
        logging.warning(
            f"Filtered out {disallowed_include_domains} as they are not listed in allowed domains {allowed_domains}. Continuing with {allowed_include_domains}"
        )
    include_domains = allowed_include_domains
    return include_domains


if __name__ == "__main__":
    main()
