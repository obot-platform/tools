import json

from main import configure, cleanup, log

from vertexai.preview.generative_models import GenerativeModel

from google.auth.exceptions import GoogleAuthError

def validate_credentials():
    try:
        configure()
        test_model()
        log("Credentials are valid")
    except Exception as e:
        print(json.dumps({"error": str(e)}))
        exit(1)
    finally:
        cleanup()

def test_model():
    try:
        _ = GenerativeModel("gemini-1.5-pro")
    except GoogleAuthError as e:
        print(json.dumps({"error": f"Invalid Google Credentials: {str(e)}"}))
        exit(1)
    except Exception as e:
        print(json.dumps({"error": f"Unknown Error: {str(e)}"}))
        exit(1)

if __name__ == "__main__":
    validate_credentials()