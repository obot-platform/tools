import json
import os
from typing import Annotated, Literal, Optional, Union

from fastmcp import FastMCP
from fastmcp.exceptions import ToolError
from fastmcp.server.dependencies import get_http_headers
from pydantic import BaseModel, Field
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
    annotations={
        "readOnlyHint": True,
    },
)
async def list_spreadsheets_tool(
    max_results: Annotated[int, Field(description="...", ge=1, le=100)] = 10,
    page_token: Annotated[str | None, Field(description="Token for pagination, pass the nextPageToken from the previous response to get the next page.")] = None,
) -> dict:
    token = _get_access_token()
    google_client = get_google_client(token, service_name="drive", version="v3")
    
    params = {
        "q": "mimeType='application/vnd.google-apps.spreadsheet'",
        "pageSize": max_results,
        "fields": "nextPageToken, files(id, name, modifiedTime, createdTime)",
        "supportsAllDrives": True,
    }
    if page_token:
        params["pageToken"] = page_token
        
    response = google_client.files().list(**params).execute()
    
    return {
        "files": response.get("files", []),
        "nextPageToken": response.get("nextPageToken")
    }


@mcp.tool(
    name="create_spreadsheet",
)
def create_spreadsheet_tool(
    spreadsheet_name: Annotated[str, Field(description="The name of the spreadsheet to create.")],
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
        raise ToolError(f"HttpError creating spreadsheet: {err}")
    except Exception as e:
        raise ToolError(f"Unexpected creating spreadsheet error: {e}")

@mcp.tool(
    name="read_spreadsheet",
    annotations={
        "readOnlyHint": True,
    },
)
def read_spreadsheet_tool(
    spreadsheet_id: Annotated[str, Field(description="The ID of the spreadsheet to read.")],
    worksheet_name: Annotated[str | None, Field(description="The name of the worksheet to read.")],
    cell_range: Annotated[str | None, Field(description="The range of cells of the worksheet to read.")],
    read_tables: Annotated[bool, Field(description="Whether to read the spreadsheet as tables. Cell ranges are not supported when reading tables.", default=False)] = False,
) -> dict:
    token = _get_access_token()
    gspread_client = get_gspread_client(token)

    try:
        spreadsheet = gspread_client.open_by_key(
            spreadsheet_id)
        
        if read_tables and cell_range is not None:
            raise ToolError("Cell ranges are not supported when reading tables.")
        
        sheet = spreadsheet.worksheet(worksheet_name) if worksheet_name else spreadsheet.sheet1
        values = sheet.get_all_values() if cell_range is None else sheet.get(cell_range)

        if read_tables:
            tables = []
            current_table = []
            for _, row in enumerate(values):
                if not any(cell.strip() for cell in row):
                    if current_table:
                        tables.append(current_table)
                        current_table = []
                else:
                    current_table.append(row)
            if current_table:
                tables.append(current_table)
            
            return {"tables": tables}
        
        if not values:
            return {"rows": 0, "columns": 0, "values": []}
        else:
            return {"rows": len(values), "columns": len(values[0]) if values else 0, "values": values}
        


    except HttpError as err:
        raise ToolError(f"HttpError reading spreadsheet: {err}")
    except Exception as err:
        raise ToolError(f"Unexpected reading spreadsheet error: {err}")


@mcp.tool(
    name="delete_spreadsheet",
)
def delete_spreadsheet_tool(
    spreadsheet_id: Annotated[str, Field(description="The ID of the spreadsheet to delete.")],
) -> str:
    token = _get_access_token()
    google_client = get_google_client(token, service_name="drive", version="v3")
    
    try:
        google_client.files().delete(fileId=spreadsheet_id).execute()
        return f"Spreadsheet with ID {spreadsheet_id} deleted."
    except HttpError as err:
        raise ToolError(f"HttpError deleting spreadsheet: {err}")
    except Exception as err:
        raise ToolError(f"Unexpected deleting spreadsheet error: {err}")


# Pydantic models for request validation
class CellUpdate(BaseModel):
    cell: str = Field(description="Cell address in A1 notation (e.g., 'A1', 'B2', 'C10')")
    value: str = Field(description="Value to set in the cell. Can be text, number, or formula (starting with '=')")


@mcp.tool(name="update_cells")
def update_cells_tool(
    spreadsheet_id: Annotated[str, Field(description="The ID of the spreadsheet to update.")],
    worksheet_name: Annotated[str | None, Field(description="The name of the worksheet to update.")],
    cells_to_update: Annotated[list[CellUpdate], Field(description="List of cells to update. Each item specifies a cell address and its new value.")],
) -> str:
    token = _get_access_token()
    gspread_client = get_gspread_client(token)

    try:
        # Open the spreadsheet
        spreadsheet = gspread_client.open_by_key(spreadsheet_id)
        sheet = spreadsheet.worksheet(worksheet_name) if worksheet_name else spreadsheet.sheet1
        
        # Prepare updates from validated Pydantic models
        updates = []
        for cell_update in cells_to_update:
            # No need for manual validation - Pydantic handles it automatically
            updates.append({
                'range': cell_update.cell,
                'values': [[cell_update.value]]
            })
        
        # Perform batch update
        if updates:
            # Use batch_update for efficiency
            sheet.batch_update(updates, value_input_option='USER_ENTERED')
            
        updated_count = len(updates)
        return f"Successfully updated {updated_count} cell(s) in spreadsheet {spreadsheet_id}"
        
    except HttpError as err:
        raise ToolError(f"HttpError updating cells: {err}")
    except Exception as err:
        raise ToolError(f"Unexpected error updating cells: {err}")

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
