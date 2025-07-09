import asyncio
from datetime import datetime
import app.vector_db as db
import app.db_schema as schemas
import app.text_processing as text_proc
import app.embeddings as embeddings

from fastmcp import FastMCP
from pydantic import Field
from typing import Annotated
from fastmcp.exceptions import ToolError
from uuid import uuid4
import hashlib
import base64
from fastmcp.server.dependencies import get_http_headers
from app.config import PORT, MCP_PATH

mcp = FastMCP(
    name="KnowledgeMCPServer",
    on_duplicate_tools="error",
    on_duplicate_resources="warn",
    on_duplicate_prompts="replace",
)


def _get_user_id() -> str:
    headers = get_http_headers()
    user_id = headers.get("x-forwarded-user", None)
    if not user_id:
        raise ToolError("No user id found in headers")
    return user_id


# Use the embedding service
generate_embedding = embeddings.generate_embedding


## manage knowledge set tools
@mcp.tool(
    name="create_knowledge_set",
)
async def create_knowledge_set() -> schemas.KnowledgeSetInfo:
    """Create a new knowledge set for the user"""
    user_id = _get_user_id()
    knowledge_set_id = str(uuid4())
    await db.create_knowledge_set(user_id, knowledge_set_id)
    return schemas.KnowledgeSetInfo(
        knowledge_set_id=knowledge_set_id, created_at=datetime.now()
    )


@mcp.tool(
    name="list_knowledge_sets",
)
async def list_knowledge_sets() -> list:
    """List all knowledge sets for the user"""
    user_id = _get_user_id()
    ks = await db.list_knowledge_sets(user_id)
    return [schemas.KnowledgeSetInfo(**k.__dict__) for k in ks]


@mcp.tool(
    name="delete_knowledge_set",
)
async def delete_knowledge_set(
    knowledge_set_id: Annotated[
        str, Field(description="The knowledge set ID to delete")
    ],
) -> dict:
    """Delete a knowledge set and all its data"""
    user_id = _get_user_id()
    await db.delete_knowledge_set(user_id, knowledge_set_id)
    return {"detail": "knowledge set and related data deleted"}


## client tools
@mcp.tool(name="ingest_file")
async def ingest_file(
    knowledge_set_id: Annotated[
        str, Field(description="The knowledge set ID to ingest the file into")
    ],
    filename: Annotated[str, Field(description="The name of the file")],
    content: Annotated[
        str, Field(description="The base64-encoded content of the file")
    ],
) -> schemas.FileUploadResponse:
    """
    Upload and process a file: extract text, chunk it, generate embeddings, and store in the database.
    """
    user_id = _get_user_id()

    try:
        file_extension = filename.split(".")[-1]

        # Decode base64 content to bytes
        try:
            content_bytes = base64.b64decode(content)
        except Exception as e:
            raise ToolError(f"Failed to decode base64 content to bytes: {e}")

        # Generate content hash for duplicate detection
        content_hash = hashlib.sha256(content_bytes).hexdigest()

        # Check for exact duplicate (same content hash)
        existing_file_by_hash = await db.find_file_by_content_hash(
            user_id, knowledge_set_id, content_hash
        )

        if existing_file_by_hash:
            # Exact duplicate found - return existing file info
            existing_file_id, existing_metadata = existing_file_by_hash
            existing_chunks = await db.count_chunks_for_file(
                user_id, knowledge_set_id, existing_file_id
            )

            return schemas.FileUploadResponse(
                file_id=existing_file_id,
                filename=filename,
                chunks_created=existing_chunks,
                message=f"File '{filename}' is identical to existing file. Using existing version.",
                is_duplicate=True,
                existing_file_id=existing_file_id,
            )

        # Check for latest version of this filename
        latest_version_info = await db.get_latest_version_info(
            user_id, knowledge_set_id, filename
        )

        if latest_version_info:
            # This is a new version of an existing file
            previous_file_id, previous_metadata, previous_version = latest_version_info
            new_version = previous_version + 1

            # Mark previous version as old and delete its chunks
            await db.mark_previous_version_as_old(
                user_id, knowledge_set_id, previous_file_id
            )

            version_message = f"New version {new_version} of '{filename}' created. Previous version {previous_version} chunks removed."
        else:
            # This is a completely new file
            new_version = 1
            previous_file_id = None
            version_message = (
                f"New file '{filename}' (version 1) processed successfully."
            )

        # Generate unique file ID for new version
        file_id = str(uuid4())

        # Extract text from file content
        md_converter_result = text_proc.extract_text_from_content(
            content_bytes, file_extension
        )
        extracted_title, extracted_text = (
            md_converter_result.title,
            md_converter_result.markdown,
        )

        # Create file metadata with version information
        file_metadata = schemas.FileMetadata(
            filename=filename,
            text=extracted_text,
            content_hash=content_hash,
            version=new_version,
            previous_version_file_id=previous_file_id,
            is_latest_version=True,
            created_at=datetime.now(),
            extra={
                "original_size": len(content_bytes),
                "processed_size": len(extracted_text),
            },
        )

        # Store file metadata
        await db.create_file(
            user_id, knowledge_set_id, file_id, file_metadata.model_dump(mode="json")
        )

        # Chunk the text
        text_chunks = text_proc.chunk_text(extracted_text)

        # Generate embeddings and create chunk objects
        chunks_to_upsert = []
        try:
            for i, (chunk_text, offset) in enumerate(text_chunks):
                # Generate embedding for this chunk
                embedding = await generate_embedding(chunk_text)

                # Create chunk metadata
                chunk_metadata = schemas.ChunkMetadata(
                    text=chunk_text,
                    offset=offset,
                    extra={"chunk_index": i, "chunk_length": len(chunk_text)},
                )

                # Create chunk upsert object
                chunk_upsert = schemas.ChunkUpsert(
                    chunk_id=f"{file_id}_chunk_{i}",
                    embedding=embedding,
                    metadata=chunk_metadata,
                )
                chunks_to_upsert.append(chunk_upsert)
        except Exception as e:
            error_str = str(e)
            if "OpenAI API error" in error_str:
                if "502" in error_str or "503" in error_str or "504" in error_str:
                    raise ToolError(
                        f"OpenAI API is temporarily unavailable (server error). Please try again in a few minutes. Details: {error_str}"
                    )
                elif "429" in error_str:
                    raise ToolError(
                        f"OpenAI API rate limit exceeded. Please try again later. Details: {error_str}"
                    )
                elif "timeout" in error_str.lower():
                    raise ToolError(
                        f"OpenAI API request timed out. Please try again. Details: {error_str}"
                    )
                else:
                    raise ToolError(f"OpenAI API error: {error_str}")
            else:
                raise ToolError(f"Failed to create embeddings: {error_str}")

        # Store all chunks
        if chunks_to_upsert:
            await db.upsert_chunks(user_id, knowledge_set_id, file_id, chunks_to_upsert)

        return schemas.FileUploadResponse(
            file_id=file_id,
            filename=filename,
            chunks_created=len(chunks_to_upsert),
            message=version_message,
            is_duplicate=False,
            existing_file_id=previous_file_id,
        )

    except Exception as e:
        raise ToolError(f"Failed to process file: {e}")


