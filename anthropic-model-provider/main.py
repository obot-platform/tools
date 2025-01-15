import json
import os

import anthropic.pagination
from anthropic import AsyncAnthropic
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse, StreamingResponse

import anthropic_common

debug = os.environ.get("GPTSCRIPT_DEBUG", "false") == "true"
client = AsyncAnthropic(
    api_key=os.environ.get("OBOT_ANTHROPIC_MODEL_PROVIDER_API_KEY", "")
)
app = FastAPI()
uri = "http://127.0.0.1:" + os.environ.get("PORT", "8000")


def log(*args):
    if debug:
        print(*args)


@app.middleware("http")
async def log_body(request: Request, call_next):
    body = await request.body()
    log("HTTP REQUEST BODY: ", body)
    return await call_next(request)


@app.post("/")
@app.get("/")
async def get_root():
    return uri


@app.get("/v1/models")
async def list_models() -> JSONResponse:
    try:
        resp: anthropic.pagination.AsyncPage[
            anthropic.types.ModelInfo
        ] = await client.models.list(limit=20)
        return JSONResponse(
            content={
                "object": "list",
                "data": [
                    set_model_usage(model.model_dump(exclude={"created_at"}))
                    for model in resp.data
                ],
            }
        )
    except Exception as e:
        return JSONResponse(content={"error": e}, status_code=500)


def set_model_usage(model: dict) -> dict:
    model["metadata"] = {"usage": "llm"}
    return model


@app.post("/v1/chat/completions")
async def completions(request: Request) -> StreamingResponse:
    data = await request.body()
    return await anthropic_common.completions(client, json.loads(data))


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
