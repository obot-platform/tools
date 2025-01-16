import sys
import json
import os
import boto3


def configure():
    # Check if any is empty
    try:
        if not all(
            [
                os.environ["OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_ACCESS_KEY_ID"],
                os.environ["OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_SECRET_ACCESS_KEY"],
                os.environ["OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_SESSION_TOKEN"],
                os.environ["OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_REGION"],
            ]
        ):
            raise KeyError(
                "Please set OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_ACCESS_KEY_ID, OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_SECRET_ACCESS_KEY, OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_SESSION_TOKEN, OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_REGION",
                file=sys.stderr,
            )
    except KeyError as e:
        raise KeyError(f"Authentication failed: {e}")

    os.environ["AWS_ACCESS_KEY_ID"] = os.environ.get(
        "OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_ACCESS_KEY_ID"
    )
    os.environ["AWS_SECRET_ACCESS_KEY"] = os.environ.get(
        "OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_SECRET_ACCESS_KEY"
    )
    os.environ["AWS_SESSION_TOKEN"] = os.environ.get(
        "OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_SESSION_TOKEN"
    )
    os.environ["AWS_REGION"] = os.environ["AWS_DEFAULT_REGION"] = os.environ.get(
        "OBOT_ANTHROPIC_BEDROCK_MODEL_PROVIDER_REGION"
    )


def validate():
    # Check authentication
    try:
        c = boto3.client("sts")
        _ = c.get_caller_identity()
    except Exception as e:
        ex = json.dumps({"error": f"AWS authentication failed: {e}"})
        print(ex, file=sys.stdout)
        print(ex, file=sys.stderr)
        sys.exit(0)
    print(
        "Anthropic on AWS Bedrock model provider credentials validated successfully.",
        file=sys.stderr,
    )


if __name__ == "__main__":
    configure()
    validate()
