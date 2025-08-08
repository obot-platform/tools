import os
import re
from typing import Annotated

from fastmcp import FastMCP
from fastmcp.exceptions import ToolError
from fastmcp.server.dependencies import get_http_headers
from pydantic import BaseModel, Field
from starlette.requests import Request
from starlette.responses import JSONResponse
from .helper import get_google_client, get_gspread_client
from googleapiclient.errors import HttpError
import gspread
from gspread.utils import a1_range_to_grid_range, rowcol_to_a1, ValueInputOption
from gspread.exceptions import APIError


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
    max_results: Annotated[int, Field(description="Maximum number of results to return. default is 10.", ge=1, le=100, default=10)] = 10,
    page_token: Annotated[str | None, Field(description="Token for pagination, pass the nextPageToken from the previous response to get the next page.")] = None,
) -> dict:
    """
    List spreadsheets(both owned and shared) in the user's Google Drive.
    Returns a list of spreadsheets and a nextPageToken for pagination.
    """
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
    name="list_worksheets",
    annotations={
        "readOnlyHint": True,
    },
)
def list_worksheets_tool(
    spreadsheet_id: Annotated[str, Field(description="The ID of the spreadsheet to list worksheets from.")],
) -> dict:
    """
    List all worksheets (tabs) in a specific spreadsheet.
    Returns worksheet names, IDs, and basic properties.
    """
    token = _get_access_token()
    gspread_client = get_gspread_client(token)
    
    try:
        spreadsheet = gspread_client.open_by_key(spreadsheet_id)
        
        worksheets = []
        for ws in spreadsheet.worksheets():
            worksheets.append({
                "id": ws.id,
                "title": ws.title,
                "index": ws.index,
                "row_count": ws.row_count,
                "col_count": ws.col_count
            })
        
        return {
            "spreadsheet_id": spreadsheet_id,
            "spreadsheet_title": spreadsheet.title,
            "worksheets": worksheets
        }
        
    except HttpError as err:
        raise ToolError(f"HttpError listing worksheets: {err}")
    except Exception as err:
        raise ToolError(f"Unexpected error listing worksheets: {err}")


@mcp.tool(name="create_worksheet")
def create_worksheet_tool(
    spreadsheet_id: Annotated[str, Field(description="The ID of the spreadsheet to create a worksheet in.")],
    worksheet_name: Annotated[str, Field(description="The name of the new worksheet.")],
    rows: Annotated[int, Field(description="Number of rows for the new worksheet.", ge=1, le=1000, default=100)] = 100,
    cols: Annotated[int, Field(description="Number of columns for the new worksheet.", ge=1, le=100, default=26)] = 26,
) -> str:
    """
    Create a new worksheet (tab) in an existing spreadsheet.
    Returns the ID of the created worksheet.
    """
    token = _get_access_token()
    gspread_client = get_gspread_client(token)
    
    try:
        spreadsheet = gspread_client.open_by_key(spreadsheet_id)
        
        # Create the new worksheet
        worksheet = spreadsheet.add_worksheet(
            title=worksheet_name, 
            rows=rows, 
            cols=cols
        )
        
        return f"Worksheet '{worksheet_name}' created with ID: {worksheet.id}"
        
    except HttpError as err:
        raise ToolError(f"HttpError creating worksheet: {err}")
    except Exception as err:
        raise ToolError(f"Unexpected error creating worksheet: {err}")


