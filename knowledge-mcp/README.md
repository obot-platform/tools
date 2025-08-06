# Knowledge MCP Server

A Model Context Protocol (MCP) server for intelligent knowledge management with vector embeddings and semantic search capabilities.

## 🚀 Features

- **📁 Knowledge Set Management** - Create, list, and delete isolated knowledge collections
- **📄 File Processing** - Support for multiple file formats (PDF, TXT, DOCX, etc.) using MarkItDown
- **🧠 Smart Chunking** - Intelligent text chunking with multiple strategies (sentence, semantic, recursive, token-based)
- **🔍 Vector Search** - Semantic search using OpenAI embeddings and PostgreSQL with pgvector
- **🔄 Version Control** - Automatic file versioning with duplicate detection
- **⚡ High Performance** - Async/await architecture with connection pooling
- **🔧 Type Safety** - Full type safety with Pydantic models and validation
- **🐳 Docker Ready** - Complete containerization with Docker Compose

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   MCP Client    │───▶│  Knowledge MCP  │───▶│   PostgreSQL    │
│                 │    │     Server      │    │   + pgvector    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   OpenAI API    │
                       │   (Embeddings)  │
                       └─────────────────┘
```

## 📋 Prerequisites

- **Python 3.13+**
- **PostgreSQL** with pgvector extension
- **OpenAI API Key** (for embeddings)
- **UV** package manager (recommended) or pip

## 🚀 Quick Start

### Option 1: Docker Compose (Recommended) with  MCP-Oauth-Proxy Integration

1. **Clone and setup**:
   ```bash
   git clone <repository>
   cd knowledge-mcp
   ```

2. **Set environment variables**:
   set openAI api key
   ```bash
   export OPENAI_API_KEY="your-openai-api-key-here"
   ```
   export Google Oauth client ID and secret too:
   ```bash
    export OAUTH_CLIENT_ID=xxx
    export OAUTH_CLIENT_SECRET=xxx
    ```

3. **Start services**:
   ```bash
   docker-compose up
   ```

4. **Connect to the proxy server at localhost:8080/mcp.**
 For example, if using MCP Inspector:
- go to : http://localhost:6274
- set URL to : http://localhost:8080/mcp
- click on `Open Auth Settings`
- click on `Quick Oauth Flow` and authenticate with a google account.
- Now click `Connect`
- Success! Maybe navigate to `Tools` now.

### Option 2: Local Development

1. **Install dependencies**:
   ```bash
   uv sync
   # or with pip: pip install -e .
   ```

2. **Start PostgreSQL with pgvector**:
   ```bash
   docker run -d \
     --name postgres-pgvector \
     -e POSTGRES_DB=postgres \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=password \
     -p 5432:5432 \
     pgvector/pgvector:0.8.0-pg15
   ```

3. **Set environment variables**:
   ```bash
   export OPENAI_API_KEY="your-openai-api-key-here"
   export DATABASE_URL="postgresql+asyncpg://postgres:password@localhost:5432/postgres"
   ```

4. **Run the server**:
   ```bash
   uv run python -m app.main
   # or: python -m app.main
   ```

## ⚙️ Configuration

All configuration is managed through environment variables with centralized validation:

### Required Configuration

| Variable | Description | Example |
|----------|-------------|---------|
| `OPENAI_API_KEY` | OpenAI API key for embeddings | `sk-...` |

### Optional Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `9000` | Server port |
| `MCP_PATH` | `/mcp/knowledge` | MCP endpoint path |
| `DATABASE_URL` | `postgresql+asyncpg://postgres:password@db:5432/postgres` | PostgreSQL connection URL |
| `EMBEDDING_MODEL` | `text-embedding-3-small` | OpenAI embedding model |
| `EMBEDDING_DIMENSION` | `1536` | Embedding vector dimension |
| `SKIP_OPENAI_VALIDATION` | `false` | Skip API key validation (testing) |

### Environment Detection

The server automatically detects the environment:
- **Docker**: Uses `db:5432` if `IS_DOCKER_ENVIRONMENT` is set
- **Local**: Uses `localhost:5432` otherwise

## 🔧 API Reference

### Knowledge Set Management

#### Create Knowledge Set
```python
create_knowledge_set() -> KnowledgeSetInfo
```
Creates a new isolated knowledge collection.

#### List Knowledge Sets
```python
list_knowledge_sets() -> List[KnowledgeSetInfo]
```
Lists all knowledge sets for the authenticated user.

#### Delete Knowledge Set
```python
delete_knowledge_set(knowledge_set_id: str) -> dict
```
Deletes a knowledge set and all associated data.

### File Operations

#### Ingest File
```python
ingest_file(
    knowledge_set_id: str,
    filename: str,
    content: str  # base64-encoded
) -> FileUploadResponse
```
Uploads and processes a file through the complete pipeline:
1. **Decode** base64 content
2. **Extract** text using MarkItDown
3. **Chunk** text intelligently
4. **Generate** embeddings via OpenAI
5. **Store** in vector database
6. **Handle** duplicates and versioning

**Supported formats**: PDF, DOCX, TXT, MD, HTML, and more via MarkItDown.

#### List Files
```python
list_files(knowledge_set_id: str) -> List[FileInfo]
```
Lists all files in a knowledge set with metadata.

