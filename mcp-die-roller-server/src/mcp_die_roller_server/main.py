from typing import Annotated

from dice_roller import Dice
from fastmcp import FastMCP

mcp = FastMCP(name="die-roller")


@mcp.tool()
async def roll_die(
    num_dice: Annotated[int, "The number of dice to roll"],
    sides: Annotated[int, "The number of sides on the dice"],
) -> int:
    dice_roller = Dice(sides)
    total = num_dice @ dice_roller
    return {"total": total.roll()}


def serve():
    mcp.run()


if __name__ == "__main__":
    mcp.run()