@mcp.tool(
    name="create_spreadsheet",
)
def create_spreadsheet_tool(
    spreadsheet_name: Annotated[str, Field(description="The name of the spreadsheet to create.")],
) -> str:
    """
    Create a new spreadsheet in the user's personal Google Drive.
    Returns the ID of the created spreadsheet.
    """
    token = _get_access_token()
    google_client = get_google_client(token, service_name="sheets", version="v4")

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
    worksheet_name: Annotated[str | None, Field(description="The name of the worksheet to read. If not provided, the first worksheet will be read.", default=None)] = None,
    cell_range: Annotated[str | None, Field(description="The range of cells of the worksheet to read.", default=None)] = None,
    read_tables: Annotated[bool, Field(description="Whether to read the spreadsheet as tables. Cell ranges are not supported when reading tables.", default=False)] = False,
) -> dict:
    """
    Read a spreadsheet by its ID.
    Returns a list of tables if `read_tables` is True, otherwise returns a list of dict, each dict contains the row number, column number, and value of the cells.
    """
    token = _get_access_token()
    gspread_client = get_gspread_client(token)

    try:
        spreadsheet = gspread_client.open_by_key(
            spreadsheet_id)
        
        if read_tables and cell_range is not None:
            raise ToolError("Cell ranges are not supported when reading tables.")
        
        sheet = spreadsheet.worksheet(worksheet_name) if worksheet_name else spreadsheet.sheet1
        values = sheet.get_all_values(pad_values=True) if cell_range is None else sheet.get(cell_range, pad_values=True)

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
    """
    Delete a spreadsheet by its ID. The operation is only valid for spreadsheets owned by the user.
    """
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
    cells_to_update: Annotated[list[CellUpdate], Field(description="List of cells to update. Each item specifies a cell address and its new value.")],
    worksheet_name: Annotated[str | None, Field(description="The name of the worksheet to update. If not provided, the first worksheet will be updated.")] = None,
    
) -> str:
    """
    Update cells in a spreadsheet. If the cell is out of bounds, the sheet will be automatically expanded to fit the cell.
    """
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
            try:
                # Use batch_update for efficiency
                sheet.batch_update(updates, value_input_option=ValueInputOption.user_entered)
            except APIError as e:
                # If it's an out-of-bounds error, expand the sheet and retry
                if "out of bounds" in str(e).lower():
                    # Find the maximum row and column needed
                    max_row = 0
                    max_col = 0
                    
                    for cell_update in cells_to_update:
                        try:
                            # Parse cell address to get row/column
                            row, col = gspread.utils.a1_to_rowcol(cell_update.cell)
                            max_row = max(max_row, row)
                            max_col = max(max_col, col)
                        except ValueError:
                            # Skip invalid cell addresses (should be caught by validation)
                            continue
                    
                    # Expand sheet if needed
                    if max_row > sheet.row_count:
                        sheet.add_rows(max_row - sheet.row_count)
                    if max_col > sheet.col_count:
                        sheet.add_cols(max_col - sheet.col_count)
                    
                    # Retry the update
                    sheet.batch_update(updates, value_input_option=ValueInputOption.user_entered)
                else:
                    raise ToolError(f"Error updating cells: {e}")
            
        updated_count = len(updates)
        return f"Successfully updated {updated_count} cell(s) in spreadsheet {spreadsheet_id}"
        
    except HttpError as err:
        raise ToolError(f"HttpError updating cells: {err}")
    except Exception as err:
        raise ToolError(f"Unexpected error updating cells: {err}")


@mcp.tool(name="update_range_with_formula")
def update_range_with_formula_tool(
    spreadsheet_id: Annotated[str, Field(description="The ID of the spreadsheet to update.")],
    target_range: Annotated[str, Field(description="Range to fill with formulas in A1 notation (e.g., 'D2:D100', 'A5:Z5')")],
    formula_template: Annotated[str, Field(description="Formula template with placeholders: {row} for row number, {col} for column letter. Must reference valid cells like 'B{row}', 'A{row}+C{row}', '{col}1', etc. Examples: '=B{row}-C{row}' or '={col}2*2'")],
    worksheet_name: Annotated[str | None, Field(description="The name of the worksheet to update. If not provided, the first worksheet will be updated.")] = None,
) -> str:
    """
    Efficiently fills a range of cells with formula templates.
    If the range is out of bounds, the sheet will be automatically expanded to fit the range.
    Supports placeholders: {row} for row number, {col} for column letter.
    Automatically expands sheet if needed.
    """
    # Validate formula template manually since we removed the Pydantic model
    
    # Fast check for most common invalid patterns
    if '{col}{row}' in formula_template or '{row}{col}' in formula_template:
        raise ToolError(f"Invalid formula template: '{formula_template}' contains invalid pattern. Use either fixed column with {{row}} (e.g., 'A{{row}}') or fixed row with {{col}} (e.g., '{{col}}1')")
    
    # Check that placeholders are valid (only if they exist)
    if '{' in formula_template:
        valid_placeholders = ['{row}', '{col}']
        used_placeholders = re.findall(r'\{[^}]+\}', formula_template)
        
        for placeholder in used_placeholders:
            if placeholder not in valid_placeholders:
                raise ToolError(f"Invalid placeholder '{placeholder}'. Only {{row}} and {{col}} are supported")

    token = _get_access_token()
    gspread_client = get_gspread_client(token)

    try:
        # Open the spreadsheet
        spreadsheet = gspread_client.open_by_key(spreadsheet_id)
        sheet = spreadsheet.worksheet(worksheet_name) if worksheet_name else spreadsheet.sheet1
        
        # Ensure formula starts with =
        if not formula_template.startswith("="):
            formula_template = f"={formula_template}"
        
        # Parse the target range
        try:
            grid = a1_range_to_grid_range(target_range)
            start_col, start_row, end_col, end_row = (
                grid["startColumnIndex"],
                grid["startRowIndex"], 
                grid["endColumnIndex"],
                grid["endRowIndex"],
            )
        except Exception as err:
            raise ToolError(f"Error parsing target_range {target_range}: {err}")

        # Generate formulas for the range
        formulas = []
        for r in range(start_row, end_row):
            row_formulas = []
            for c in range(start_col, end_col):
                # Get the column letter ("A", "B", ..., "AA", etc.)
                col_letter = rowcol_to_a1(1, c + 1)[:-1]  # +1 because rowcol_to_a1 is 1-indexed
                formula = formula_template.format(row=r + 1, col=col_letter)
                row_formulas.append(formula)
            formulas.append(row_formulas)

        # Update the range with formulas
        try:
            sheet.update(
                target_range,
                formulas,
                value_input_option=ValueInputOption.user_entered,
            )
        except APIError as e:
            # If it's an out-of-bounds error, expand the sheet and retry
            if "out of bounds" in str(e).lower():
                needed_rows = end_row - sheet.row_count
                if needed_rows > 0:
                    sheet.add_rows(needed_rows)
                needed_cols = end_col - sheet.col_count
                if needed_cols > 0:
                    sheet.add_cols(needed_cols)
                
                # Retry the update
                sheet.update(
                    target_range,
                    formulas,
                    value_input_option=ValueInputOption.user_entered,
                )
            else:
                raise ToolError(f"Error updating range with formula: {e}")
        
        cell_count = (end_row - start_row) * (end_col - start_col)
        return f"Successfully updated {cell_count} cells in range {target_range} with formula template: {formula_template}"
        
    except HttpError as err:
        raise ToolError(f"HttpError updating range with formula: {err}")
    except Exception as err:
        raise ToolError(f"Unexpected error updating range with formula: {err}")


