package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/microsoft/kiota-abstractions-go/serialization"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	util "github.com/obot-platform/tools/microsoft365/excel-mcp-go/utils"
)

var httpAddr = flag.String("http", ":3000", "HTTP address to listen on for streamable HTTP server")

// StaticTokenCredential implements azcore.TokenCredential
type StaticTokenCredential struct {
	token string
}

func (s StaticTokenCredential) GetToken(_ context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: s.token}, nil
}

type UpdateBody struct {
	Values   [][]any `json:"values,omitempty"`
	Formulas [][]any `json:"formulas,omitempty"`
}

// ExcelMCPServer wraps the Microsoft Graph client for Excel operations
type ExcelMCPServer struct {
	client *msgraphsdkgo.GraphServiceClient
	token  string
}

// NewExcelMCPServer creates a new Excel MCP server with the given token
func NewExcelMCPServer(token string) (*ExcelMCPServer, error) {
	credential := StaticTokenCredential{token: token}
	client, err := msgraphsdkgo.NewGraphServiceClientWithCredentials(credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Graph client: %w", err)
	}

	return &ExcelMCPServer{client: client, token: token}, nil
}

// WorkbookInfo represents basic workbook information
type WorkbookInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// WorksheetInfo represents basic worksheet information
type WorksheetInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListWorkbooksArgs represents arguments for listing workbooks
type ListWorkbooksArgs struct{}

// ListWorksheetsArgs represents arguments for listing worksheets
type ListWorksheetsArgs struct {
	WorkbookID string `json:"workbook_id" jsonschema:"ID of the workbook to list worksheets from"`
}

// GetWorksheetDataArgs represents arguments for getting worksheet data
type GetWorksheetDataArgs struct {
	WorkbookID  string `json:"workbook_id" jsonschema:"ID of the workbook to get worksheet data from"`
	WorksheetID string `json:"worksheet_id" jsonschema:"ID of the worksheet to get data from"`
}

// GetWorksheetColumnHeadersArgs represents arguments for getting worksheet column headers
type GetWorksheetColumnHeadersArgs struct {
	WorkbookID  string `json:"workbook_id" jsonschema:"ID of the workbook to get worksheet data from"`
	WorksheetID string `json:"worksheet_id" jsonschema:"ID of the worksheet to get data from"`
}

// GetWorksheetTablesArgs represents arguments for getting worksheet tables
type GetWorksheetTablesArgs struct {
	WorkbookID  string `json:"workbook_id" jsonschema:"ID of the workbook to get worksheet data from"`
	WorksheetID string `json:"worksheet_id" jsonschema:"ID of the worksheet to get data from"`
}

// QueryWorksheetDataArgs represents arguments for querying worksheet data
type QueryWorksheetDataArgs struct {
	WorkbookID  string  `json:"workbook_id" jsonschema:"ID of the workbook to get worksheet data from"`
	WorksheetID string  `json:"worksheet_id" jsonschema:"ID of the worksheet to get data from"`
	ShowColumns *string `json:"show_columns,omitempty" jsonschema:"a comma-delimited list of columns to show in the output (Optional, by default shows first 5 columns)"`
	Query       *string `json:"query,omitempty" jsonschema:"The sql-like query to run against the spreadsheet. Should be the format expected by the pandas query function (e.g. column1 == 'value1' and column2 > 10)"`
}

// AddWorksheetRowArgs represents arguments for adding worksheet rows
type AddWorksheetRowArgs struct {
	WorkbookID  string     `json:"workbook_id" jsonschema:"ID of the workbook to add row to"`
	WorksheetID string     `json:"worksheet_id" jsonschema:"ID of the worksheet to add row to"`
	Contents    [][]string `json:"contents" jsonschema:"The rows to add to the worksheet. Cells within a row must be pipe-separated (|)."`
}

