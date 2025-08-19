# Docker Deployment Guide

## Prerequisites

- Docker and Docker Compose installed
- Slack Bot Token (get from https://api.slack.com/apps)

## Quick Start

1. **Set up environment variables:**

   ```bash
   # Create .env file
   echo "SLACK_BOT_TOKEN=xoxb-your-bot-token-here" > .env
   echo "PORT=3000" >> .env
   ```

2. **Build and run with Docker Compose:**

   ```bash
   docker-compose up -d
   ```

3. **Check if it's running:**
   ```bash
   docker-compose ps
   curl http://localhost:3000/health
   ```

## Manual Docker Build

1. **Build the image:**

   ```bash
   docker build -t slack-mcp .
   ```

2. **Run the container:**
   ```bash
   docker run -d \
     --name slack-mcp \
     -p 3000:3000 \
     -e SLACK_BOT_TOKEN=xoxb-your-bot-token-here \
     slack-mcp
   ```

## Environment Variables

- `SLACK_BOT_TOKEN` (required): Your Slack bot token
- `PORT` (optional): Server port, defaults to 3000

## Health Check

The container includes a health check that verifies the server is responding:

```bash
docker inspect slack-mcp | grep Health -A 10
```

## Logs

View container logs:

```bash
docker-compose logs -f slack-mcp
```

## Stopping

```bash
docker-compose down
```

## Production Deployment

For production, consider:

- Using Docker secrets for sensitive data
- Setting up proper logging
- Using a reverse proxy (nginx)
- Implementing proper monitoring