@mcp.tool(name="append_row")
def append_row_tool(
    spreadsheet_id: Annotated[str, Field(description="The ID of the spreadsheet to append to.")],
    row_values: Annotated[list[str], Field(description="Values to append as a new row. Each value will be placed in sequential columns (A, B, C, etc.)")],
    worksheet_name: Annotated[str | None, Field(description="The name of the worksheet to append to. If not provided, the first worksheet will be appended to.")] = None,
) -> str:
    """
    Append a new row to the end of existing data in a spreadsheet.
    Automatically finds the next empty row and adds the data there.
    """
    token = _get_access_token()
    gspread_client = get_gspread_client(token)

    try:
        # Open the spreadsheet
        spreadsheet = gspread_client.open_by_key(spreadsheet_id)
        sheet = spreadsheet.worksheet(worksheet_name) if worksheet_name else spreadsheet.sheet1
        
        # Use gspread's built-in append_row method
        # This automatically finds the next empty row and handles expansion
        try:
            sheet.append_row(row_values, value_input_option=ValueInputOption.user_entered)
        except APIError as e:
            # If there's an issue, try to expand the sheet if needed
            if "out of bounds" in str(e).lower():
                # Calculate needed columns
                needed_cols = len(row_values)
                if needed_cols > sheet.col_count:
                    sheet.add_cols(needed_cols - sheet.col_count)
                
                # Retry append
                sheet.append_row(row_values, value_input_option=ValueInputOption.user_entered)
            else:
                raise ToolError(f"Error appending row: {e}")
        
        # Find the row number where data was added (for confirmation)
        total_rows = len(sheet.get_all_values())
        
        return f"Successfully appended row to spreadsheet {spreadsheet_id} at row {total_rows}"
        
    except HttpError as err:
        raise ToolError(f"HttpError appending row: {err}")
    except Exception as err:
        raise ToolError(f"Unexpected error appending row: {err}")


@mcp.tool(name="clear_range")
def clear_range_tool(
    spreadsheet_id: Annotated[str, Field(description="The ID of the spreadsheet to clear data from.")],
    cell_range: Annotated[str, Field(description="Range of cells to clear in A1 notation (e.g., 'A1:C10', 'D2:D100')")],
    worksheet_name: Annotated[str | None, Field(description="The name of the worksheet to clear data from. If not provided, the first worksheet will be used.")] = None,
) -> str:
    """
    Clear content from a range of cells in a spreadsheet.
    Removes values but preserves formatting.
    """
    token = _get_access_token()
    gspread_client = get_gspread_client(token)

    try:
        # Open the spreadsheet
        spreadsheet = gspread_client.open_by_key(spreadsheet_id)
        sheet = spreadsheet.worksheet(worksheet_name) if worksheet_name else spreadsheet.sheet1
        
        # Clear the range
        sheet.batch_clear([cell_range])
        
        return f"Successfully cleared range {cell_range} in spreadsheet {spreadsheet_id}"
        
    except HttpError as err:
        raise ToolError(f"HttpError clearing range: {err}")
    except Exception as err:
        raise ToolError(f"Unexpected error clearing range: {err}")


def streamable_http_server():
    """Main entry point for the Google Sheets MCP server."""
    mcp.run(
        transport="streamable-http",  # fixed to streamable-http
        host="0.0.0.0",
        port=PORT,
        path=MCP_PATH,
    )


if __name__ == "__main__":
    streamable_http_server()