// AddWorksheetColumnArgs represents arguments for adding worksheet columns
type AddWorksheetColumnArgs struct {
	WorkbookID  string   `json:"workbook_id" jsonschema:"ID of the workbook to add column to"`
	WorksheetID string   `json:"worksheet_id" jsonschema:"ID of the worksheet to add column to"`
	Contents    []string `json:"contents" jsonschema:"Values to add as a new column"`
	ColumnID    string   `json:"column_id" jsonschema:"The starting address of the column to put the values in (e.g. B1)"`
}

// CreateWorksheetArgs represents arguments for creating worksheets
type CreateWorksheetArgs struct {
	WorkbookID string `json:"workbook_id" jsonschema:"ID of the workbook to create worksheet in"`
	Name       string `json:"name" jsonschema:"Name of the new worksheet"`
}

// GetDatesFromSerialsArgs represents arguments for converting Excel serials to dates
type GetDatesFromSerialsArgs struct {
	Serials []int `json:"serials" jsonschema:"A list of excel serial numbers to transform into dates"`
}

// ListWorkbooks lists all Excel workbooks in the user's drive
func (e *ExcelMCPServer) ListWorkbooks(ctx context.Context, req *mcp.CallToolRequest, args ListWorkbooksArgs) (*mcp.CallToolResult, any, error) {
	// Get the user's drive
	drive, err := e.client.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user's drive: %w", err)
	}

	if drive.GetId() == nil {
		return nil, nil, fmt.Errorf("failed to retrieve user's drive ID")
	}

	// Get the root folder
	root, err := e.client.Drives().ByDriveId(*drive.GetId()).Root().Get(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get drive root: %w", err)
	}

	if root.GetId() == nil {
		return nil, nil, fmt.Errorf("failed to retrieve drive root ID")
	}

	// Get children of the root folder
	children, err := e.client.Drives().ByDriveId(*drive.GetId()).Items().ByDriveItemId(*root.GetId()).Children().Get(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get drive items: %w", err)
	}

	var workbooks []WorkbookInfo
	if children.GetValue() != nil {
		for _, item := range children.GetValue() {
			if item.GetName() != nil && item.GetId() != nil &&
				strings.HasSuffix(strings.ToLower(*item.GetName()), ".xlsx") {
				workbooks = append(workbooks, WorkbookInfo{
					ID:   *item.GetId(),
					Name: *item.GetName(),
				})
			}
		}
	}

	result, err := json.MarshalIndent(workbooks, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal workbooks: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(result),
			},
		},
	}, nil, nil
}

// ListWorksheets lists all worksheets in a workbook
func (e *ExcelMCPServer) ListWorksheets(ctx context.Context, req *mcp.CallToolRequest, args ListWorksheetsArgs) (*mcp.CallToolResult, any, error) {
	// Get the user's drive to construct the workbook path
	drive, err := e.client.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user's drive: %w", err)
	}

	if drive.GetId() == nil {
		return nil, nil, fmt.Errorf("failed to retrieve user's drive ID")
	}

	// Get worksheets
	worksheets, err := e.client.Drives().ByDriveId(*drive.GetId()).Items().ByDriveItemId(args.WorkbookID).Workbook().Worksheets().Get(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get worksheets: %w", err)
	}

	var worksheetInfos []WorksheetInfo
	if worksheets.GetValue() != nil {
		for _, ws := range worksheets.GetValue() {
			if ws.GetId() != nil && ws.GetName() != nil {
				worksheetInfos = append(worksheetInfos, WorksheetInfo{
					ID:   *ws.GetId(),
					Name: *ws.GetName(),
				})
			}
		}
	}

	result, err := json.MarshalIndent(worksheetInfos, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal worksheets: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(result),
			},
		},
	}, nil, nil
}

