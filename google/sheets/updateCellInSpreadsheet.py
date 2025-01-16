import os

from auth import gspread_client


def main():
    spreadsheet_id = os.getenv('SPREADSHEET_ID')
    if spreadsheet_id is None:
        raise ValueError("spreadsheet_id parameter must be set")
    cell = os.getenv('CELL')
    if cell is None:
        raise ValueError("cell parameter must be set")

    data = os.getenv('DATA')
    if data is None:
        raise ValueError("data parameter must be set")

    service = gspread_client()
    try:
        spreadsheet = service.open_by_key(
            spreadsheet_id)
        sheet = spreadsheet.sheet1
        sheet.update_acell(cell, data)
        print("Data written successfully")
    except Exception as err:
        print(err)




if __name__ == "__main__":
    main()
