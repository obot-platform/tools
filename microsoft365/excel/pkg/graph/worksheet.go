package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/obot-platform/tools/microsoft365/excel/pkg/global"
	"github.com/obot-platform/tools/microsoft365/excel/pkg/util"
	"github.com/microsoft/kiota-abstractions-go/serialization"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/drives"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

type WorkbookInfo struct {
	ID, Name string
}

type Table struct {
	ID, Name string
}

type HTTPErrorBody struct {
	Error struct {
		Code       string `json:"code,omitempty"`
		Message    string `json:"message,omitempty"`
		InnerError struct {
			Code            string `json:"code,omitempty"`
			Message         string `json:"message,omitempty"`
			Date            string `json:"date,omitempty"`
			RequestID       string `json:"request-id,omitempty"`
			ClientRequestID string `json:"client-request-id,omitempty"`
		} `json:"innerError,omitempty"`
	} `json:"error,omitempty"`
}

type UpdateBody struct {
	Values   [][]any `json:"values,omitempty"`
	Formulas [][]any `json:"formulas,omitempty"`
}

func (u *UpdateBody) AppendColumnCellToValues(newColumnCell []any) {
	u.Values = append(u.Values, newColumnCell)
	u.Formulas = append(u.Formulas, []any{nil})
}

func (u *UpdateBody) AppendColumnCellToFormulas(newColumnCell []any) {
	u.Formulas = append(u.Formulas, newColumnCell)
	u.Values = append(u.Values, []any{nil})
}

func (u *UpdateBody) AppendRowToValues(newRow []any) {
	u.Values = append(u.Values, newRow)
}

func (u *UpdateBody) AppendRowToFormulas(newRow []any) {
	u.Formulas = append(u.Formulas, newRow)
}

func numberToColumnLetter(n int) string {
	if n <= 0 {
		return ""
	}

	column := ""
	for n > 0 {
		n--
		column = string(rune('A'+(n%26))) + column
		n /= 26
	}

	return column
}

func ListWorkbooks(ctx context.Context, c *msgraphsdkgo.GraphServiceClient) ([]WorkbookInfo, error) {
	drive, err := c.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Get the root folder items
	root, err := c.Drives().ByDriveId(util.Deref(drive.GetId())).Root().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	var infos []WorkbookInfo
	items, err := c.Drives().ByDriveId(util.Deref(drive.GetId())).Items().ByDriveItemId(util.Deref(root.GetId())).Children().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Filter for Excel files
	for _, item := range items.GetValue() {
		name := util.Deref(item.GetName())
		if strings.HasSuffix(strings.ToLower(name), ".xlsx") {
			infos = append(infos, WorkbookInfo{
				ID:   util.Deref(item.GetId()),
				Name: name,
			})
		}
	}

	return infos, nil
}

func CreateWorksheet(ctx context.Context, c *msgraphsdkgo.GraphServiceClient, workbookID, name string) (string, error) {
	drive, err := c.Me().Drive().Get(ctx, nil)
	if err != nil {
		return "", err
	}

	requestBody := drives.NewItemItemsItemWorkbookWorksheetsAddPostRequestBody()
	requestBody.SetName(util.Ptr(name))
	worksheet, err := c.Drives().ByDriveId(util.Deref(drive.GetId())).Items().ByDriveItemId(workbookID).Workbook().Worksheets().Add().Post(ctx, requestBody, nil)
	if err != nil {
		return "", err
	}
	return util.Deref(worksheet.GetId()), nil
}

type WorksheetInfo struct {
	ID, Name, WorkbookID string
}

func ListWorksheetsInWorkbook(ctx context.Context, c *msgraphsdkgo.GraphServiceClient, workbookID string) ([]WorksheetInfo, error) {
	drive, err := c.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	sheets, err := c.Drives().ByDriveId(util.Deref(drive.GetId())).Items().ByDriveItemId(workbookID).Workbook().Worksheets().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	var infos []WorksheetInfo
	for _, sheet := range sheets.GetValue() {
		infos = append(infos, WorksheetInfo{
			ID:         util.Deref(sheet.GetId()),
			Name:       util.Deref(sheet.GetName()),
			WorkbookID: workbookID,
		})
	}
	return infos, nil
}

func GetWorksheetData(ctx context.Context, c *msgraphsdkgo.GraphServiceClient, workbookID, worksheetID string) ([][]any, models.WorkbookRangeable, error) {
	drive, err := c.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	usedRange, err := c.Drives().ByDriveId(util.Deref(drive.GetId())).Items().ByDriveItemId(workbookID).Workbook().Worksheets().ByWorkbookWorksheetId(worksheetID).UsedRange().Get(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	result, err := serialization.SerializeToJson(usedRange.GetValues())
	if err != nil {
		return nil, nil, err
	}

	var data [][]any
	if err = json.Unmarshal(result, &data); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}
	return data, usedRange, nil
}

