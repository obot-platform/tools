import os
from typing import Annotated, Literal, Optional, Union

from fastmcp import FastMCP
from fastmcp.exceptions import ToolError
from fastmcp.server.dependencies import get_http_headers
from pydantic import Field
from starlette.requests import Request
from starlette.responses import JSONResponse
from .helper import get_google_client, get_gspread_client
from googleapiclient.errors import HttpError


# Configure server-specific settings
PORT = int(os.getenv("PORT", 9000))
MCP_PATH = os.getenv("MCP_PATH", "/mcp/google-sheets")

mcp = FastMCP(
    name="GoogleSheetsMCPServer",
    on_duplicate_tools="error",  # Handle duplicate registrations
    on_duplicate_resources="warn",
    on_duplicate_prompts="replace",
)


@mcp.custom_route("/health", methods=["GET"])
async def health_check(request: Request):
    return JSONResponse({"status": "healthy"})


def _get_access_token() -> str:
    headers = get_http_headers()
    access_token = headers.get("x-forwarded-access-token", None)
    if not access_token:
        raise ToolError(
            "No access token found in headers, available headers: " + str(headers)
        )
    return access_token


@mcp.tool(
    name="list_spreadsheets",
)
async def list_spreadsheets_tool(
    max_results: Annotated[
        int, "Maximum number of results to return. default is 10."
    ] = 10,
) -> list[dict]:
    token = _get_access_token()
    google_client = get_google_client(token)

    try:
        files = []
        page_token = None
        q = "mimeType='application/vnd.google-apps.spreadsheet'"
        params = {
            "q": q,
            "pageSize": max_results,
            "fields": "nextPageToken, files(id, name, modifiedTime, createdTime)",
            "pageToken": page_token,
            "supportsAllDrives": True,
        }
        while True:
            response = google_client.files().list(**params).execute()
            files.extend(response.get("files", []))

            # Break if we've reached max_results
            if max_results and len(files) >= max_results:
                files = files[:max_results]
                break

            page_token = response.get("nextPageToken")
            if not page_token:
                break

            params["pageToken"] = page_token

        return files
    except HttpError as err:
        raise HttpError(f"Error listing spreadsheets: {err}")
    except Exception as e:
        raise ToolError(f"Unexpected listing spreadsheets error: {e}")


@mcp.tool(
    name="create_spreadsheet",
)
def create_spreadsheet_tool(
    spreadsheet_name: Annotated[str, "The name of the spreadsheet to create."],
) -> str:
    token = _get_access_token()
    google_client = get_google_client(token)

    props = {
        'properties': {
            'title': spreadsheet_name
        }
    }
    try:
        spreadsheet = google_client.spreadsheets().create(body=props, fields='spreadsheetId').execute()
        return f"Spreadsheet named {spreadsheet_name} created with ID: {spreadsheet.get('spreadsheetId')}"
    except HttpError as err:
        raise HttpError(f"Error creating spreadsheet: {err}")
    except Exception as e:
        raise ToolError(f"Unexpected creating spreadsheet error: {e}")

@mcp.tool(
    name="read_spreadsheet",
)
def read_spreadsheet_tool(
    spreadsheet_id: Annotated[str, "The ID of the spreadsheet to read."],
    worksheet_name: Annotated[str | None, "The name of the worksheet to read."] = None,
    range: Annotated[str | None, "The range of cells of the worksheet to read."] = None,
) -> str:
    token = _get_access_token()
    gspread_client = get_gspread_client(token)

    try:
        spreadsheet = gspread_client.open_by_key(
            spreadsheet_id)
        if worksheet_name is None:
            sheet = spreadsheet.sheet1
        else:
            sheet = spreadsheet.worksheet(worksheet_name)
            
        all_values = sheet.get_all_values()
        if range is None:
            values = sheet.get_all_values()
        else:
            values = sheet.get(range)

        if not values:
            print("No data found.")
            return
        else:
            print(
                f"Overview of the sheet: {len(all_values)} rows and {len(all_values[0])} columns"
            )
            if range is None:
                print("Sheet data:")
            else:
                print(f"Data in range {range}:")
            for row in values:
                print(row)

    except HttpError as err:
        raise HttpError(f"Error reading spreadsheet: {err}")
    except Exception as err:
        raise ToolError(f"Unexpected reading spreadsheet error: {err}")


def streamable_http_server():
    """Main entry point for the Gmail MCP server."""
    mcp.run(
        transport="streamable-http",  # fixed to streamable-http
        host="0.0.0.0",
        port=PORT,
        path=MCP_PATH,
    )


if __name__ == "__main__":
    streamable_http_server()
