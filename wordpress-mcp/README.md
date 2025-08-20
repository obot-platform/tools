# WordPress MCP Server

A Model Context Protocol (MCP) server for managing WordPress sites via the REST API. This server provides 18 tools for comprehensive WordPress management including posts, users, media, categories, tags, and site settings.

## Features

### Posts Management (6 tools)
- **list_posts**: List posts with filtering options (status, date, author, categories, tags)
- **retrieve_post**: Get detailed information about a specific post
- **create_post**: Create new posts with full metadata support
- **update_post**: Update existing posts with selective field updates
- **delete_post**: Delete posts (with trash/permanent delete options)

### Users Management (3 tools)
- **list_users**: List site users (admin permission required)
- **get_me**: Get current user information and capabilities
- **validate_credential**: Test WordPress authentication

### Media Management (3 tools)
- **list_media**: List media files with filtering options
- **update_media**: Update media metadata (title, slug, author)
- **delete_media**: Delete media files

### Categories Management (4 tools)
- **list_categories**: List categories with search and filtering
- **create_category**: Create new categories with hierarchy support
- **update_category**: Update existing category information
- **delete_category**: Remove categories

### Tags Management (4 tools)
- **list_tags**: List tags with search and filtering
- **create_tag**: Create new tags
- **update_tag**: Update existing tag information
- **delete_tag**: Remove tags

### Site Management (1 tool)
- **get_site_settings**: Get WordPress site settings (admin required)

## Prerequisites

### WordPress Site Requirements
1. **WordPress.com sites are NOT supported** - only self-hosted WordPress sites
2. **Permalinks must be configured** - go to Settings > Permalinks and choose any structure other than "Plain"
3. **Application Password required** - regular WordPress passwords won't work

### Setting up Application Password
1. Go to your WordPress admin dashboard
2. Navigate to Users > Your Profile
3. Scroll to "Application Passwords" section
4. Click "Add New Application Password"
5. Give it a name (e.g., "MCP Server") and click "Add New Application Password"
6. Copy the generated password - you'll need this for the `WORDPRESS_PASSWORD` environment variable

## Installation

1. Clone or download this project
2. Install dependencies using uv:
```bash
cd wordpress-mcp
uv sync
```

## Configuration

Create a `.env` file in the project root with your WordPress credentials:

```env
WORDPRESS_SITE=https://your-wordpress-site.com
WORDPRESS_USERNAME=your-username
WORDPRESS_PASSWORD=your-application-password
```

**Important Notes:**
- `WORDPRESS_SITE`: The full URL to your WordPress site (with https:// or http://)
- `WORDPRESS_USERNAME`: Your WordPress username (not email)
- `WORDPRESS_PASSWORD`: The application password you created (NOT your regular WordPress password)

## Usage

### Running the MCP Server

```bash
# Using uv
uv run python main.py

# Or if you have the environment activated
python main.py
```

### Using with MCP Clients

The server exposes all tools via the Model Context Protocol. You can connect any MCP-compatible client to use these tools for WordPress management.

## Tool Examples

### Creating a Post
```python
# This would be called by an MCP client
create_post(
    title="My New Blog Post",
    content="<p>This is the content with <strong>HTML formatting</strong>.</p>",
    status="draft",
    categories="1,5",
    tags="10,15"
)
```

### Listing Posts
```python
list_posts(
    status="publish,draft",
    per_page=20,
    order="desc",
    categories="1,2,3"
)
```

### Managing Categories
```python
# Create a category
create_category(
    category_name="Technology",
    description="Posts about technology topics",
    parent_id=1
)

# Update a category
update_category(
    category_id=5,
    name="New Technology",
    description="Updated description"
)
```

## Error Handling

The server provides detailed error messages for common issues:

- **401 Authentication Failed**: Check your credentials and application password
- **403 Permission Denied**: Your user may not have sufficient permissions
- **400 Bad Request**: Invalid parameters or data format
- **404 Not Found**: Requested resource doesn't exist

## Security Notes

- This server uses WordPress Application Passwords for authentication
- All API calls use HTTPS when possible
- Credentials are loaded from environment variables
- No sensitive data is logged

## Troubleshooting

### Common Issues

1. **"Authentication failed"**
   - Verify your application password is correct
   - Ensure you're using the username, not email
   - Check that the WordPress site URL is correct

2. **"Permission denied"** 
   - Some operations require admin privileges
   - Check your user role in WordPress

3. **"Invalid site URL"**
   - URL must start with `https://` or `http://`
   - Don't include `/wp-json` in the URL

4. **Tools not working**
   - Ensure permalinks are configured (not "Plain")
   - Check that the WordPress REST API is enabled

### Testing Your Setup

Use the `validate_credential` tool to test your configuration:

```python
validate_credential()
```

This will return `{"valid": true}` if everything is configured correctly.

## Development

The project structure:

```
wordpress-mcp/
├── src/
│   ├── config.py          # Configuration and WordPress client
│   ├── server.py          # FastMCP server setup
│   └── tools/             # Individual tool implementations
│       ├── posts.py       # Posts management
│       ├── users.py       # Users management
│       ├── media.py       # Media management
│       ├── categories.py  # Categories management
│       ├── tags.py        # Tags management
│       └── site.py        # Site settings
├── main.py                # Server entry point
└── pyproject.toml         # Dependencies and configuration
```

## License

This project is open source and available under the MIT License.