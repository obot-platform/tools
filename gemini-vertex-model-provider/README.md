Expects you to be authenticated with Google Cloud

example:
```
gcloud auth application-default login
```

## Usage Example

```
gptscript --default-model='gemini-1.0-pro from github.com/gptscript-ai/gemini-vertexai-provider' examples/bob.gpt
```

## Development

Run using the following commands

```
python -m venv .venv
source ./.venv/bin/activate
pip install -r requirements.txt
DEBUG=true ./run.sh
```

```
gptscript --default-model='gemini-1.0-pro from http://127.0.0.1:8000/v1' examples/bob.gpt
```
