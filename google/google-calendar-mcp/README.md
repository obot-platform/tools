# Obot Google Calendar MCP Server
- Obot Google Calendar mcp server, converted from the google-calendar tool bundle.
- supports streamable HTTP
- tools of this mcp server expect `cred_token`(access_token of google oauth) as part of the tool input.

## uv install
```
uv pip install
```

## Run the Server
```
cd src
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