// GetWorksheetData gets all data from a worksheet
func (e *ExcelMCPServer) GetWorksheetData(ctx context.Context, req *mcp.CallToolRequest, args GetWorksheetDataArgs) (*mcp.CallToolResult, any, error) {
	drive, err := e.client.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user's drive: %w", err)
	}

	if drive.GetId() == nil {
		return nil, nil, fmt.Errorf("failed to retrieve user's drive ID")
	}

	// Get used range of the worksheet
	usedRange, err := e.client.Drives().ByDriveId(*drive.GetId()).Items().ByDriveItemId(args.WorkbookID).
		Workbook().Worksheets().ByWorkbookWorksheetId(args.WorksheetID).UsedRange().Get(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get used range: %w", err)
	}

	result, err := serialization.SerializeToJson(usedRange.GetValues())
	if err != nil {
		return nil, nil, err
	}

	var data [][]any
	if err = json.Unmarshal(result, &data); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	csvData, err := convertToCSV(data)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: csvData,
			},
		},
	}, nil, nil
}

// GetWorksheetColumnHeaders gets the first 3 rows to determine column headers
func (e *ExcelMCPServer) GetWorksheetColumnHeaders(ctx context.Context, req *mcp.CallToolRequest, args GetWorksheetColumnHeadersArgs) (*mcp.CallToolResult, any, error) {
	drive, err := e.client.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user's drive: %w", err)
	}

	if drive.GetId() == nil {
		return nil, nil, fmt.Errorf("failed to retrieve user's drive ID")
	}

	// Get range A1:Z3 to capture headers
	address := "A1:Z3"
	rangeObj, err := e.client.Drives().ByDriveId(*drive.GetId()).Items().ByDriveItemId(args.WorkbookID).
		Workbook().Worksheets().ByWorkbookWorksheetId(args.WorksheetID).RangeWithAddress(&address).Get(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get header range: %w", err)
	}

	result, err := serialization.SerializeToJson(rangeObj.GetValues())
	if err != nil {
		return nil, nil, err
	}

	var data [][]any
	if err = json.Unmarshal(result, &data); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	csvData, err := convertToCSV(data)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: csvData,
			},
		},
	}, nil, nil
}

// GetWorksheetTables gets tables in a worksheet
func (e *ExcelMCPServer) GetWorksheetTables(ctx context.Context, req *mcp.CallToolRequest, args GetWorksheetTablesArgs) (*mcp.CallToolResult, any, error) {
	drive, err := e.client.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user's drive: %w", err)
	}

	if drive.GetId() == nil {
		return nil, nil, fmt.Errorf("failed to retrieve user's drive ID")
	}

	// Get tables
	tables, err := e.client.Drives().ByDriveId(*drive.GetId()).Items().ByDriveItemId(args.WorkbookID).
		Workbook().Worksheets().ByWorkbookWorksheetId(args.WorksheetID).Tables().Get(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get tables: %w", err)
	}

	var tableInfos []map[string]interface{}
	if tables.GetValue() != nil {
		for _, table := range tables.GetValue() {
			tableInfo := make(map[string]interface{})
			if table.GetId() != nil {
				tableInfo["id"] = *table.GetId()
			}
			if table.GetName() != nil {
				tableInfo["name"] = *table.GetName()
			}
			tableInfos = append(tableInfos, tableInfo)
		}
	}

	result, err := json.MarshalIndent(tableInfos, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal tables: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(result),
			},
		},
	}, nil, nil
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

// AddWorksheetRow adds rows to a worksheet
func (e *ExcelMCPServer) AddWorksheetRow(ctx context.Context, req *mcp.CallToolRequest, args AddWorksheetRowArgs) (*mcp.CallToolResult, any, error) {
	_, usedRange, err := GetWorksheetData(ctx, e.client, args.WorkbookID, args.WorksheetID)
	if err != nil {
		return nil, nil, err
	}
	lastUsedRow, err := getLastUsedRow(util.Deref(usedRange.GetAddress()))
	if err != nil {
		return nil, nil, err
	}
	newRowNumber := lastUsedRow + 1
	newEndRowNumber := newRowNumber + len(args.Contents) - 1

	maxColumn := 0
	for _, row := range args.Contents {
		if maxColumn < len(row) {
			maxColumn = len(row)
		}
	}
	endColumnLetter := numberToColumnLetter(maxColumn)
	address := fmt.Sprintf("A%d:%s%d", newRowNumber, endColumnLetter, newEndRowNumber)

	// Update the worksheet.
	// Unfortunately, the SDK lacks a function to do what we need to do, so we need to make a raw HTTP request.
	var values, formulas [][]any
	for _, row := range args.Contents {
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
		return nil, nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	newReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/items/%s/workbook/worksheets/%s/range(address='%s')", args.WorkbookID, args.WorksheetID, address), strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}
	newReq.Header.Set("Content-Type", "application/json")
	newReq.Header.Set("Authorization", "Bearer "+e.token)

	resp, err := http.DefaultClient.Do(newReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send request: %w", err)
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
		return nil, nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Added %d rows successfully", len(args.Contents)),
			},
		},
	}, nil, nil
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

