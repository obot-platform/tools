#!/usr/bin/env python3
import asyncio
import os
import sys
import json
import signal
import contextlib
from pathlib import Path
from tools import file_operations

FILES_DIR = "files"

FILE_ENV = os.getenv("FILENAME", "")
DIR_ENV = os.getenv("DIR", "")

async def main():
    if len(sys.argv) == 1:
        print("""
Subcommands: read, write, copy, input, list
env: FILENAME, CONTENT, TO_FILENAME, GPTSCRIPT_WORKSPACE_DIR
Usage: python script.py <command>
        """)
        return
    
    cmd = sys.argv[1]
    if cmd == "read" and (not FILE_ENV or FILE_ENV.endswith("/")):
        cmd = "list"
    
    match cmd:
        case "input":
            await file_operations.input_handler()
        case "list":
            await file_operations.list_files(DIR_ENV)
        case "read":
            await file_operations.read_file(FILE_ENV)
        case "write":
            content = os.getenv("CONTENT", "")
            await file_operations.write_file(FILE_ENV, content)
        case "copy":
            to_filename = os.getenv("TO_FILENAME", "")
            await file_operations.copy_file(FILE_ENV, to_filename)
        case _:
            print("Unknown command")

if __name__ == "__main__":
    asyncio.run(main())
