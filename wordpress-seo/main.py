from scipy import optimize
from helper import setup_logger
import sys
from keywords_suggestions import keywords_suggestions
from content_optimize_metrics import content_optimize_metrics

logger = setup_logger(__name__)

def main():
    if len(sys.argv) < 2:
        print("Usage: python main.py <command>")
        sys.exit(1)

    command = sys.argv[1]
    try:
        match command:
            case "optimize":
                content_optimize_metrics()
            case "KeywordsSuggestions":
                keywords_suggestions()
            case _:
                print(f"Unknown command: {command}")
                sys.exit(1)
    except Exception as e:
        print(f"Running command: {' '.join(sys.argv)} failed. Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()