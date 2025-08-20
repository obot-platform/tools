"""WordPress MCP Server implementation."""

from fastmcp import FastMCP
from src.config import config

# Create the MCP server instance
mcp = FastMCP("WordPress")

def register_tools():
    """Register all WordPress tools with the MCP server."""
    # Import all tool modules to register them
    from src.tools import posts, users, media, categories, tags, site

# Register tools when module is imported
register_tools()