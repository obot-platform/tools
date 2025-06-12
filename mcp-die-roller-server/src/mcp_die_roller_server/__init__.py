import asyncio
import logging
import sys
from pathlib import Path

import click

from .main import roll_die, serve


@click.command()
@click.option("-v", "--verbose", count=True)
def main(verbose: bool) -> None:
    """MCP Die roller Server - Dice rolling functionality for MCP"""

    logging_level = logging.WARN
    if verbose == 1:
        logging_level = logging.INFO
    elif verbose >= 2:
        logging_level = logging.DEBUG

    logging.basicConfig(level=logging_level, stream=sys.stderr)
    asyncio.run(serve())


if __name__ == "__main__":
    main()

__all__ = ["roll_die"]
