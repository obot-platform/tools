from typing import List, Dict, Optional, Any
from typing_extensions import Annotated
from pydantic import BaseModel, Field
from datetime import datetime


# Knowledge Set schemas
class KnowledgeSetCreate(BaseModel):
    knowledge_set_id: str


class KnowledgeSetInfo(BaseModel):
    knowledge_set_id: str
    created_at: datetime


# File-level metadata
class FileMetadata(BaseModel):
    filename: str
    content_type: Optional[str] = None
    text: Optional[str] = None
    content_hash: Optional[str] = None  # SHA-256 hash of original content
    version: int = 1  # Version number (1, 2, 3, etc.)
    previous_version_file_id: Optional[str] = None  # Reference to previous version
    is_latest_version: bool = True  # Whether this is the latest version
    created_at: datetime = Field(default_factory=datetime.now)
    extra: Dict[str, Any] = Field(default_factory=dict)


class FileCreate(BaseModel):
    file_id: str
    metadata: FileMetadata


class FileInfo(BaseModel):
    file_id: str
    metadata: FileMetadata
    created_at: datetime


# Chunk-level metadata including text
class ChunkMetadata(BaseModel):
    text: str  # content of the chunk
    offset: Optional[int] = None  # chunk sequence or offset
    extra: Dict[str, Any] = Field(default_factory=dict)


class ChunkUpsert(BaseModel):
    chunk_id: str
    embedding: Annotated[List[float], Field(min_items=1)]
    metadata: ChunkMetadata


class ChunkInfo(BaseModel):
    chunk_id: str
    score: float
    metadata: ChunkMetadata


class QueryRequest(BaseModel):
    embedding: Annotated[List[float], Field(min_items=1)]
    top_k: int = 5


class QueryResult(BaseModel):
    file_id: str
    chunk_id: str
    score: float
    metadata: ChunkMetadata


class FileUploadResponse(BaseModel):
    file_id: str
    filename: str
    chunks_created: int
    message: str
    is_duplicate: bool = False  # Whether this was a duplicate file
    existing_file_id: Optional[str] = None  # If duplicate, the ID of existing file
