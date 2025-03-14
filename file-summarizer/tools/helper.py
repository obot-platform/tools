#!/usr/bin/env python3
import os
import logging
import sys
from openai import OpenAI


def setup_logger(name):
    """Setup a logger that writes to sys.stderr. This will eventually show up in GPTScript's debugging logs.

    Args:
        name (str): The name of the logger.

    Returns:
        logging.Logger: The logger.
    """
    # Create a logger
    logger = logging.getLogger(name)
    logger.setLevel(logging.DEBUG)  # Set the logging level

    # Create a stream handler that writes to sys.stderr
    stderr_handler = logging.StreamHandler(sys.stderr)

    # Create a log formatter
    formatter = logging.Formatter(
        "[File Summarizer Debugging Log]: %(asctime)s - %(name)s - %(levelname)s - %(message)s"
    )
    stderr_handler.setFormatter(formatter)

    # Add the handler to the logger
    logger.addHandler(stderr_handler)

    return logger


logger = setup_logger(__name__)


def get_openai_client() -> OpenAI:
    # Check for OPENAI_API_KEY
    if "OPENAI_API_KEY" not in os.environ:
        sys.exit(
            "ERROR: OPENAI_API_KEY environment variable not found.\n"
            "Please set it before running the script, e.g.:\n\n"
            "  export OPENAI_API_KEY='sk-xxxxxxx'\n"
        )
    try:
        base_url = os.getenv("OPENAI_BASE_URL", "https://api.openai.com/v1")
        logger.debug(f"Using base_url: {base_url}")
        api_key = os.environ["OPENAI_API_KEY"]
        client = OpenAI(base_url=base_url, api_key=api_key)
        return client
    except Exception as e:
        raise Exception(f"ERROR: Failed to initialize OpenAI client: {e}")
