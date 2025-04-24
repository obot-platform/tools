# syntax=docker/dockerfile:1
FROM cgr.dev/chainguard/wolfi-base AS base

RUN apk add --no-cache go make git nodejs npm pnpm curl python-3.13 py3.13-pip

FROM base AS tools
WORKDIR /obot-tools/tools
COPY . /obot-tools/tools
RUN --mount=type=cache,id=pnpm,target=/root/.local/share/pnpm/store \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/.cache/uv \
    --mount=type=cache,target=/root/go/pkg/mod \
    UV_LINK_MODE=copy BIN_DIR=/bin make package-tools

FROM base AS providers
WORKDIR /obot-tools
COPY ./Makefile /obot-tools/
COPY ./scripts/package-providers.sh /obot-tools/scripts/

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    BIN_DIR=/bin make package-providers