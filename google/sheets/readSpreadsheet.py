import asyncio
import os

from auth import gspread_client


async def main():
    spreadsheet_id = os.getenv('SPREADSHEET_ID')
    if spreadsheet_id is None:
        raise ValueError("spreadsheet_id parameter must be set")
    range = os.getenv('RANGE')
    sheet_name = os.getenv('SHEET_NAME')

    service = gspread_client()
    try:
        spreadsheet = service.open_by_key(
            spreadsheet_id)
        if sheet_name is None:
            sheet = spreadsheet.sheet1
        else:
            sheet = spreadsheet.worksheet(sheet_name)
        if range is None:
            values = sheet.get_all_values()
        else:
            values = sheet.get(range)

        if not values:
            print("No data found.")
            return
        else:
            for row in values:
                print(row)

    except Exception as err:
        print(err)


def get_cell_reference(row_idx, col_idx):
    col_letter = chr(col_idx + 64)
    return f"{col_letter}{row_idx}"


if __name__ == "__main__":
    asyncio.run(main())
