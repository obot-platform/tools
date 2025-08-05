import os
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker, declarative_base
from sqlalchemy import Column, String, JSON, TIMESTAMP, text, select, delete
from sqlalchemy.dialects.postgresql import insert as pg_insert
from pgvector.sqlalchemy import Vector
from datetime import datetime
from typing import Optional

# Import configuration from centralized config
from app.config import DATABASE_URL, EMBEDDING_DIMENSION

# Database setup
engine = create_async_engine(DATABASE_URL, echo=False)
AsyncSessionLocal = sessionmaker(
    bind=engine, class_=AsyncSession, expire_on_commit=False
)
Base = declarative_base()


# models
class KnowledgeSet(Base):
    __tablename__ = "knowledge_sets"
    user_id = Column(String, primary_key=True)
    knowledge_set_id = Column(String, primary_key=True)
    created_at = Column(
        TIMESTAMP(timezone=True), server_default=text("now()"), nullable=False
    )


class FileRecord(Base):
    __tablename__ = "files"
    user_id = Column(String, primary_key=True)
    knowledge_set_id = Column(String, primary_key=True)
    file_id = Column(String, primary_key=True)
    file_metadata = Column(JSON, nullable=False)
    created_at = Column(
        TIMESTAMP(timezone=True), server_default=text("now()"), nullable=False
    )


class ChunkEntry(Base):
    __tablename__ = "chunks"
    user_id = Column(String, primary_key=True)
    knowledge_set_id = Column(String, primary_key=True)
    file_id = Column(String, primary_key=True)
    chunk_id = Column(String, primary_key=True)
    embedding = Column(Vector(EMBEDDING_DIMENSION), nullable=False)
    chunk_metadata = Column(JSON, nullable=False)
    created_at = Column(
        TIMESTAMP(timezone=True), server_default=text("now()"), nullable=False
    )


# init database
async def init_db():
    async with engine.begin() as conn:
        # ensure pgvector extension is enabled
        await conn.execute(text("CREATE EXTENSION IF NOT EXISTS vector;"))
        # create tables and index
        await conn.run_sync(Base.metadata.create_all)
        await conn.execute(
            text(
                "CREATE INDEX IF NOT EXISTS idx_chunks_embedding "
                "ON chunks USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);"
            )
        )


# validation helpers
async def validate_knowledge_set_exists(
    session: AsyncSession, user_id: str, knowledge_set_id: str
) -> bool:
    """Check if a knowledge set exists for the given user."""
    result = await session.execute(
        select(KnowledgeSet).where(
            (KnowledgeSet.user_id == user_id)
            & (KnowledgeSet.knowledge_set_id == knowledge_set_id)
        )
    )
    return result.first() is not None


async def validate_file_exists(
    session: AsyncSession, user_id: str, knowledge_set_id: str, file_id: str
) -> bool:
    """Check if a file exists in the given knowledge set for the user."""
    result = await session.execute(
        select(FileRecord).where(
            (FileRecord.user_id == user_id)
            & (FileRecord.knowledge_set_id == knowledge_set_id)
            & (FileRecord.file_id == file_id)
        )
    )
    return result.first() is not None


# knowledge set CRUD
async def create_knowledge_set(user_id: str, knowledge_set_id: str):
    async with AsyncSessionLocal() as session:
        stmt = pg_insert(KnowledgeSet).values(
            user_id=user_id, knowledge_set_id=knowledge_set_id
        )
        stmt = stmt.on_conflict_do_nothing()
        await session.execute(stmt)
        await session.commit()


async def list_knowledge_sets(user_id: str):
    async with AsyncSessionLocal() as session:
        res = await session.execute(
            select(KnowledgeSet).where(KnowledgeSet.user_id == user_id)
        )
        return res.scalars().all()


async def delete_knowledge_set(user_id: str, knowledge_set_id: str):
    async with AsyncSessionLocal() as session:
        await session.execute(
            delete(ChunkEntry).where(
                (ChunkEntry.user_id == user_id)
                & (ChunkEntry.knowledge_set_id == knowledge_set_id)
            )
        )
        await session.execute(
            delete(FileRecord).where(
                (FileRecord.user_id == user_id)
                & (FileRecord.knowledge_set_id == knowledge_set_id)
            )
        )
        await session.execute(
            delete(KnowledgeSet).where(
                (KnowledgeSet.user_id == user_id)
                & (KnowledgeSet.knowledge_set_id == knowledge_set_id)
            )
        )
        await session.commit()


