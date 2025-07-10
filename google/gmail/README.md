# Obot Gmail MCP Server
- Obot Gmail mcp server, converted from the google-gmail tool bundle.
- supports streamable HTTP
- tools of this mcp server expect `cred_token`(access_token of google oauth) as part of the tool input.

## Installation

### Docker-compose
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
```bash
# Run directly from the current directory
uvx --from . obot-gmail-mcp
```
or stdio server:
```
uvx --from . obot-gmail-mcp-stdio
```

### Using uv (development)
```bash
uv pip install -e .
```

## Run the Server

### Using uvx
```bash
uvx --from . obot-gmail-mcp
```

### Using uv
```bash
uv run server.py
```

## Testing

### Unit-test with pytest
```
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