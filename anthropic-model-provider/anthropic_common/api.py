import json
import logging
import time
from typing import Any, AsyncIterable

from anthropic import AsyncAnthropic, AsyncAnthropicBedrock, AsyncStream
from anthropic._types import NOT_GIVEN
from anthropic.types import (
    RawContentBlockDeltaEvent,
    RawContentBlockStartEvent,
    RawContentBlockStopEvent,
    RawMessageDeltaEvent,
    RawMessageStartEvent,
    RawMessageStopEvent,
    RawMessageStreamEvent, TextDelta, ThinkingDelta,
)
from fastapi.responses import JSONResponse, StreamingResponse
from openai.types.chat import ChatCompletionChunk
from openai.types.chat.chat_completion_chunk import Choice, ChoiceDelta

from .helpers import log, map_messages, map_resp, map_tools


async def completions(client: AsyncAnthropic | AsyncAnthropicBedrock, input: dict):
    is_thinking = False

    model = str(input["model"])
    if model.endswith("-thinking"):
        is_thinking = True
        model = model.removesuffix("-thinking")

    # max_tokens defaults:
    # - 4096 for regular models, so that it works with even the smallest models
    # - 64000 for thinking models - the max for 3.7 Sonnet with extended thinking mode right now
    max_tokens = input.get("max_tokens", 4096 if not is_thinking else 64000)
    if max_tokens is not None:
        max_tokens = int(max_tokens)

    thinking_config: Any | NOT_GIVEN = NOT_GIVEN
    if is_thinking:
        thinking_config = {
            "type": "enabled",
            "budget_tokens": round(
                max_tokens / 2
            ),  # TODO: figure out a good percentage of max_tokens to use for thinking
        }

    tools = input.get("tools", NOT_GIVEN)
    if tools is not NOT_GIVEN:
        tools = map_tools(tools)

    system, messages = map_messages(input["messages"])

    temperature = input.get("temperature", NOT_GIVEN) if not is_thinking else NOT_GIVEN
    if temperature is not NOT_GIVEN:
        temperature = float(temperature)

    top_k = input.get("top_k", NOT_GIVEN) if not is_thinking else NOT_GIVEN
    if top_k is not NOT_GIVEN:
        top_k = int(top_k)

    top_p = input.get("top_p", NOT_GIVEN) if not is_thinking else NOT_GIVEN
    if top_p is not NOT_GIVEN:
        top_p = float(top_p)

    stream = input.get("stream", False)

    logging.error(f"@@@ thinking_config: {thinking_config}")
    try:
        response = await client.messages.create(
            thinking=thinking_config,
            max_tokens=max_tokens,
            system=system,
            messages=messages,
            model=model,
            temperature=temperature,
            tools=tools,
            top_k=top_k,
            top_p=top_p,
            stream=stream,
        )
        if stream:
            async for event in response:
                logging.error(f"@@@Anthropic event: {event.model_dump_json()}")
                log(f"Anthropic event: {event.model_dump_json()}")
            return StreamingResponse(
                "data: " + event.model_dump_json() + "\n\n",
                media_type="application/x-ndjson",
            )
        else:
            logging.error(f"@@@ Anthropic response: {response.model_dump_json()}")
            log(f"Anthropic response: {response.model_dump_json()}")

            mapped_response = map_resp(response)

            logging.error(f"@@@ Mapped Anthropic response: {mapped_response.model_dump_json()}")
            log(f"Mapped Anthropic response: {mapped_response.model_dump_json()}")
            return StreamingResponse(
                "data: " + mapped_response.model_dump_json() + "\n\n",
                media_type="application/x-ndjson",
            )

    except Exception as e:
        logging.error(f"@@@Anthropic API error: {e}")
        return JSONResponse(
            content={"error": str(e)}, status_code=e.__dict__.get("status_code", 500)
        )

def map_event(event: RawMessageStreamEvent, model: str) -> ChatCompletionChunk:
    if isinstance(event, RawContentBlockStartEvent):
        if event.type == "tool_use":
            c = ChatCompletionChunk(
                id="0",
                choices=[
                    Choice(
                        delta=ChoiceDelta(
                            content=None,
                            tool_calls=map_tools(event.content_block.tool_calls),
                            role="assistant",
                        ),
                        finish_reason=None,
                        index=0,
                    )
                ],
                created=int(time.time()),
                model=model,
                object="chat.completion.chunk",
            )
    elif isinstance(event, RawContentBlockDeltaEvent):
        content = ""
        if isinstance(event.delta, TextDelta):
            content = event.delta.text
        elif isinstance(event.delta, ThinkingDelta):
            content = event.delta.thinking
        c = ChatCompletionChunk(
            id="0",
            choices=[
                Choice(
                    delta=ChoiceDelta(
                        content=content,
                        tool_calls=None,
                        role="assistant",
                    ),
                    finish_reason=None,
                    index=0,
                )
            ],
            created=int(time.time()),
            model=model,
            object="chat.completion.chunk",
        )

        if hasattr(event, "tool_calls"):

        return c
    elif isinstance(event, RawContentBlockStopEvent):
        pass
    elif isinstance(event, RawMessageStartEvent):
        pass
    elif isinstance(event, RawMessageDeltaEvent):
        pass
    elif isinstance(event, RawMessageStopEvent):
        pass
    else:
        raise ValueError(f"Unknown event type: {event}")


async def convert_stream(
        stream: AsyncStream[RawMessageStreamEvent], model: str
) -> AsyncIterable[str]:
    async for chunk in stream:
        yield "data: " + map_event(chunk, model).model_dump_json() + "\n\n"