# file CRUD
async def find_file_by_content_hash(
    user_id: str, knowledge_set_id: str, content_hash: str
) -> Optional[tuple]:
    """Find a file by its content hash. Returns (file_id, file_metadata) if found."""
    async with AsyncSessionLocal() as session:
        # Validate knowledge set exists
        if not await validate_knowledge_set_exists(session, user_id, knowledge_set_id):
            raise ValueError(f"Knowledge set '{knowledge_set_id}' not found for user")

        result = await session.execute(
            select(FileRecord.file_id, FileRecord.file_metadata).where(
                (FileRecord.user_id == user_id)
                & (FileRecord.knowledge_set_id == knowledge_set_id)
                & (FileRecord.file_metadata.op("->>")("content_hash") == content_hash)
            )
        )
        row = result.first()
        return (row.file_id, row.file_metadata) if row else None


async def find_file_by_filename(
    user_id: str, knowledge_set_id: str, filename: str
) -> Optional[tuple]:
    """Find a file by its filename. Returns (file_id, file_metadata) if found."""
    async with AsyncSessionLocal() as session:
        # Validate knowledge set exists
        if not await validate_knowledge_set_exists(session, user_id, knowledge_set_id):
            raise ValueError(f"Knowledge set '{knowledge_set_id}' not found for user")

        result = await session.execute(
            select(FileRecord.file_id, FileRecord.file_metadata).where(
                (FileRecord.user_id == user_id)
                & (FileRecord.knowledge_set_id == knowledge_set_id)
                & (FileRecord.file_metadata.op("->>")("filename") == filename)
            )
        )
        row = result.first()
        return (row.file_id, row.file_metadata) if row else None


async def update_file_metadata(
    user_id: str, knowledge_set_id: str, file_id: str, metadata: dict
):
    """Update file metadata for an existing file."""
    async with AsyncSessionLocal() as session:
        # Validate knowledge set exists
        if not await validate_knowledge_set_exists(session, user_id, knowledge_set_id):
            raise ValueError(f"Knowledge set '{knowledge_set_id}' not found for user")

        # Update the file metadata
        stmt = pg_insert(FileRecord).values(
            user_id=user_id,
            knowledge_set_id=knowledge_set_id,
            file_id=file_id,
            file_metadata=metadata,
        )
        stmt = stmt.on_conflict_do_update(
            index_elements=[
                FileRecord.user_id,
                FileRecord.knowledge_set_id,
                FileRecord.file_id,
            ],
            set_={"file_metadata": stmt.excluded.file_metadata},
        )
        await session.execute(stmt)
        await session.commit()


async def mark_previous_version_as_old(
    user_id: str, knowledge_set_id: str, previous_file_id: str
):
    """Mark a previous version as no longer the latest and delete its chunks."""
    async with AsyncSessionLocal() as session:
        # Get the current file metadata
        result = await session.execute(
            select(FileRecord.file_metadata).where(
                (FileRecord.user_id == user_id)
                & (FileRecord.knowledge_set_id == knowledge_set_id)
                & (FileRecord.file_id == previous_file_id)
            )
        )
        row = result.first()
        if not row:
            return False

        # Update the metadata to mark as not latest
        current_metadata = row.file_metadata
        current_metadata["is_latest_version"] = False

        # Update the file record
        await update_file_metadata(
            user_id, knowledge_set_id, previous_file_id, current_metadata
        )

        # Delete chunks for this old version
        await session.execute(
            delete(ChunkEntry).where(
                (ChunkEntry.user_id == user_id)
                & (ChunkEntry.knowledge_set_id == knowledge_set_id)
                & (ChunkEntry.file_id == previous_file_id)
            )
        )
        await session.commit()
        return True


async def get_latest_version_info(
    user_id: str, knowledge_set_id: str, filename: str
) -> Optional[tuple]:
    """Get the latest version info for a filename. Returns (file_id, metadata, version)."""
    async with AsyncSessionLocal() as session:
        # Validate knowledge set exists
        if not await validate_knowledge_set_exists(session, user_id, knowledge_set_id):
            raise ValueError(f"Knowledge set '{knowledge_set_id}' not found for user")

        result = await session.execute(
            select(FileRecord.file_id, FileRecord.file_metadata).where(
                (FileRecord.user_id == user_id)
                & (FileRecord.knowledge_set_id == knowledge_set_id)
                & (FileRecord.file_metadata.op("->>")("filename") == filename)
                & (FileRecord.file_metadata.op("->>")("is_latest_version") == "true")
            )
        )
        row = result.first()
        if not row:
            return None

        metadata = row.file_metadata
        version = metadata.get("version", 1)
        return (row.file_id, metadata, version)


async def create_file(
    user_id: str, knowledge_set_id: str, file_id: str, metadata: dict
):
    async with AsyncSessionLocal() as session:
        # Validate knowledge set exists
        if not await validate_knowledge_set_exists(session, user_id, knowledge_set_id):
            raise ValueError(f"Knowledge set '{knowledge_set_id}' not found for user")

        stmt = (
            pg_insert(FileRecord)
            .values(
                user_id=user_id,
                knowledge_set_id=knowledge_set_id,
                file_id=file_id,
                file_metadata=metadata,
            )
            .on_conflict_do_nothing()
        )
        await session.execute(stmt)
        await session.commit()


