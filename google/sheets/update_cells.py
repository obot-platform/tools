import json
import os

import gspread
from gspread.worksheet import Cell
from gspread.utils import ValueInputOption
from auth import gspread_client


def update_cells():
    spreadsheet_id = os.getenv("SPREADSHEET_ID")
    if spreadsheet_id is None:
        raise ValueError("spreadsheet_id parameter must be set")
    sheet_name = os.getenv("SHEET_NAME")

    update_cells = os.getenv("UPDATE_CELLS")
    if update_cells is None:
        raise ValueError("update_cells parameter must be set")
    try:
        update_cells_object = json.loads(update_cells)
        print(update_cells_object)
    except json.JSONDecodeError as err:
        print(f"JSON parsing error for update_cells input: {err}")
        exit(1)

    service = gspread_client()

    data_cells = []
    formula_cells = []

    try:
        for cell in update_cells_object:
            if not all(key in cell for key in ("coordinate", "data")):
                raise ValueError(f"Missing required keys in cell dictionary: {cell}")
            row, col = gspread.utils.a1_to_rowcol(cell["coordinate"])
            cell_object = Cell(row=row, col=col, value=cell["data"])
            if cell["data"].startswith("="):
                formula_cells.append(cell_object)
            else:
                data_cells.append(cell_object)
    except Exception as err:
        print(f"Error mapping input to Cell: {err}")
        exit(1)

    try:
        spreadsheet = service.open_by_key(spreadsheet_id)
        if sheet_name is not None:
            sheet = spreadsheet.worksheet(sheet_name)
        else:
            sheet = spreadsheet.sheet1
    except Exception as err:
        print(f"Error opening spreadsheet: {err}")
        exit(1)

    try:
        if len(data_cells) > 0:
            sheet.update_cells(data_cells)
        if len(formula_cells) > 0:
            sheet.update_cells(
                formula_cells, value_input_option=ValueInputOption.user_entered
            )
    except Exception as err:
        print(f"Error updating cells: {err}")
        exit(1)

    print("Data written successfully")


if __name__ == "__main__":
    update_cells()