func GetWorksheetColumnHeaders(ctx context.Context, c *msgraphsdkgo.GraphServiceClient, workbookID, worksheetID string) ([][]any, models.WorkbookRangeable, error) {
	drive, err := c.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	address := "1:3"
	usedRange, err := c.Drives().ByDriveId(util.Deref(drive.GetId())).Items().ByDriveItemId(workbookID).Workbook().Worksheets().ByWorkbookWorksheetId(worksheetID).RangeWithAddress(&address).UsedRange().Get(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	result, err := serialization.SerializeToJson(usedRange.GetValues())
	if err != nil {
		return nil, nil, err
	}

	var data [][]any
	if err = json.Unmarshal(result, &data); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}
	return data, usedRange, nil
}

func GetWorksheetTables(ctx context.Context, c *msgraphsdkgo.GraphServiceClient, workbookID, worksheetID string) ([]Table, error) {
	drive, err := c.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	result, err := c.Drives().ByDriveId(util.Deref(drive.GetId())).Items().ByDriveItemId(workbookID).Workbook().Worksheets().ByWorkbookWorksheetId(worksheetID).Tables().Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	var tables []Table
	for _, table := range result.GetValue() {
		tables = append(tables, Table{ID: util.Deref(table.GetId()), Name: util.Deref(table.GetName())})
	}

	return tables, nil
}

func AddWorksheetRow(ctx context.Context, c *msgraphsdkgo.GraphServiceClient, workbookID, worksheetID string, contents [][]string) error {
	_, usedRange, err := GetWorksheetData(ctx, c, workbookID, worksheetID)
	if err != nil {
		return err
	}
	lastUsedRow, err := getLastUsedRow(util.Deref(usedRange.GetAddress()))
	if err != nil {
		return err
	}
	newRowNumber := lastUsedRow + 1
	newEndRowNumber := newRowNumber + len(contents) - 1

	maxColumn := 0
	for _, row := range contents {
		if maxColumn < len(row) {
			maxColumn = len(row)
		}
	}
	endColumnLetter := numberToColumnLetter(maxColumn)
	address := fmt.Sprintf("A%d:%s%d", newRowNumber, endColumnLetter, newEndRowNumber)

	// Update the worksheet.
	// Unfortunately, the SDK lacks a function to do what we need to do, so we need to make a raw HTTP request.
	var values, formulas [][]any
	for _, row := range contents {
		var rowValues, rowFormulas []any
		for _, cell := range row {
			if strings.HasPrefix(cell, "=") {
				rowFormulas = append(rowFormulas, cell)
				rowValues = append(rowValues, nil)
			} else {
				rowValues = append(rowValues, cell)
				rowFormulas = append(rowFormulas, nil)
			}
		}
		values = append(values, rowValues)
		formulas = append(formulas, rowFormulas)
	}
	body := &UpdateBody{
		Values:   values,
		Formulas: formulas,
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/items/%s/workbook/worksheets/%s/range(address='%s')", workbookID, worksheetID, address), strings.NewReader(string(bodyJSON)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv(global.CredentialEnv))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("failed to close response body: %s", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			fmt.Println(string(body))
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func AddWorksheetColumn(ctx context.Context, workbookID, worksheetID, columnID string, contents []string) error {
	re := regexp.MustCompile(`\d+$`)
	startRow, err := strconv.Atoi(re.FindString(columnID))
	if err != nil {
		return fmt.Errorf("failed to parse starting row: %v", err)
	}
	endRow := startRow + len(contents) - 1

	endColumnID := re.ReplaceAllString(columnID, fmt.Sprintf("%d", endRow))
	address := fmt.Sprintf("%s:%s", columnID, endColumnID)

	body := new(UpdateBody)
	for _, v := range contents {
		if strings.HasPrefix(v, "=") {
			body.AppendColumnCellToFormulas([]any{v})
		} else {
			body.AppendColumnCellToValues([]any{v})
		}
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/items/%s/workbook/worksheets/%s/range(address='%s')", workbookID, worksheetID, address), strings.NewReader(string(bodyJSON)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv(global.CredentialEnv))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("failed to close response body: %s", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			fmt.Println(string(body))
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func getLastUsedRow(address string) (int, error) {
	parts := strings.Split(address, "!")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid address format")
	}

	rangePart := parts[1] // Extract "A2:C3"
	cellRanges := strings.Split(rangePart, ":")
	if len(cellRanges) > 2 {
		return 0, fmt.Errorf("invalid range format: %s", rangePart)
	}

	var endCell string
	if len(cellRanges) == 1 {
		return 0, nil
	} else {
		endCell = cellRanges[1] // "C3"
	}
	var rowNumStr string
	for _, ch := range endCell {
		if ch >= '0' && ch <= '9' { // Extract numeric part
			rowNumStr += string(ch)
		}
	}

	lastRow, err := strconv.Atoi(rowNumStr)
	if err != nil {
		return 0, fmt.Errorf("invalid row number")
	}

	return lastRow, nil
}
