# Authentication Providers

This file describes the requirements that a tool must meet in order to function properly as an Obot auth provider.

## Requirements

### Daemon Tool

All auth providers must be daemon tools. They should only have one tool defined in `tool.gpt`.
This tool should start the HTTP server, and follow the same requirements as all other daemon tools.

### Metadata

The tool must have a metadata line in the `tool.gpt` called `envVars`, that lists all the required configuration
parameters for the auth provider.

Optionally, the tool may also include an `optionalEnvVars` metadata line, that lists all the optional configuration parameters.

Example from the GitHub auth provider:

```
...
Metadata: envVars: OBOT_GITHUB_AUTH_PROVIDER_CLIENT_ID,OBOT_GITHUB_AUTH_PROVIDER_CLIENT_SECRET,OBOT_AUTH_PROVIDER_COOKIE_SECRET,OBOT_AUTH_PROVIDER_EMAIL_DOMAINS
Metadata: optionalEnvVars: OBOT_GITHUB_AUTH_PROVIDER_TEAMS,OBOT_GITHUB_AUTH_PROVIDER_ORG,OBOT_GITHUB_AUTH_PROVIDER_REPO,OBOT_GITHUB_AUTH_PROVIDER_TOKEN,OBOT_GITHUB_AUTH_PROVIDER_ALLOW_USERS
...
```

### Placeholder Credential

The auth provider must reference the placeholder credential in this repo, like this:

```
Credential: ../placeholder-credential as <credential name>
```

### Implementation Details

The auth provider must implement user authentication using OAuth2. The user should be able to log in using a standard
OAuth2 authorization code flow.

#### Token Cookie

The auth provider must store the token in a cookie called `obot_access_token`.
This cookie should be set as `Secure` only if the `OBOT_SERVER_URL` environment variable starts with `https://`.
The cookie should be encrypted using the `OBOT_AUTH_PROVIDER_COOKIE_SECRET` environment variable.

#### URL Paths

The auth provider must implement the following URL paths:

- `/oauth2/start`: This path should start the OAuth2 flow by sending the user to the OAuth2 provider's authorization URL.
  - It must check for the `rd` query parameter. This is the URL to redirect the user to after the full OAuth2 flow is complete.
    This value can be stored alongside the state if needed.
- `/oauth2/callback`: This path should handle the OAuth2 callback from the OAuth2 provider.
  - After exchanging the code for the access token, it should redirect the user to the URL stored in the `rd` query parameter
    from the `/oauth2/start` request.
- `/oauth2/sign_out`: This path should sign the user out by clearing the `obot_access_token` cookie and redirecting the user to
  the URL in the `rd` query parameter.
- `/obot-get-icon-url`: This path should take the user's access token from the `Authorization` header and use it to get
  the user's profile picture URL. It should return a JSON object with the URL in the `iconURL` field.
  The `Authorization` header will be in the format `Bearer <access token>`.
- `/obot-get-state`: More details in the next section.

##### Obot-Get-State

`/obot-get-state` is the path that Obot uses to get information about the user making a request.
Requests to this path will include a JSON body with the following JSONSchema:

```json
{
  "type": "object",
  "properties": {
    "method": {
      "type": "string",
      "description": "The HTTP method of the request (e.g., GET, POST)."
    },
    "url": {
      "type": "string",
      "format": "uri",
      "description": "The URL of the request."
    },
    "header": {
      "type": "object",
      "additionalProperties": {
        "type": "array",
        "items": {
          "type": "string"
        }
      },
      "description": "Headers of the request, where keys are header names and values are arrays of header values."
    }
  },
  "required": ["method", "url", "header"],
  "additionalProperties": false
}
```

This object represents a serialized HTTP request that Obot received from a user.
The auth provider must return information about the authenticated user that made this request.
Under most (if not all) circumstances, the auth provider only needs to look at the cookie header,
which it can then decrypt and use to get information about the user.

The auth provider must return a JSON object with information about the user,
matching the following JSONSchema:

```json
{
  "type": "object",
  "properties": {
    "accessToken": {
      "type": "string",
      "description": "The access token for the user."
    },
    "preferredUsername": {
      "type": "string",
      "description": "The preferred username of the user."
    },
    "user": {
      "type": "string",
      "description": "The identifier for the user."
    },
    "email": {
      "type": "string",
      "format": "email",
      "description": "The email address of the user."
    }
  },
  "required": ["accessToken", "preferredUsername", "user", "email"],
  "additionalProperties": false
}
```

Here is an example:

```json
{
  "accessToken": "xyz",
  "preferredUsername": "johndoe",
  "user": "johndoe",
  "email": "johndoe@example.com"
}
```

If the `obot_access_token` cookie is not present or is invalid, the auth provider should return a 400 status code.

## Reference Implementation

The Google auth provider in this repo should be considered the standard reference implementation.
It follows all the requirements listed above and can be used as a reference when implementing a new auth provider.
It is recommended to make use of the [OAuth2 Proxy](https://github.com/obot-platform/oauth2-proxy) like the Google and GitHub auth providers do.