// AddWorksheetColumn adds a column to a worksheet
func (e *ExcelMCPServer) AddWorksheetColumn(ctx context.Context, req *mcp.CallToolRequest, args AddWorksheetColumnArgs) (*mcp.CallToolResult, any, error) {
	re := regexp.MustCompile(`\d+$`)
	startRow, err := strconv.Atoi(re.FindString(args.ColumnID))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse starting row: %v", err)
	}
	endRow := startRow + len(args.Contents) - 1

	endColumnID := re.ReplaceAllString(args.ColumnID, fmt.Sprintf("%d", endRow))
	address := fmt.Sprintf("%s:%s", args.ColumnID, endColumnID)

	body := new(UpdateBody)
	for _, v := range args.Contents {
		if strings.HasPrefix(v, "=") {
			body.Formulas = append(body.Formulas, []any{v})
		} else {
			body.Values = append(body.Values, []any{v})
		}
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal body: %w", err)
	}

	newReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/items/%s/workbook/worksheets/%s/range(address='%s')", args.WorkbookID, args.WorksheetID, address), strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}
	newReq.Header.Set("Content-Type", "application/json")
	newReq.Header.Set("Authorization", "Bearer "+e.token)

	resp, err := http.DefaultClient.Do(newReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("failed to close response body: %s", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read response body: %w", err)
		}
		return nil, nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Added %d columns successfully", len(args.Contents)),
			},
		},
	}, nil, nil
}

// CreateWorksheet creates a new worksheet in a workbook
func (e *ExcelMCPServer) CreateWorksheet(ctx context.Context, req *mcp.CallToolRequest, args CreateWorksheetArgs) (*mcp.CallToolResult, any, error) {
	drive, err := e.client.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user's drive: %w", err)
	}

	if drive.GetId() == nil {
		return nil, nil, fmt.Errorf("failed to retrieve user's drive ID")
	}

	// Create new worksheet
	_, err = e.client.Drives().ByDriveId(*drive.GetId()).Items().ByDriveItemId(args.WorkbookID).
		Workbook().Worksheets().Post(ctx, nil, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create worksheet: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Worksheet \"%s\" created successfully", args.Name),
			},
		},
	}, nil, nil
}

// GetDatesFromSerials converts Excel serial numbers to dates
func (e *ExcelMCPServer) GetDatesFromSerials(ctx context.Context, req *mcp.CallToolRequest, args GetDatesFromSerialsArgs) (*mcp.CallToolResult, any, error) {
	var dates []string

	startDate := time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)
	for _, v := range args.Serials {
		// Excel treats the year 1900 as a leap year because of Lotus 1-2-3, and it starts with 1 representing January 1st, 1900.
		// To convert the Excel serial number to a date, we must therefore subtract 2 (days) from the serial number to account for those 2 days.
		// https://learn.microsoft.com/en-us/office/troubleshoot/excel/wrongly-assumes-1900-is-leap-year
		date := startDate.AddDate(0, 0, v-2).Format(time.DateOnly)
		dates = append(dates, date)
	}

	result := strings.Join(dates, "|")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: result,
			},
		},
	}, nil, nil
}

// ExtractTokenFromRequest extracts the bearer token from HTTP request headers
func extractTokenFromRequest(req *http.Request) (string, error) {
	// Try X-Forwarded-Access-Token first
	if token := req.Header.Get("X-Forwarded-Access-Token"); token != "" {
		return token, nil
	}

	return "", fmt.Errorf("no access token found in request headers")
}

