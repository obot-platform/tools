import json

from fastmcp import Client

from mcp_die_roller_server.main import mcp


async def test_roll_single_die():
    async with Client(mcp) as client:
        result = await client.call_tool("roll_die", {"num_dice": 1, "sides": 6})
        parsed_result = json.loads(result[0].text)
        assert 1 <= int(parsed_result["total"]) <= 6


async def test_roll_multiple_dice():
    async with Client(mcp) as client:
        num_dice = 3
        result = await client.call_tool("roll_die", {"num_dice": 3, "sides": 6})
        parsed_result = json.loads(result[0].text)
        total = int(parsed_result["total"])
        assert num_dice <= total <= (num_dice * 6)


async def test_different_dice_sizes():
    async with Client(mcp) as client:
        # Test d4
        result = await client.call_tool("roll_die", {"num_dice": 1, "sides": 4})
        parsed_result = json.loads(result[0].text)
        assert 1 <= int(parsed_result["total"]) <= 4

        # Test d8
        result = await client.call_tool("roll_die", {"num_dice": 1, "sides": 8})
        parsed_result = json.loads(result[0].text)
        assert 1 <= int(parsed_result["total"]) <= 8

        # Test d12
        result = await client.call_tool("roll_die", {"num_dice": 1, "sides": 12})
        parsed_result = json.loads(result[0].text)
        assert 1 <= int(parsed_result["total"]) <= 12

        # Test d20
        result = await client.call_tool("roll_die", {"num_dice": 1, "sides": 20})
        parsed_result = json.loads(result[0].text)
        assert 1 <= int(parsed_result["total"]) <= 20


async def test_edge_cases():
    async with Client(mcp) as client:
        # Test minimum number of dice
        result = await client.call_tool("roll_die", {"num_dice": 1, "sides": 6})
        parsed_result = json.loads(result[0].text)
        assert 1 <= int(parsed_result["total"]) <= 6

        # Test large number of dice
        num_dice = 10
        result = await client.call_tool("roll_die", {"num_dice": num_dice, "sides": 6})
        parsed_result = json.loads(result[0].text)
        total = int(parsed_result["total"])
        assert num_dice <= total <= (num_dice * 6)


async def test_invalid_inputs():
    async with Client(mcp) as client:
        # Test zero dice
        try:
            await client.call_tool("roll_die", {"num_dice": 0, "sides": 6})
            assert False, "Should have raised an error for zero dice"
        except Exception:
            pass

        # Test negative dice
        try:
            await client.call_tool("roll_die", {"num_dice": -1, "sides": 6})
            assert False, "Should have raised an error for negative dice"
        except Exception:
            pass

        # Test invalid sides
        try:
            await client.call_tool("roll_die", {"num_dice": 1, "sides": 0})
            assert False, "Should have raised an error for zero sides"
        except Exception:
            pass
