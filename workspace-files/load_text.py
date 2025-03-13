from helper import load_from_gptscript_workspace, setup_logger, load_from_knowledge_tool

logger = setup_logger(__name__)


SUPPORTED_TEXT_FILE_TYPES = (
    ".md",
    ".txt",
    ".markdown",
    ".text",
    ".mdx",
    ".mdtxt",
    ".mdtxtx",
)
SUPPORTED_KNOWLEDGE_DOC_FILE_TYPES = (
    ".docx",
    ".doc",
    ".pdf",
    ".pptx",
    ".ppt",
    ".html",
    ".htm",
    ".ipynb",
)
ALL_SUPPORTED_FILE_TYPES = (
    SUPPORTED_TEXT_FILE_TYPES + SUPPORTED_KNOWLEDGE_DOC_FILE_TYPES
)


async def load_text_from_file(file_path: str) -> str:

    if not file_path.endswith(ALL_SUPPORTED_FILE_TYPES):
        raise ValueError(
            f"Error: the input file must end with one of the following file types: {ALL_SUPPORTED_FILE_TYPES}, other file types are not supported yet."
        )

    try:
        file_content = await load_from_gptscript_workspace(file_path)
    except Exception as e:
        logger.error(
            f"Failed to load file from GPTScript workspace file {file_path}, Error: {e}"
        )
        raise ValueError(
            f"Failed to load file from GPTScript workspace file {file_path}, Error: {e}"
        )

    if file_path.endswith(SUPPORTED_TEXT_FILE_TYPES):
        return file_content.decode("utf-8")
    elif file_path.endswith(SUPPORTED_KNOWLEDGE_DOC_FILE_TYPES):
        return await load_from_knowledge_tool(file_path)
    else:
        raise ValueError(f"Unsupported file type: {file_path}")
