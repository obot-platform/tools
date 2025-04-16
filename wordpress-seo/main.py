from tools.helper import setup_logger
import sys
from tools.keywords_suggestions import keywords_suggestions_tool
from tools.keyword_density import keyword_density_metrics_tool
from tools.readability import readability_metrics_tool
from tools.long_tail_keywords import google_search_suggestions_tool
import os
import json

logger = setup_logger(__name__)


def main():
    if len(sys.argv) < 2:
        print("Usage: python main.py <command>")
        sys.exit(1)

    command = sys.argv[1]
    try:
        match command:
            case "KeywordDensityMetrics":
                title = os.getenv("TITLE")
                content = os.getenv("CONTENT")
                primary_keyword = os.getenv("PRIMARY_KEYWORD")
                if not primary_keyword:
                    raise ValueError("PRIMARY_KEYWORD is required")
                secondary_keywords = os.getenv("SECONDARY_KEYWORDS", [])
                if secondary_keywords:
                    try:
                        secondary_keywords = json.loads(secondary_keywords)
                    except json.JSONDecodeError:
                        raise ValueError(
                            "SECONDARY_KEYWORDS must be a valid JSON array of strings"
                        )
                res = keyword_density_metrics_tool(
                    title, content, primary_keyword, secondary_keywords
                )
            case "KeywordsSuggestions":
                content = os.getenv("CONTENT")
                res = keywords_suggestions_tool(content)
            case "ReadabilityMetrics":
                content = os.getenv("CONTENT")
                res = readability_metrics_tool(content)
            case "LongTailKeywords":
                seed_keyword = os.getenv("SEED_KEYWORD")
                num_suggestions = int(os.getenv("NUM_SUGGESTIONS", 5))
                res = google_search_suggestions_tool(seed_keyword, num_suggestions)
            case _:
                print(f"Unknown command: {command}")
                sys.exit(1)

        print(json.dumps(res, indent=4))
    except Exception as e:
        print(f"Running command: {' '.join(sys.argv)} failed. Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
