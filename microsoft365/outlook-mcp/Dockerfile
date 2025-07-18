FROM ghcr.io/astral-sh/uv:python3.13-bookworm-slim

# Prevents Python from writing .pyc files and buffers stdout/stderr
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

WORKDIR /app

# Install system deps (if any needed) - using apt for Debian
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

COPY pyproject.toml uv.lock ./

# Install dependencies only (without building the local package)
RUN uv sync --frozen --no-dev --no-install-project

COPY app/ ./app/

RUN uv sync --frozen --no-dev --no-editable

EXPOSE 9000

# Start the server using uv run (ensures proper virtual environment)
CMD ["uv", "run", "python", "-m", "app.main"]