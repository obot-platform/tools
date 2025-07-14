FROM ghcr.io/astral-sh/uv:python3.13-bookworm-slim

# Prevents Python from writing .pyc files and buffers stdout/stderr
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

WORKDIR /app

# Install system deps (if any needed) - using apt for Debian
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Copy uv project files first for better Docker layer caching
COPY pyproject.toml uv.lock README.md ./

# Install dependencies using uv sync (faster and more reliable)
RUN uv sync --frozen --no-dev

# Copy your application code
COPY . .

ENV PORT=9000
EXPOSE 9000

# Start the server using uv run (ensures proper virtual environment)
CMD ["uv", "run", "python", "-m", "server"]