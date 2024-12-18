import json
import os
from typing import AsyncIterable, List

import google.auth.exceptions
import vertexai.preview.generative_models as generative_models
from fastapi import FastAPI, HTTPException, Request
from fastapi.encoders import jsonable_encoder
from fastapi.responses import JSONResponse, StreamingResponse
from fastapi.routing import APIRouter
from openai.types.chat import ChatCompletionChunk
from openai.types.chat.chat_completion_chunk import Choice, ChoiceDelta
from vertexai.preview.generative_models import Content, FunctionDeclaration, GenerationConfig, GenerationResponse, \
    GenerativeModel, Part, Tool

debug = os.environ.get("GPTSCRIPT_DEBUG", "false") == "true"
app = FastAPI()
router = APIRouter()

uri = "http://127.0.0.1:" + os.environ.get("PORT", "8000")



def log(*args):
    if debug:
        print(*args)


@app.middleware("http")
async def log_body(request: Request, call_next):
    body = await request.body()
    log("REQUEST BODY: ", body)
    return await call_next(request)


@app.post("/")
@app.get("/")
async def get_root():
    return uri


@app.get("/v1/models")
def list_models() -> JSONResponse:
    content = {
        "data": [
            {
                "id": "gemini-1.0-pro",
                "name": "Gemini 1.0 Pro",
            },
            {
                "id": "gemini-1.5-pro",
                "name": "Gemini 1.5 Pro",
            },
            {
                "id": "gemini-1.5-pro-preview-0409",
                "name": "Gemini 1.5 Pro Preview 0409"
            }
        ]
    }
    return JSONResponse(content=content)


async def map_tools(req_tools: List | None = None) -> List[Tool] | None:
    if req_tools is None or len(req_tools) < 1:
        return None

    function_declarations = []
    for tool in req_tools:
        parameters = tool['function'].get('parameters', {
            "properties": {},
            "type": "object"
        })

        function_declarations.append(
            FunctionDeclaration(
                name=tool["function"]["name"],
                description=tool["function"]["description"],
                parameters=parameters,
            )
        )

    tools: list["Tool"] = [Tool.from_function_declarations(function_declarations)]

    return tools


def merge_consecutive_dicts_with_same_value(list_of_dicts, key) -> list[dict]:
    merged_list = []
    index = 0
    while index < len(list_of_dicts):
        current_dict = list_of_dicts[index]
        value_to_match = current_dict.get(key)
        compared_index = index + 1
        while compared_index < len(list_of_dicts) and list_of_dicts[compared_index].get(key) == value_to_match:
            list_of_dicts[compared_index]["content"] = current_dict["content"] + "\n" + list_of_dicts[compared_index][
                "content"]
            current_dict.update(list_of_dicts[compared_index])
            compared_index += 1
        merged_list.append(current_dict)
        index = compared_index
    return merged_list


async def map_messages(req_messages: list) -> list[Content] | None:
    messages: list[Content] = []
    log(req_messages)

    if req_messages is not None:
        system: str = """
You are a task oriented system.
Be as brief as possible when answering the user.
Only give the required answer.
Do not give your thought process.
Use functions or tools as needed to complete the tasks given to you.
You are referred to as a tool.
Do not call functions or tools unless you need to.
Ensure you are passing the correct arguments to the functions or tools you call.
Do not move on to the next task until the current task is completed.
Do not make up arguments for tools.
Call functions one at a time to make sure you have the correct inputs.
"""
        req_messages = [
                           {"role": "system", "content": system},
                           {"role": "model", "content": "Understood."}
                       ] + req_messages

        for message in req_messages:
            match message["role"]:
                case "system":
                    message['role'] = "user"
                case "user":
                    message['role'] = "user"
                case "assistant":
                    message['role'] = "model"
                case "model":
                    message['role'] = "model"
                case "tool":
                    message['role'] = "function"
                case _:
                    message['role'] = "user"
        req_messages = merge_consecutive_dicts_with_same_value(req_messages, "role")

        for message in req_messages:
            if 'tool_call_id' in message.keys():
                convert_message = Content(
                    role=message['role'],
                    parts=[Part.from_function_response(
                        name=message.get('name', ''),
                        response={
                            'name': message.get('name', ''),
                            'content': message['content']
                        },
                    )]
                )
            elif 'tool_calls' in message.keys():
                tool_call_parts: list[Part] = []
                for tool_call in message['tool_calls']:
                    function_call = {
                        "functionCall": {
                            "name": tool_call['function']['name'],
                            "args": json.loads(tool_call['function']['arguments'])
                        }
                    }
                    tool_call_parts.append(Part.from_dict(function_call))
                convert_message = Content(
                    role=message['role'],
                    parts=tool_call_parts
                )
            elif 'content' in message.keys():
                convert_message = Content(
                    role=message['role'],
                    parts=[Part.from_text(message["content"])]
                )
            messages.append(convert_message)

        return messages

    return None


