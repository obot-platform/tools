import logging
import os
import sys

from tavily import TavilyClient
from urllib3.util import parse_url


def main():
    if len(sys.argv) < 2:
        print("Usage: python main.py [search | extract]")
        sys.exit(1)

    command = sys.argv[1]
    client = TavilyClient()  # env TAVILY_API_KEY required

    match command:
        case "search":
            query = os.getenv("QUERY", "").strip()
            if not query:
                print("No search query provided")
                sys.exit(1)

            domains_str = os.getenv("INCLUDE_DOMAINS", "")
            include_domains = [
                domain.strip() for domain in domains_str.split(",") if domain.strip()
            ]
            response = client.search(
                query=query,
                include_answer=False,  # no LLM-generated answer needed - we'll do that
                include_raw_content=True,
                max_results=20
                if len(include_domains) == 0
                else 5
                * len(
                    include_domains
                ),  # broader search if general, more narrow if scoped to specific sites
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


if __name__ == "__main__":
    main()
