import json
import os
from validate import configure, validate
import claude3_provider_common
from anthropic import AsyncAnthropicBedrock
from fastapi import FastAPI, Request
from contextlib import asynccontextmanager
from fastapi.responses import JSONResponse, StreamingResponse

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


@app.get("/v1/models")
async def list_models() -> JSONResponse:
    return await claude3_provider_common.list_models(client)


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
