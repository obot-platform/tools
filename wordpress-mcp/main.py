"""WordPress MCP Server main entry point."""

import asyncio
from src.server import mcp


def main():
    """Main entry point for the WordPress MCP Server."""
    try:
        asyncio.run(mcp.run())
    except KeyboardInterrupt:
        print("\nWordPress MCP Server shutting down...")
    except Exception as e:
        print(f"Error starting WordPress MCP Server: {e}")


if __name__ == "__main__":
    main()
