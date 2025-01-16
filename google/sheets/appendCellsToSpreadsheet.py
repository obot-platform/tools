import csv
import io
import os

from gspread.exceptions import APIError
from gspread.utils import ValueInputOption

from auth import gspread_client


def main():
    spreadsheet_id = os.getenv('SPREADSHEET_ID')
    if spreadsheet_id is None:
        raise ValueError("spreadsheet_id must be set")

    raw_data = os.getenv('DATA')
    if raw_data is None:
        raise ValueError("data must be set")
    else:
        data_csv_io = io.StringIO(raw_data)
        data_csv = csv.reader(data_csv_io)
        data = [row for row in data_csv]

    service = gspread_client()
    try:
        spreadsheet = service.open_by_key(
            spreadsheet_id)
        sheet = spreadsheet.sheet1
        sheet.append_rows(data, value_input_option=ValueInputOption.user_entered)
        print("Data written successfully")
    except APIError as err:
        print(err)



if __name__ == "__main__":
    main()
