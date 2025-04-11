import os
import asyncio
from tools.reader import read_file
from tools.gptscript_workspace import write_file_in_workspace

async def main():
    input_file = os.getenv("INPUT_FILE", "")
    if not input_file:
        raise ValueError("Error: INPUT_FILE environment variable is not set")

    final_summary = await read_file(input_file)
    output_file = os.getenv("OUTPUT_FILE", "NONE")
    # Handle output
    if output_file.upper() == "NONE":
        print(final_summary)
    else:
        if output_file == "":
            directory, file_name = os.path.split(input_file)
            name, ext = os.path.splitext(file_name)
            summary_file_name = f"{name}_summary{ext}"
            output_file = os.path.join(directory, summary_file_name)

        try:
            await write_file_in_workspace(output_file, final_summary)
            print(f"Summary written to workspace file: {output_file}")
        except Exception as e:
            print(f"File Summary:\n{final_summary}")
            raise Exception(
                f"Failed to save summary to GPTScript workspace file {output_file}, Error: {e}"
            )


if __name__ == "__main__":
    asyncio.run(main())