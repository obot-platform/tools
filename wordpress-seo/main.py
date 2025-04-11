from scipy import optimize
from helper import setup_logger
import sys
from keywords_suggestions import keywords_suggestions
from content_optimize_metrics import content_optimize_metrics
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
            case "ContentOptimizeMetrics":
                content = os.getenv("CONTENT")
                primary_keyword = os.getenv("PRIMARY_KEYWORD")
                if not primary_keyword:
                    raise ValueError("PRIMARY_KEYWORD is required")
                secondary_keywords = os.getenv("SECONDARY_KEYWORDS", [])
                if secondary_keywords:
                    try:
                        secondary_keywords = json.loads(secondary_keywords)
                    except json.JSONDecodeError:
                        raise ValueError("SECONDARY_KEYWORDS must be a valid JSON array of strings")
                content_optimize_metrics(content, primary_keyword, secondary_keywords)
            case "KeywordsSuggestions":
                content = os.getenv("CONTENT")
                keywords_suggestions(content)
            case _:
                print(f"Unknown command: {command}")
                sys.exit(1)
    except Exception as e:
        print(f"Running command: {' '.join(sys.argv)} failed. Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()