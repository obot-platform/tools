import json
import os

from gspread.worksheet import Cell

from auth import gspread_client


def main():
    spreadsheet_id = os.getenv('SPREADSHEET_ID')
    if spreadsheet_id is None:
        raise ValueError("spreadsheet_id parameter must be set")

    update_cells = os.getenv('UPDATE_CELLS')
    if update_cells is None:
        raise ValueError("update_cells parameter must be set")
    try:
        update_cells_object = json.loads(update_cells)
    except json.JSONDecodeError as err:
        print(f"JSON parsing error for update_cells input: {err}")
        exit(1)

    service = gspread_client()

    cells = []
    try:
        for cell in update_cells_object:
            if not all(key in cell for key in ("row", "column", "value")):
                raise ValueError(f"Missing required keys in cell dictionary: {cell}")
            cell_object = Cell(row=cell["row"], col=cell["column"], value=cell["value"])
            cells.append(cell_object)
    except Exception as err:
        print(f"Error mapping input to Cell: {err}")
        exit(1)

    try:
        spreadsheet = service.open_by_key(
            spreadsheet_id)
        sheet = spreadsheet.sheet1
    except Exception as err:
        print(f"Error opening spreadsheet: {err}")
        exit(1)

    try:
        sheet.update_cells(cells)
    except Exception as err:
        print(f"Error updating cells: {err}")
        exit(1)

    print("Data written successfully")


if __name__ == "__main__":
    main()
