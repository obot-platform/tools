FROM ghcr.io/astral-sh/uv:python3.13-bookworm-slim

# Prevents Python from writing .pyc files and buffers stdout/stderr
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

WORKDIR /app

# Install system deps (if any needed) - using apt for Debian
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    git \
    && rm -rf /var/lib/apt/lists/*

# Copy uv project files first for better Docker layer caching
COPY pyproject.toml uv.lock ./

# Copy your application code (needed for local package build)
COPY . .

# Install dependencies using uv sync (faster and more reliable)
RUN uv sync --frozen --no-dev --no-editable

EXPOSE 9000

# Start the server using uv run (ensures proper virtual environment)
CMD ["uv", "run", "python", "-m", "app.server"]