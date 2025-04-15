# syntax=docker/dockerfile:1
FROM cgr.dev/chainguard/wolfi-base AS base

RUN apk add --no-cache go make git nodejs npm pnpm

FROM base AS tools
ARG TOOL_REGISTRY_REPOS='github.com/obot-platform/tools'
RUN apk add --no-cache curl python-3.13 py3.13-pip
WORKDIR /app
COPY . .
RUN --mount=type=cache,id=pnpm,target=/root/.local/share/pnpm/store \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/.cache/uv \
    --mount=type=cache,target=/root/go/pkg/mod \
    --mount=type=secret,id=GITHUB_TOKEN,env=GITHUB_TOKEN \
    UV_LINK_MODE=copy BIN_DIR=/bin TOOL_REGISTRY_REPOS=$TOOL_REGISTRY_REPOS make package-tools