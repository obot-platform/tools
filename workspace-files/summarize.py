#!/usr/bin/env python3
import os
import asyncio
from tools.summarizer import DocumentSummarizer, MODEL, MAX_CHUNK_TOKENS, MAX_WORKERS
from tools.load_text import load_text_from_file
from tools.helper import save_to_gptscript_workspace, get_openai_client

async def main():
    input_file = os.getenv("INPUT_FILE", "")
    if not input_file:
        raise ValueError("Error: INPUT_FILE environment variable is not set")

    file_content = await load_text_from_file(input_file)

    if len(file_content) == 0:
        print("Warning: File is empty, skipping summarization")
        return

    output_file = os.getenv("OUTPUT_FILE", "")

    summarizer = DocumentSummarizer(
        get_openai_client(),
        model=MODEL,
        max_chunk_tokens=MAX_CHUNK_TOKENS,
        max_workers=MAX_WORKERS,
    )

    try:
        final_summary = summarizer.summarize(file_content)
    except Exception as e:
        raise Exception(f"ERROR: Summarization failed: {e}")

    # Handle output
    if output_file.upper() == "NONE":
        print(final_summary)
    else:
        if output_file == "":
            directory, file_name = os.path.split(input_file)
            name, ext = os.path.splitext(file_name)
            summary_file_name = f"{name}_summary.md"
            output_file = os.path.join(directory, summary_file_name)

        try:
            await save_to_gptscript_workspace(output_file, final_summary)
            print(f"Summary written to workspace file: {output_file}")
        except Exception as e:
            print(f"File Summary:\n{final_summary}")
            raise Exception(
                f"Failed to save summary to GPTScript workspace file {output_file}, Error: {e}"
            )


if __name__ == "__main__":
    asyncio.run(main())