@mcp.tool(name="list_files")
async def list_files(
    knowledge_set_id: Annotated[
        str, Field(description="The knowledge set ID to list files from")
    ],
) -> list[schemas.FileInfo]:
    """List all files for the specified knowledge set."""
    user_id = _get_user_id()
    try:
        files = await db.list_files(user_id, knowledge_set_id)
        if not files:
            return []
        return [
            schemas.FileInfo(
                file_id=f.file_id,
                metadata=schemas.FileMetadata(**f.file_metadata),
                created_at=f.created_at,
            )
            for f in files
        ]
    except ValueError as e:
        raise ToolError(str(e))


@mcp.tool(name="remove_file")
async def delete_file(
    knowledge_set_id: Annotated[
        str, Field(description="The knowledge set ID containing the file")
    ],
    file_id: str,
) -> dict:
    """Delete a file and all its chunks."""
    user_id = _get_user_id()
    try:
        ok = await db.delete_file(user_id, knowledge_set_id, file_id)
        if not ok:
            return {
                "detail": f"File '{file_id}' was not found (may have been already deleted)",
                "deleted": False,
            }
        return {"detail": "file deleted", "deleted": True}
    except ValueError as e:
        raise ToolError(str(e))


@mcp.tool(name="query")
async def text_query(
    knowledge_set_id: Annotated[
        str, Field(description="The knowledge set ID to query")
    ],
    query_text: Annotated[
        str, Field(description="The text to query the knowledge base")
    ],
) -> list[schemas.QueryResult]:
    top_k = 5
    """Given a query text, return the top 5 most relevant chunks from the knowledge base"""
    user_id = _get_user_id()

    # Generate embedding for the query text
    query_embedding = await generate_embedding(query_text)

    # Query the database
    rows = await db.query_chunks(user_id, knowledge_set_id, query_embedding, top_k)
    return [
        schemas.QueryResult(
            file_id=r.file_id,
            chunk_id=r.chunk_id,
            score=r.score,
            metadata=schemas.ChunkMetadata(**r.chunk_metadata),
        )
        for r in rows
    ]


async def streamable_http_server():
    """Main entry point for the MCP server."""
    await db.init_db()
    await mcp.run_async(
        transport="streamable-http",  # fixed to streamable-http
        host="0.0.0.0",
        port=PORT,
        path=MCP_PATH,
    )


if __name__ == "__main__":
    asyncio.run(streamable_http_server())
