FROM golang:1.25.0-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o word-mcp-go .

FROM alpine:latest

# Install runtime dependencies for document conversion
# - pandoc: Primary markdown to DOCX converter
# - libreoffice: Fallback converter (soffice command)
# - ca-certificates: For HTTPS connections
# - ttf-dejavu: Basic font support for document rendering
RUN apk --no-cache add \
    ca-certificates \
    pandoc \
    libreoffice \
    ttf-dejavu \
    && rm -rf /var/cache/apk/*

# Create a non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -S -u 1001 -G appgroup appuser

# Create necessary directories
RUN mkdir -p /tmp && chmod 1777 /tmp

WORKDIR /root/

COPY --from=builder /app/word-mcp-go .

# Make sure the binary is executable
RUN chmod +x word-mcp-go

# Switch to non-root user (but keep as root for now due to temp file access needs)
# USER appuser

EXPOSE 9000

# Set environment variables for LibreOffice
ENV HOME=/tmp
ENV TMPDIR=/tmp

CMD ["./word-mcp-go"]