func convertToCSV(data [][]any) (string, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	for _, row := range data {
		strRow := make([]string, len(row))
		for i, val := range row {
			strRow[i] = fmt.Sprintf("%v", val)
		}
		if err := writer.Write(strRow); err != nil {
			return "", fmt.Errorf("error writing row: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("error flushing writer: %w", err)
	}
	return builder.String(), nil
}

func main() {
	flag.Parse()

	// Create server factory that extracts token from each request
	serverFactory := func(req *http.Request) *mcp.Server {
		token, err := extractTokenFromRequest(req)
		if err != nil {
			log.Fatalf("Failed to extract token from request: %v", err)
		}

		excelServer, err := NewExcelMCPServer(token)
		if err != nil {
			log.Fatalf("Failed to create Excel MCP server: %v", err)
		}

		server := mcp.NewServer(&mcp.Implementation{Name: "excel-mcp-server"}, nil)

		// Create JSON schemas for the tools
		listWorksheetsSchema, _ := jsonschema.For[ListWorksheetsArgs](nil)
		getWorksheetDataSchema, _ := jsonschema.For[GetWorksheetDataArgs](nil)
		getWorksheetColumnHeadersSchema, _ := jsonschema.For[GetWorksheetColumnHeadersArgs](nil)
		getWorksheetTablesSchema, _ := jsonschema.For[GetWorksheetTablesArgs](nil)
		addWorksheetRowSchema, _ := jsonschema.For[AddWorksheetRowArgs](nil)
		addWorksheetColumnSchema, _ := jsonschema.For[AddWorksheetColumnArgs](nil)
		createWorksheetSchema, _ := jsonschema.For[CreateWorksheetArgs](nil)
		getDatesFromSerialsSchema, _ := jsonschema.For[GetDatesFromSerialsArgs](nil)

		// Register all tools with proper schemas
		mcp.AddTool(server, &mcp.Tool{
			Name:        "list_workbooks",
			Description: "Lists all workbooks available to the user.",
		}, excelServer.ListWorkbooks)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "list_worksheets",
			Description: "Lists all worksheets available in a workbook.",
			InputSchema: listWorksheetsSchema,
		}, excelServer.ListWorksheets)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "get_worksheet_data",
			Description: "Get all the data of a worksheet in a workbook.",
			InputSchema: getWorksheetDataSchema,
		}, excelServer.GetWorksheetData)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "get_worksheet_column_headers",
			Description: "Get the first 3 rows of a worksheet in a workbook to determine if there are column headers.",
			InputSchema: getWorksheetColumnHeadersSchema,
		}, excelServer.GetWorksheetColumnHeaders)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "get_worksheet_tables",
			Description: "Get the names and IDs of the tables on a worksheet in a workbook.",
			InputSchema: getWorksheetTablesSchema,
		}, excelServer.GetWorksheetTables)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "add_worksheet_row",
			Description: "Adds rows to an existing worksheet in a workbook.",
			InputSchema: addWorksheetRowSchema,
		}, excelServer.AddWorksheetRow)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "add_worksheet_column",
			Description: "Adds a column to an existing worksheet in a workbook.",
			InputSchema: addWorksheetColumnSchema,
		}, excelServer.AddWorksheetColumn)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "create_worksheet",
			Description: "Creates a new worksheet in a workbook.",
			InputSchema: createWorksheetSchema,
		}, excelServer.CreateWorksheet)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "get_dates_from_serials",
			Description: "Gets the date in 'YYYY-MM-DD' format from Excel serial numbers",
			InputSchema: getDatesFromSerialsSchema,
		}, excelServer.GetDatesFromSerials)

		return server
	}

	if *httpAddr != "" {
		handler := mcp.NewStreamableHTTPHandler(serverFactory, nil)
		log.Printf("Excel MCP server listening at %s", *httpAddr)
		if err := http.ListenAndServe(*httpAddr, handler); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("HTTP address is required")
	}
}