@app.post("/v1/chat/completions")
async def chat_completion(request: Request):
    data = await request.body()
    data = json.loads(data)

    req_tools = data.get("tools", None)
    tools: list[Tool] | None = None
    if req_tools is not None:
        tools = await map_tools(req_tools)

    req_messages = data["messages"]
    messages = await map_messages(req_messages)

    temperature = data.get("temperature", None)
    if temperature is not None:
        temperature = float(temperature)

    stream = data.get("stream", False)

    top_k = data.get("top_k", None)
    if top_k is not None:
        top_k = float(top_k)

    top_p = data.get("top_p", None)
    if top_p is not None:
        top_p = float(top_p)

    max_output_tokens = data.get("max_tokens", None)
    if max_output_tokens is not None:
        max_output_tokens = float(max_output_tokens)

    log()
    log("GEMINI TOOLS: ", tools)
    try:
        model = GenerativeModel(data["model"])
    except google.auth.exceptions.GoogleAuthError as e:
        log("AUTH ERROR: ", e)
        raise HTTPException(status_code=401,
                            detail="Authentication error. Please ensure you are properly authenticated with GCP and have the correct project configured.")
    except Exception as e:
        log("ERROR: ", e)
        log(type(e))
        raise HTTPException(status_code=500, detail=str(e))
    try:
        response = model.generate_content(
            contents=messages,
            tools=tools,
            generation_config=GenerationConfig(
                temperature=temperature,
                top_k=top_k,
                top_p=top_p,
                candidate_count=1,
                max_output_tokens=max_output_tokens,
            ),
            safety_settings={
                generative_models.HarmCategory.HARM_CATEGORY_HATE_SPEECH: generative_models.HarmBlockThreshold.BLOCK_ONLY_HIGH,
                generative_models.HarmCategory.HARM_CATEGORY_DANGEROUS_CONTENT: generative_models.HarmBlockThreshold.BLOCK_ONLY_HIGH,
                generative_models.HarmCategory.HARM_CATEGORY_SEXUALLY_EXPLICIT: generative_models.HarmBlockThreshold.BLOCK_ONLY_HIGH,
                generative_models.HarmCategory.HARM_CATEGORY_HARASSMENT: generative_models.HarmBlockThreshold.BLOCK_ONLY_HIGH,
            }
        )
        if not stream:
            return JSONResponse(content=jsonable_encoder(response))

        return StreamingResponse(to_chunk(data['model'], response), media_type="application/x-ndjson")

    except Exception as e:
        log("ERROR: ", e)
        raise HTTPException(status_code=500, detail=str(e))


async def to_chunk(model: str, response: GenerationResponse) -> AsyncIterable[str]:
    mapped_chunk = map_resp(model, response)
    if mapped_chunk is None:
        yield "data: " + json.dumps({}) + "\n\n"
    else:
        log("RESPONSE CHUNK: ", mapped_chunk.model_dump_json())
        yield "data: " + mapped_chunk.model_dump_json() + "\n\n"


def map_resp(model: str, chunk: GenerationResponse) -> ChatCompletionChunk | None:
    tool_calls = []
    if len(chunk.candidates) > 0:
        if len(chunk.candidates[0].function_calls) > 0:
            parts = chunk.candidates[0].to_dict().get('content', {}).get('parts', [])

            for idx, part in enumerate(parts):
                call = part.get('function_call', None)
                if not call:
                    continue

                tool_calls.append({
                    "index": idx,
                    "id": call['name'] + "_" + str(idx),
                    "function": {
                        "name": call['name'],
                        "arguments": json.dumps(call['args'])
                    },
                    "type": "function"
                })

        try:
            content = chunk.candidates[0].content.text
        except:
            content = None

        match chunk.candidates[0].content.role:
            case "system":
                role = "user"
            case "user":
                role = "user"
            case "assistant":
                role = "model"
            case "model":
                role = "assistant"
            case "function":
                role = "tool"
            case _:
                role = "user"

        try:
            if len(tool_calls) > 0:
                finish_reason = "tool_calls"
            else:
                finish_reason = map_finish_reason(str(chunk.candidates[0].finish_reason))
        except KeyError:
            finish_reason = None

        log("FINISH_REASON: ", finish_reason)

        resp = ChatCompletionChunk(
            id="0",
            choices=[
                Choice(
                    delta=ChoiceDelta(
                        content=content,
                        tool_calls=tool_calls,
                        role=role
                    ),
                    finish_reason=finish_reason,
                    index=0,
                )
            ],
            created=0,
            model=model,
            object="chat.completion.chunk",
        )
        return resp

    return None


def map_finish_reason(finish_reason: str) -> str:
    if (finish_reason == "ERROR"):
        return "stop"
    elif (finish_reason == "FINISH_REASON_UNSPECIFIED" or finish_reason == "STOP"):
        return "stop"
    elif finish_reason == "SAFETY":
        return "content_filter"
    elif finish_reason == "STOP":
        return "stop"
    elif finish_reason == "0":
        return "stop"
    elif finish_reason == "1":
        return "stop"
    elif finish_reason == "2":
        return "length"
    elif finish_reason == "3":
        return "content_filter"
    elif finish_reason == "4":
        return "content_filter"
    elif finish_reason == "5":
        return "stop"
    elif finish_reason == "6":
        return "content_filter"
    elif finish_reason == "7":
        return "content_filter"
    elif finish_reason == "8":
        return "content_filter"
    # elif finish_reason == None:
    #     return "tool_calls"
    return finish_reason


if __name__ == "__main__":
    import uvicorn
    import asyncio

    try:
        uvicorn.run("main:app", host="127.0.0.1", port=int(os.environ.get("PORT", "8000")),
                log_level="debug" if debug else "critical", access_log=debug)
    except (KeyboardInterrupt, asyncio.CancelledError):
        pass
