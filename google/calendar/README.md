# Obot Google Calendar MCP Server

- supports streamable HTTP and stdio

## Installation & Running

### Docker-compose (Recommended)
Export (Google's) Oauth CLient ID and Secret for Oauth Proxy
```bash
export OAUTH_CLIENT_ID=xxx
export OAUTH_CLIENT_SECRET=xxx
```

then:
```bash
docker-compose up
```

### Using uvx 
install from local directory:
```bash
uvx --from . google-calendar-mcp
```

### Using uv (Development)
Install dependencies:
```bash
uv pip install
```

Run the server:
```bash
uv run server.py
```

## Testing

### Unit-test with pytest
```bash
uv run python -m pytest
```

### Integration Testing

#### Get Your Access Token
This MCP server assumes Obot will take care of the Oauth2.0 flow and supply an access token. To test locally or without Obot, you need to get an access token by yourself. I use [postman workspace](https://blog.postman.com/how-to-access-google-apis-using-oauth-in-postman/) to create and manage my tokens.

#### Local Example Client
```
export GOOGLE_OAUTH_TOKEN=xxx
```
and then
```
uv run example_client.py
```