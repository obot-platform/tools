#!/usr/bin/env python3
from tools.helper import setup_logger
import gptscript
import os
import json
from gptscript.gptscript import Options
from tools.gptscript_workspace import read_file_in_workspace

logger = setup_logger(__name__)

SUPPORTED_KNOWLEDGE_DOC_FILE_TYPES = (
    ".pdf",
    ".pptx",
    ".ppt",
    ".html",
    ".css",
    ".md",
    ".txt",
    ".docx",
    ".doc",
    ".odt",
    ".rtf",
    ".csv",
    ".ipynb",
    ".json",
    ".cpp",
    ".c",
    ".go",
    ".java",
    ".js",
    ".py",
    ".ts",
)

MAX_FILE_SIZE = 100_000_000


async def load_from_knowledge_tool(input_file: str) -> str:
    """Load text from a workspace file using the knowledge-load tool.

    Args:
        input_file (str): The path to the file to load.

    Returns:
        str: The content of the file.
    """
    gptscript_client = gptscript.GPTScript()
    run = gptscript_client.run(
        "github.com/obot-platform/tools/knowledge/file-loader.gpt",
        Options(
            input=json.dumps({"input": input_file}),
            workspace=os.environ.get("GPTSCRIPT_WORKSPACE_ID"),
            env=[
                f"OPENAI_EMBEDDING_MODEL={os.environ.get('OBOT_DEFAULT_TEXT_EMBEDDING_MODEL')}",
            ],
        ),
    )
    text = await run.text()
    return text

async def load_text_from_workspace_file(file_path: str, max_file_size: int = MAX_FILE_SIZE) -> str:
    """Logic to load text from a workspace file.

    Args:
        file_path (str): The path to the file to load.
        max_file_size (int): The maximum file size to load. defaults to 100MB

    Raises:
        ValueError: If the file is not found in the workspace.
        ValueError: If the file is not a supported knowledge doc file type.
        ValueError: If the file content is not a valid UTF-8 encoded string.

    Returns:
        str: The content of the file.
    """

    # first read from gptscript workspace
    try:
        file_content: bytes = await read_file_in_workspace(file_path)
    except Exception as e:
        logger.error(
            f"Failed to load file from GPTScript workspace file {file_path}, Error: {e}"
        )
        raise ValueError(
            f"Failed to load file from GPTScript workspace file {file_path}, Error: {e}"
        )
    if len(file_content) > max_file_size:
        raise Exception(
            f"File size exceeds {max_file_size} bytes"
        )

    # if the file is not a supported knowledge doc file type, try to decode it as a plain text file using utf-8 encoding
    if not file_path.endswith(SUPPORTED_KNOWLEDGE_DOC_FILE_TYPES):
        try:
            file_content = file_content.decode("utf-8")
            return file_content
        except UnicodeDecodeError:
            logger.error(
                f"Failed to decode file content from GPTScript workspace file {file_path}, Error: {e}"
            )

    # if the file is a supported knowledge doc file type, or the file is not a plain text file, try to load it using the knowledge-load tool
    try:
        return await load_from_knowledge_tool(file_path)
    except Exception as e:
        logger.error(
            f"Failed to load file from GPTScript workspace file {file_path}, Error: {e}"
        )
        raise ValueError(
            f"Failed to load file from GPTScript workspace file {file_path}, Error: {e}"
        )
