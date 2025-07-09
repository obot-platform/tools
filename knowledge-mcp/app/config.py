"""
Centralized configuration management for knowledge-mcp.
All environment variables and their defaults are defined here.
"""

import os
from pydantic import Field, field_validator
from pydantic_settings import BaseSettings


class Config(BaseSettings):
    """Application configuration loaded from environment variables."""

    # Server Configuration
    port: int = Field(default=9000, description="Server port", alias="PORT")
    mcp_path: str = Field(
        default="/mcp/knowledge", description="MCP endpoint path", alias="MCP_PATH"
    )

    # Database Configuration
    database_url: str = Field(
        default_factory=lambda: (
            "postgresql+asyncpg://postgres:password@db:5432/postgres"
            if "IS_DOCKER_ENVIRONMENT" in os.environ
            else "postgresql+asyncpg://postgres:password@localhost:5432/postgres"
        ),
        description="PostgreSQL database connection URL",
        alias="DATABASE_URL",
    )

    # OpenAI Configuration
    openai_api_key: str = Field(
        ..., description="OpenAI API key (required)", alias="OPENAI_API_KEY"
    )
    embedding_model: str = Field(
        default="text-embedding-3-small",
        description="OpenAI embedding model to use",
        alias="EMBEDDING_MODEL",
    )
    embedding_dimension: int = Field(
        default=1536,
        description="Embedding vector dimension",
        alias="EMBEDDING_DIMENSION",
    )

    # Optional Configuration
    skip_openai_validation: bool = Field(
        default=False,
        description="Skip OpenAI API key validation (for testing)",
        alias="SKIP_OPENAI_VALIDATION",
    )

    @field_validator("openai_api_key")
    @classmethod
    def validate_openai_api_key(cls, v):
        """Validate OpenAI API key is not empty."""
        if not v or not v.strip():
            raise ValueError("OPENAI_API_KEY cannot be empty")
        return v.strip()

    @field_validator("port")
    @classmethod
    def validate_port(cls, v):
        """Validate port is in valid range."""
        if not 1 <= v <= 65535:
            raise ValueError("PORT must be between 1 and 65535")
        return v

    @field_validator("embedding_dimension")
    @classmethod
    def validate_embedding_dimension(cls, v):
        """Validate embedding dimension is positive."""
        if v <= 0:
            raise ValueError("EMBEDDING_DIMENSION must be positive")
        return v

    model_config = {
        # Map environment variable names to field names
        "env_prefix": "",  # No prefix, use exact names
        "env_file": ".env",
        "env_file_encoding": "utf-8",
    }


# Global configuration instance
# This will be initialized when the module is imported
try:
    config = Config()
    print("✓ Configuration loaded successfully")
except Exception as e:
    raise ValueError(f"Configuration validation failed: {e}")


# Convenience exports for backward compatibility
PORT = config.port
MCP_PATH = config.mcp_path
DATABASE_URL = config.database_url
OPENAI_API_KEY = config.openai_api_key
EMBEDDING_MODEL = config.embedding_model
EMBEDDING_DIMENSION = config.embedding_dimension
SKIP_OPENAI_VALIDATION = config.skip_openai_validation
