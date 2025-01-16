import json
import os
from contextlib import asynccontextmanager

import boto3
import claude3_provider_common
from anthropic import AsyncAnthropicBedrock
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse, StreamingResponse

from validate import configure, validate

debug = os.environ.get("GPTSCRIPT_DEBUG", "false") == "true"


def log(*args):
    if debug:
        print(*args)


client: AsyncAnthropicBedrock | None = None


@asynccontextmanager
async def lifespan(a: FastAPI):
    try:
        configure()
        validate()
        global client
        client = AsyncAnthropicBedrock()
    except Exception as ex:
        raise ex
    yield  # App shutdown
    await client.close()


app = FastAPI(lifespan=lifespan)
uri = "http://127.0.0.1:" + os.environ.get("PORT", "8000")


@app.middleware("http")
async def log_body(request: Request, call_next):
    body = await request.body()
    log("HTTP REQUEST BODY: ", body)
    return await call_next(request)


@app.post("/")
@app.get("/")
async def get_root():
    return uri

# https://docs.anthropic.com/en/api/claude-on-amazon-bedrock#list-available-models
@app.get("/v1/models")
async def list_models() -> JSONResponse:
    try:
        bedrock = boto3.client(service_name="bedrock")
        response = bedrock.list_foundation_models(byProvider="anthropic")
    except Exception as e:
        return JSONResponse(content={"error": f"Failed to list models - bedrock client error: {e}"}, status_code=500)
    try:
        if response["ResponseMetadata"]["HTTPStatusCode"] != 200:
            log(f"Failed to list models: {response}")
            return JSONResponse(content={"error": f"Failed to list models - unexpected status code: {response}"},
                                status_code=response["ReponseMetadata"]["HTTPStatusCode"])
    except KeyError as e:
        return JSONResponse(content={"error": f"Failed to list models - bad response: {e}"}, status_code=500)

    try:
        models = []
        for model in response["modelSummaries"]:
            if model["modelLifecycle"]["status"] != "ACTIVE" or "PROVISIONED" not in model["inferenceTypesSupported"]:
                continue
            models.append({
                "id": model["modelId"],
                "name": f"AWS Bedrock Anthropic {model['modelName']}",
                "metadata": {"usage": "llm"},
            })
        return JSONResponse(content={"object": "list", "data": models})
    except KeyError as e:
        log(f"Bad model list: {e}")
        return JSONResponse(content={"error": f"Failed to list models - bad model list: {e}"}, status_code=500)


@app.post("/v1/chat/completions")
async def completions(request: Request) -> StreamingResponse:
    data = await request.body()
    return await claude3_provider_common.completions(client, json.loads(data))


if __name__ == "__main__":
    import uvicorn
    import asyncio

    try:
        uvicorn.run(
            "main:app",
            host="127.0.0.1",
            port=int(os.environ.get("PORT", "8000")),
            workers=4,
            log_level="debug" if debug else "critical",
            access_log=debug,
        )
    except (KeyboardInterrupt, asyncio.CancelledError):
        pass
