#!/usr/bin/env python3
from tools.load_text import load_text_from_workspace_file
from tools.helper import setup_logger, get_openai_client

import tiktoken

logger = setup_logger(__name__)

TIKTOKEN_MODEL = "gpt-4o"
enc = tiktoken.encoding_for_model(TIKTOKEN_MODEL)
TOKEN_THRESHOLD = 10000

MAX_FILE_SIZE = 100_000_000

async def read_file(input_file: str, max_file_size: int = MAX_FILE_SIZE):
    """the enhanced workspace_read tool
    This tool reads a file from the GPTScript workspace and returns the file content.
    If the file has too many tokens, it summarizes the file content and returns the summary instead.

    Raises:
        ValueError: If the INPUT_FILE environment variable is not set
        Exception: If the file content is not a valid UTF-8 encoded string
    """

    from tools.summarizer import (
        DocumentSummarizer,
        MODEL,
        MAX_CHUNK_TOKENS,
        MAX_WORKERS,
    )

    logger.info(f"Input file: {input_file}")
    if not input_file:
        raise ValueError("Error: INPUT_FILE environment variable is not set")

    file_content: str = await load_text_from_workspace_file(input_file, max_file_size)
    tokens = enc.encode(file_content)

    # if the file has too many tokens, summarize it and return the summary
    if len(tokens) > TOKEN_THRESHOLD:
        response_str = f"The original file {input_file} contains too many tokens ({len(tokens)}), summarizing it...\n"
        summarizer = DocumentSummarizer(
            get_openai_client(),
            model=MODEL,
            max_chunk_tokens=MAX_CHUNK_TOKENS,
            max_workers=MAX_WORKERS,
        )
        try:
            final_summary: str = summarizer.summarize(file_content)
        except Exception as e:
            logger.error(f"Summarization failed: {e}")
            raise Exception(f"ERROR: Summarization failed: {e}")

        response_str += f"Here is the summary of the file {input_file}'s content:\n\n{final_summary}"
        return response_str

    # if the file has less than TOKEN_THRESHOLD tokens, directly return the file content
    else:
        return file_content