#### Remove File
```python
remove_file(knowledge_set_id: str, file_id: str) -> dict
```
Deletes a file and all its associated chunks.

### Search & Query

#### Query Knowledge
```python
query(
    knowledge_set_id: str,
    query_text: str
) -> List[QueryResult]
```
Performs semantic search across the knowledge base:
1. **Generate** query embedding
2. **Search** using cosine similarity
3. **Return** top 5 most relevant chunks with scores

## 💡 Usage Examples

### Using a Client App (Cursor, VSCode, etc.)

1. **Start the services** using Docker Compose (includes OAuth proxy):
   ```bash
   docker-compose up
   ```

2. **Configure your MCP client** with these settings:
   - **Server URL**: `http://localhost:8080/mcp`
   - **Authentication**: OAuth flow will be handled automatically by the proxy

3. **Connect and authenticate**:
   - Your client app will redirect you to authenticate with Google OAuth
   - Once authenticated, you'll have access to all Knowledge MCP tools

**Note**: The OAuth proxy service exposes the MCP server at port 8080, handling authentication seamlessly so your client app can focus on using the knowledge management features.

### Using Python FastMCP Client

```python
from fastmcp import Client
from fastmcp.client.transports import StreamableHttpTransport

async with Client(transport=StreamableHttpTransport(
    "http://localhost:9000/mcp/knowledge",
    headers={"x-forwarded-user": "user123"}
)) as client:
    # Create knowledge set
    result = await client.call_tool("create_knowledge_set", {})
    
    # Ingest file
    result = await client.call_tool("ingest_file", {
        "knowledge_set_id": "...",
        "filename": "document.pdf",
        "content": "base64-encoded-content"
    })
    
    # Query
    result = await client.call_tool("query", {
        "knowledge_set_id": "...",
        "query_text": "search query"
    })
```

## 🧩 Text Processing Features

### Intelligent Chunking

The system supports multiple chunking strategies via the Chonkie library:

- **Sentence Chunking**: Respects sentence boundaries
- **Semantic Chunking**: Groups semantically similar content
- **Recursive Chunking**: Hierarchical separator-based chunking
- **Token Chunking**: Token-aware chunking with proper tokenization
- **Basic Chunking**: Fallback with improved boundary detection

### File Format Support

Powered by MarkItDown for robust file processing:
- **Documents**: PDF, DOCX, PPTX, XLSX
- **Text**: TXT, MD, HTML, XML
- **Images**: OCR support for text extraction
- **Audio**: Transcription capabilities
- **And more**: Extensible format support

## 🔍 Vector Database

### PostgreSQL + pgvector

- **Efficient Storage**: Optimized vector storage and indexing
- **Cosine Similarity**: Fast similarity search with IVFFlat indexing
- **Scalability**: Handles large document collections
- **ACID Compliance**: Reliable data consistency

### Schema Design

```sql
-- Knowledge Sets (user isolation)
CREATE TABLE knowledge_sets (
    user_id VARCHAR PRIMARY KEY,
    knowledge_set_id VARCHAR PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW()
);

-- File Metadata
CREATE TABLE files (
    user_id VARCHAR,
    knowledge_set_id VARCHAR,
    file_id VARCHAR,
    file_metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, knowledge_set_id, file_id)
);

-- Vector Chunks
CREATE TABLE chunks (
    user_id VARCHAR,
    knowledge_set_id VARCHAR,
    file_id VARCHAR,
    chunk_id VARCHAR,
    embedding VECTOR(1536),
    chunk_metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, knowledge_set_id, file_id, chunk_id)
);
```

## 🛠️ Development

### Project Structure

```
knowledge-mcp/
├── app/
│   ├── __init__.py
│   ├── main.py              # FastMCP server and endpoints
│   ├── config.py            # Centralized configuration
│   ├── embeddings.py        # OpenAI embedding integration
│   ├── vector_db.py         # PostgreSQL/pgvector operations
│   ├── db_schema.py         # Pydantic models
│   └── text_processing.py   # File processing and chunking
├── docker-compose.yml       # Development environment
├── Dockerfile              # Production container
├── pyproject.toml          # Dependencies and project config
└── README.md               # This file
```

### Key Dependencies

- **FastMCP**: Model Context Protocol server framework
- **SQLAlchemy**: Database ORM with async support
- **pgvector**: PostgreSQL vector extension
- **Pydantic**: Data validation and settings management
- **MarkItDown**: Universal document text extraction
- **Chonkie**: Intelligent text chunking
- **HTTPX**: Async HTTP client for OpenAI API

## 🔒 Security Considerations

- **User Isolation**: All data is isolated by user ID
- **Input Validation**: Comprehensive validation via Pydantic
- **API Key Security**: Secure handling of OpenAI credentials
- **SQL Injection**: Protected via SQLAlchemy ORM
- **Rate Limiting**: OpenAI API rate limit handling with exponential backoff

## 📊 Performance

- **Async Architecture**: Non-blocking I/O for high concurrency
- **Connection Pooling**: Efficient database connections
- **Vector Indexing**: Fast similarity search with IVFFlat
- **Batch Processing**: Efficient embedding generation
- **Caching**: Duplicate detection and content hashing
---

**Built with ❤️ using FastMCP, PostgreSQL, and OpenAI**
