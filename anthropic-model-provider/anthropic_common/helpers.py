import json
import os
from typing import Optional, Literal

from openai.types.chat import ChatCompletionChunk
from openai.types.chat.chat_completion_chunk import Choice, ChoiceDelta

disable_provider_system_prompt = (
    os.getenv("DISABLE_PROVIDER_SYSTEM_PROMPT", "false") == "true"
)

debug = os.getenv("GPTSCRIPT_DEBUG", "false") == "true"


def log(*args):
    if debug:
        print(*args)


def map_tools(tools: list[dict]) -> list[dict]:
    anthropic_tools: list[dict] = []
    for tool in tools:
        anthropic_tool = {
            "name": tool["function"]["name"],
            "description": tool["function"].get("description", ""),
            "input_schema": tool["function"].get(
                "parameters", {"type": "object", "properties": {}}
            ),
        }
        anthropic_tools.append(anthropic_tool)
    return anthropic_tools


def map_messages(messages: dict) -> tuple[str, list[dict]]:
    system: str = ""
    if disable_provider_system_prompt:
        system = """
You are task oriented system.
You receive input from a user, process the input from the given instructions, and then output the result.
Your objective is to provide consistent and correct results.
You do not need to explain the steps taken, only provide the result to the given instructions.
You are referred to as a tool.
You don't move to the next step until you have a result.
"""
    mapped_messages: list[dict] = []

    for message in messages:
        role = message.get("role")
        if "content" in message.keys():
            message["content"] = map_content(message.get("content"))

        if role == "system":
            if disable_provider_system_prompt:
                system += message["content"] + "\n"
            else:
                # @TODO: figure out if there is a better way to do this - tool use claude acts differently than regular claude.
                # If it sees the first user message as not being relevant to the conversation, it will complain.
                mapped_messages.append(
                    {
                        "role": "user",
                        "content": [
                            {"text": message["content"], "type": "text"},
                        ],
                    }
                )

        elif role == "user":
            if isinstance(message["content"], list):
                mapped_messages.append(
                    {
                        "role": "user",
                        "content": message["content"],
                    }
                )
            else:
                mapped_messages.append(
                    {
                        "role": "user",
                        "content": [
                            {"text": message["content"], "type": "text"},
                        ],
                    }
                )

        elif role == "tool":
            mapped_messages.append(
                {
                    "role": "user",
                    "content": [
                        {
                            "type": "tool_result",
                            "tool_use_id": message["tool_call_id"],
                            "content": message["content"],
                        }
                    ],
                }
            )

        elif role == "assistant":
            if "tool_calls" in message.keys():
                tool_calls = []
                for tool_call in message["tool_calls"]:
                    tool_calls.append(
                        {
                            "type": "tool_use",
                            "id": tool_call["id"],
                            "name": tool_call["function"]["name"],
                            "input": json.loads(tool_call["function"]["arguments"]),
                        }
                    )
                mapped_messages.append(
                    {
                        "role": "assistant",
                        "content": tool_calls,
                    }
                )
            elif "content" in message.keys() and message["content"] is not None:
                mapped_messages.append(
                    {
                        "role": "assistant",
                        "content": message["content"],
                    }
                )

    if not disable_provider_system_prompt:

        def prepend_if_unique(lst, new_dict, key, value):
            if not any(d.get(key) == value for d in lst):
                lst.insert(0, new_dict)

        prepend_if_unique(
            mapped_messages, {"role": "user", "content": "."}, "role", "user"
        )
        if mapped_messages[0]["role"] != "user":
            mapped_messages.insert(0, {"role": "user", "content": "."})

    mapped_messages = merge_consecutive_dicts_with_same_value(mapped_messages, "role")

    return system, mapped_messages


def merge_consecutive_dicts_with_same_value(list_of_dicts, key) -> list[dict]:
    merged_list = []
    index = 0
    while index < len(list_of_dicts):
        current_dict = list_of_dicts[index]
        value_to_match = current_dict.get(key)
        compared_index = index + 1
        while (
            compared_index < len(list_of_dicts)
            and list_of_dicts[compared_index].get(key) == value_to_match
        ):
            list_of_dicts[compared_index]["content"] = (
                current_dict["content"] + (list_of_dicts[compared_index]["content"])
            )
            current_dict.update(list_of_dicts[compared_index])
            compared_index += 1
        merged_list.append(current_dict)
        index = compared_index
    return merged_list


def map_resp(response):
    parsed_tool_calls = []
    content: str | None = None

    for item in response.content:
        if response.stop_reason == "tool_use":
            if item.type == "tool_use":
                index = len(parsed_tool_calls)
                parsed_tool_calls.append(
                    {
                        "index": index,
                        "id": item.id,
                        "type": "function",
                        "function": {
                            "name": item.name,
                            "arguments": json.dumps(item.input),
                        },
                    }
                )
                content = None
        else:
            if item.type == "text":
                content = item.text

    role = response.role
    finish_reason = map_finish_reason(response.stop_reason)

    resp = ChatCompletionChunk(
        id="0",
        choices=[
            Choice(
                delta=ChoiceDelta(
                    content=content, tool_calls=parsed_tool_calls, role=role
                ),
                finish_reason=finish_reason,
                index=0,
            )
        ],
        created=0,
        model="",
        object="chat.completion.chunk",
    )
    return resp


def map_finish_reason(
    finish_reason: str,
) -> Literal["stop", "length", "tool_calls", "content_filter", "function_call"]:
    if finish_reason == "end_turn":
        return "stop"
    elif finish_reason == "stop_sequence":
        return "stop"
    elif finish_reason == "max_tokens":
        return "length"
    elif finish_reason == "tool_use":
        return "tool_calls"
    return "stop"


def map_content(content: Optional[list | dict]) -> Optional[list | dict]:
    if content is not None and isinstance(content, list):
        for i, item in enumerate(content):
            if not isinstance(item, dict):
                continue
            if item.get("type") == "image_url":
                image_url = item.get("image_url")
                if image_url is not None and isinstance(image_url, dict):
                    url = image_url.get("url")
                    if (
                        url is not None
                        and isinstance(url, str)
                        and url.startswith("data:")
                    ):
                        x = url.split(";")
                        content[i] = {
                            "type": "image",
                            "source": {
                                "type": "base64",
                                "media_type": x[0].removeprefix("data:"),
                                "data": x[1].removeprefix("base64,"),
                            },
                        }
                        log(
                            "replaced openai-style image_url with anthropic-style image request"
                        )

    return content
