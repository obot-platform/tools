#!/usr/bin/env python3
from dataclasses import dataclass
from tools.load_text import load_text_from_workspace_file
from tools.helper import setup_logger, get_openai_client
from tools.summarizer import (
    DocumentSummarizer,
    MODEL,
    TIKTOKEN_MODEL,
    MAX_CHUNK_TOKENS,
    MAX_WORKERS,
)
import os
import tiktoken
import json
from pathlib import Path
from tools.gptscript_workspace import read_file_in_workspace, write_file_in_workspace, list_files_in_workspace

logger = setup_logger(__name__)

TIKTOKEN_MODEL = "gpt-4o"
enc = tiktoken.encoding_for_model(TIKTOKEN_MODEL)
TOKEN_THRESHOLD = 10000

FILE_ENV = os.getenv("FILENAME", "")
DIR_ENV = os.getenv("DIR", "")
MAX_FILE_SIZE = 250_000

UNSUPPORTED_FILE_TYPES = {".pdf", ".docx", ".doc", ".pptx", ".ppt", ".xlsx", ".xls", ".jpg", ".png", ".gif", ".mp3", ".mp4", ".zip", ".rar"}


@dataclass
class Explain:
    filename: str
    selection: str

@dataclass
class Data:
    prompt: str = ""
    explain: Explain = None
    improve: Explain = None
    changedFiles: dict = None



async def read_file():
    """the enhanced workspace_read tool
    This tool reads a file from the GPTScript workspace and returns the file content.
    If the file has too many tokens, it summarizes the file content and returns the summary instead.

    Raises:
        ValueError: If the INPUT_FILE environment variable is not set
        Exception: If the file content is not a valid UTF-8 encoded string
    """
    input_file = os.getenv("INPUT_FILE")
    logger.info(f"Input file: {input_file}")
    if not input_file:
        raise ValueError("Error: INPUT_FILE environment variable is not set")

    file_content : str = await load_text_from_workspace_file(input_file)
    tokens = enc.encode(file_content)

    # if the file has too many tokens, summarize it and return the summary
    if len(tokens) > TOKEN_THRESHOLD: 
        summarizer = DocumentSummarizer(
            get_openai_client(),
            model=MODEL,
            max_chunk_tokens=MAX_CHUNK_TOKENS,
            max_workers=MAX_WORKERS,
        )
        try:
            final_summary : str = summarizer.summarize(file_content)
        except Exception as e:
            logger.error(f"Summarization failed: {e}")
            raise Exception(f"ERROR: Summarization failed: {e}")

        response_str = f"The uploaded file {input_file} contains too many tokens ({len(tokens)}), here is the summary of the file content:\n\n{final_summary}"
        print(response_str)
    
    # if the file has less than TOKEN_THRESHOLD tokens, directly return the file content
    else:  
        print(file_content)
    

def in_back_ticks(s: str) -> str:
    return f"\n```\n{s}\n```\n"

async def input_handler():
    input_data = os.getenv("INPUT", "")
    try:
        data : Data = Data(**json.loads(input_data))
    except json.JSONDecodeError:
        print(input_data)
        return
    
    if data.explain:
        print(f'Explain the following selection from the "{data.explain.filename}" workspace file: {in_back_ticks(data.explain.selection)}')
    
    if data.improve:
        if data.improve.selection == "":
            print(f'Referring to the workspace file "{data.improve.filename}", {data.prompt} \nWrite any suggested changes back to the file')
        else:
            print(f'Referring to the selection from "{data.improve.filename}", {data.prompt}: {in_back_ticks(data.improve.selection)} \nWrite any suggested changes back to the file')
    
    if data.changedFiles:
        for filename, content in data.changedFiles.items():
            if await write_file_in_workspace(filename, content):
                if not printed:
                    printed = True
                    print("The following files have been externally changed in the workspace, re-read them if the up to date content needs to be known:")
                print(f"File: {filename}\n{in_back_ticks(content)}\n")
        
        print("")

    if data.prompt != "":
        print(data.prompt)

async def list_files(directory: str):
    print(await list_files_in_workspace(directory))

async def write_file(filename: str, content: str):
    ext = Path(filename).suffix.lower()
    if ext in UNSUPPORTED_FILE_TYPES:
        print(f"Writing to files with extension {ext} is not supported")
        return
    
    if await write_file_in_workspace(filename, content):
        print(f"Successfully wrote {len(content)} bytes to {filename}")
    else:
        print("Failed to write file to GPTScript workspace")

async def copy_file(src: str, dest: str):
    data : bytes = await read_file_in_workspace(src)
    if await write_file_in_workspace(dest, data):
        print(f"Successfully copied {len(data)} bytes from {src} to {dest}")
    else:
        print("Failed to copy file to GPTScript workspace")