async def list_files(user_id: str, knowledge_set_id: str):
    async with AsyncSessionLocal() as session:
        # Validate knowledge set exists
        if not await validate_knowledge_set_exists(session, user_id, knowledge_set_id):
            raise ValueError(f"Knowledge set '{knowledge_set_id}' not found for user")

        res = await session.execute(
            select(
                FileRecord.file_id, FileRecord.file_metadata, FileRecord.created_at
            ).where(
                (FileRecord.user_id == user_id)
                & (FileRecord.knowledge_set_id == knowledge_set_id)
            )
        )
        return res.all()


async def delete_file(user_id: str, knowledge_set_id: str, file_id: str):
    async with AsyncSessionLocal() as session:
        # First check if the knowledge set exists for this user
        if not await validate_knowledge_set_exists(session, user_id, knowledge_set_id):
            raise ValueError(f"Knowledge set '{knowledge_set_id}' not found for user")

        # Check if file exists before attempting to delete
        file_exists = await session.execute(
            select(FileRecord).where(
                (FileRecord.user_id == user_id)
                & (FileRecord.knowledge_set_id == knowledge_set_id)
                & (FileRecord.file_id == file_id)
            )
        )
        if not file_exists.first():
            return False  # File doesn't exist

        # Delete chunks first
        chunks_deleted = await session.execute(
            delete(ChunkEntry).where(
                (ChunkEntry.user_id == user_id)
                & (ChunkEntry.knowledge_set_id == knowledge_set_id)
                & (ChunkEntry.file_id == file_id)
            )
        )

        # Delete file record
        file_deleted = await session.execute(
            delete(FileRecord).where(
                (FileRecord.user_id == user_id)
                & (FileRecord.knowledge_set_id == knowledge_set_id)
                & (FileRecord.file_id == file_id)
            )
        )

        await session.commit()
        return file_deleted.rowcount > 0


# chunk CRUD
async def count_chunks_for_file(
    user_id: str, knowledge_set_id: str, file_id: str
) -> int:
    """Count the number of chunks for a specific file."""
    async with AsyncSessionLocal() as session:
        result = await session.execute(
            select(ChunkEntry).where(
                (ChunkEntry.user_id == user_id)
                & (ChunkEntry.knowledge_set_id == knowledge_set_id)
                & (ChunkEntry.file_id == file_id)
            )
        )
        return len(result.all())


async def upsert_chunks(
    user_id: str, knowledge_set_id: str, file_id: str, chunks: list
):
    async with AsyncSessionLocal() as session:
        stmt = pg_insert(ChunkEntry)
        vals = []
        for chunk in chunks:
            vals.append(
                {
                    "user_id": user_id,
                    "knowledge_set_id": knowledge_set_id,
                    "file_id": file_id,
                    "chunk_id": chunk.chunk_id,
                    "embedding": chunk.embedding,
                    "chunk_metadata": chunk.metadata.dict(),
                }
            )
        stmt = stmt.values(vals)
        stmt = stmt.on_conflict_do_update(
            index_elements=[
                ChunkEntry.user_id,
                ChunkEntry.knowledge_set_id,
                ChunkEntry.file_id,
                ChunkEntry.chunk_id,
            ],
            set_={
                "embedding": stmt.excluded.embedding,
                "chunk_metadata": stmt.excluded.chunk_metadata,
            },
        )
        await session.execute(stmt)
        await session.commit()


async def delete_chunk(
    user_id: str, knowledge_set_id: str, file_id: str, chunk_id: str
):
    async with AsyncSessionLocal() as session:
        res = await session.execute(
            delete(ChunkEntry).where(
                (ChunkEntry.user_id == user_id)
                & (ChunkEntry.knowledge_set_id == knowledge_set_id)
                & (ChunkEntry.file_id == file_id)
                & (ChunkEntry.chunk_id == chunk_id)
            )
        )
        await session.commit()
        return res.rowcount > 0


async def query_chunks(
    user_id: str, knowledge_set_id: str, embedding: list, top_k: int
):
    async with AsyncSessionLocal() as session:
        q = (
            select(
                ChunkEntry.file_id,
                ChunkEntry.chunk_id,
                ChunkEntry.chunk_metadata,
                (1 - ChunkEntry.embedding.cosine_distance(embedding)).label("score"),
            )
            .where(
                (ChunkEntry.user_id == user_id)
                & (ChunkEntry.knowledge_set_id == knowledge_set_id)
            )
            .order_by(text("score DESC"))  # DESC because higher is now better
            .limit(top_k)
        )
        res = await session.execute(q)
        return res.all()
