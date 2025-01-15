from anthropic import AsyncAnthropic, AsyncAnthropicBedrock
from anthropic._types import NOT_GIVEN
from fastapi.responses import JSONResponse, StreamingResponse

from .helpers import log, map_messages, map_resp, map_tools


async def completions(client: AsyncAnthropic | AsyncAnthropicBedrock, input: dict):
    tools = input.get("tools", NOT_GIVEN)
    if tools is not NOT_GIVEN:
        tools = map_tools(tools)

    system, messages = map_messages(input["messages"])

    max_tokens = input.get("max_tokens", 1024)
    if max_tokens is not None:
        max_tokens = int(max_tokens)

    temperature = input.get("temperature", NOT_GIVEN)
    if temperature is not NOT_GIVEN:
        temperature = float(temperature)

    top_k = input.get("top_k", NOT_GIVEN)
    if top_k is not NOT_GIVEN:
        top_k = int(top_k)

    top_p = input.get("top_p", NOT_GIVEN)
    if top_p is not NOT_GIVEN:
        top_p = float(top_p)

    try:
        response = await client.messages.create(
            max_tokens=max_tokens,
            system=system,
            messages=messages,
            model=input["model"],
            temperature=temperature,
            tools=tools,
            top_k=top_k,
            top_p=top_p,
        )
    except Exception as e:
        return JSONResponse(
            content={"error": str(e)}, status_code=e.__dict__.get("status_code", 500)
        )

    log(f"Anthropic response: {response.model_dump_json()}")

    mapped_response = map_resp(response)

    log(f"Mapped Anthropic response: {mapped_response.model_dump_json()}")

    return StreamingResponse(
        "data: " + mapped_response.model_dump_json() + "\n\n",
        media_type="application/x-ndjson",
    )
