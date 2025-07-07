import os

from auth import gspread_client
from gspread.utils import a1_range_to_grid_range, rowcol_to_a1, ValueInputOption
from gspread.exceptions import APIError


def update_with_formula():
    spreadsheet_id = os.getenv("SPREADSHEET_ID")
    if spreadsheet_id is None:
        raise ValueError("spreadsheet_id parameter must be set")

    sheet_name = os.getenv("SHEET_NAME")

    target_range = os.getenv("TARGET_RANGE")
    if target_range is None:
        raise ValueError("target_range parameter must be set")

    formula_template = os.getenv("FORMULA_TEMPLATE")
    if formula_template is None:
        raise ValueError("formula_template parameter must be set")
    if not formula_template.startswith("="):
        formula_template = f"={formula_template}"

    service = gspread_client()

    try:
        spreadsheet = service.open_by_key(spreadsheet_id)
        if sheet_name is not None:
            sheet = spreadsheet.worksheet(sheet_name)
        else:
            sheet = spreadsheet.sheet1
    except Exception as err:
        print(f"Error opening spreadsheet: {err}")
        exit(1)

    """
    Injects formulas into the given range on a sheet.

    :param spreadsheet_id: the “key” of the spreadsheet (from its URL)
    :param sheet_name: the name (tab) within the spreadsheet
    :param target_range: A1 notation of the cells to fill (e.g. "D2:D100" or "A5:Z5")
    :param formula_template:
        a template for Google Sheets formulas that supports:
          - {row}      → the row number
          - {col}      → the column letter
        e.g. "=B{row}-C{row}"  or  "={col}2*2"
    """

    try:
        grid = a1_range_to_grid_range(target_range)
        start_col, start_row, end_col, end_row = (
            grid["startColumnIndex"],
            grid["startRowIndex"],
            grid["endColumnIndex"],
            grid["endRowIndex"],
        )
    except Exception as err:
        print(f"Error parsing target_range {target_range}: {err}")
        return

    formulas = []
    for r in range(start_row, end_row):
        row_formulas = []
        for c in range(start_col, end_col):
            # get the column letter ("A", "B", ..., "AA", etc.)
            col_letter = rowcol_to_a1(1, c)[:-1]
            row_formulas.append(formula_template.format(row=r + 1, col=col_letter))
        formulas.append(row_formulas)

    try:
        sheet.update(
            target_range,
            formulas,
            value_input_option=ValueInputOption.user_entered,
        )
    except APIError as e:
        # if it's an out-of-bounds error, grow and retry
        if "out of bounds" in str(e):
            needed_rows = end_row - sheet.row_count
            if needed_rows > 0:
                sheet.add_rows(needed_rows)
            needed_cols = end_col - sheet.col_count
            if needed_cols > 0:
                sheet.add_cols(needed_cols)
            sheet.update(
                target_range,
                formulas,
                value_input_option=ValueInputOption.USER_ENTERED,
            )
        else:
            print(f"Error updating cells with formula {formula_template}: {e}")
            exit(1)
    print(f"Successfully updated {target_range} with formula {formula_template}")


if __name__ == "__main__":
    update_with_formula()
