package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type fileHandler struct {
	apiKey        string
	baseURL       string
	failOnError   bool
	failOnUnknown bool
}

type fileScanRequest struct {
	Contents []byte `json:"contents"`
}

func (f *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req fileScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("failed to decode request body: %v", err), http.StatusBadRequest)
		return
	}

	result, err := f.checkHash(r.Context(), fmt.Sprintf("%X", req.Contents))
	if err == nil && result == 0 {
		// If no error occurred and the result is 0, the file is not malicious.
		return
	}

	analysisID, err := f.uploadFile(r.Context(), req.Contents)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	passed, err := f.pollForResult(r.Context(), analysisID)
	if err != nil {
		if m := (*maliciousErr)(nil); errors.As(err, &m) {
			http.Error(w, m.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !passed {
		http.Error(w, "file was flagged as malicious", http.StatusForbidden)
	}
}

func (f *fileHandler) checkHash(ctx context.Context, hexedHash string) (int, error) {
	var resp scanResultResponse
	return resp.ScanResults.ScanAllResultI, doMultipartRequest(ctx, f.apiKey, f.baseURL+"/hash/"+hexedHash, http.MethodGet, nil, &resp)
}

func (f *fileHandler) uploadFile(ctx context.Context, contents []byte) (string, error) {
	var resp fileUploadResponse
	return resp.DataID, doMultipartRequest(ctx, f.apiKey, f.baseURL+"/file", http.MethodPost, map[string]io.Reader{
		"file": bytes.NewReader(contents),
	}, &resp)
}

func (f *fileHandler) pollForResult(ctx context.Context, dataID string) (bool, error) {
	var (
		resp scanResultResponse
		err  error
	)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-ticker.C:
			if err = doMultipartRequest(ctx, f.apiKey, f.baseURL+"/file/"+dataID, http.MethodGet, nil, &resp); err != nil {
				return false, err
			}

			if resp.ScanResults.ProgressPercentage == 100 {
				if f.failOnError && resp.ScanResults.ScanAllResultI == 3 {
					return false, &maliciousErr{msg: "failed to scan file"}
				}
				if f.failOnUnknown && resp.ScanResults.ScanAllResultI == 2 {
					return false, &maliciousErr{msg: "an unknown error occurred while scanning the file"}
				}
				return resp.ScanResults.ScanAllResultI == 0, nil
			}
		}
	}
}

func doMultipartRequest(ctx context.Context, apiKey, url, method string, body map[string]io.Reader, responseBody any) error {
	var (
		buffer      bytes.Buffer
		contentType string
		part        io.Writer
		err         error
	)

	if len(body) != 0 {
		writer := multipart.NewWriter(&buffer)

		for key, reader := range body {
			if key == "file" {
				part, err = writer.CreateFormFile(key, "file")
				if err != nil {
					return fmt.Errorf("failed to create form file: %w", err)
				}
			} else {
				part, err = writer.CreateFormField(key)
				if err != nil {
					return fmt.Errorf("failed to create form field: %w", err)
				}
			}

			if _, err = io.Copy(part, reader); err != nil {
				return fmt.Errorf("failed to write form field: %w", err)
			}
		}

		if err = writer.Close(); err != nil {
			return fmt.Errorf("failed to close multipart writer: %w", err)
		}

		contentType = writer.FormDataContentType()
	}

	req, err := http.NewRequestWithContext(ctx, method, url, &buffer)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Accept", "application/json")

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var e apiError
		bodyText, _ := io.ReadAll(resp.Body)
		if err = json.Unmarshal(bodyText, &e); err != nil || len(e.Error.Messages) == 0 {
			return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(bodyText))
		}
		return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, e.Error.Messages[0])
	}

	if responseBody != nil {
		if err = json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
			return fmt.Errorf("failed to decode response body: %w", err)
		}
	}

	return nil
}

type maliciousErr struct {
	msg string
}

func (e *maliciousErr) Error() string {
	return e.msg
}

type apiError struct {
	Error struct {
		Code     int      `json:"code"`
		Messages []string `json:"messages"`
	} `json:"error"`
}

type scanResultResponse struct {
	ScanResultHistoryLength int         `json:"scan_result_history_length"`
	FileID                  string      `json:"file_id"`
	DataID                  string      `json:"data_id"`
	Sanitized               Sanitized   `json:"sanitized"`
	ProcessInfo             ProcessInfo `json:"process_info"`
	ScanResults             ScanResults `json:"scan_results"`
	FileInfo                FileInfo    `json:"file_info"`
	RestVersion             string      `json:"rest_version"`
	AdditionalInfo          []string    `json:"additional_info"`
	Votes                   Votes       `json:"votes"`
}

type Sanitized struct {
	Reason string `json:"reason"`
	Result string `json:"result"`
}

type ProcessInfo struct {
	Result              string         `json:"result"`
	Profile             string         `json:"profile"`
	PostProcessing      PostProcessing `json:"post_processing"`
	FileTypeSkippedScan bool           `json:"file_type_skipped_scan"`
	BlockedReason       string         `json:"blocked_reason"`
}

type PostProcessing struct {
	CopyMoveDestination  string `json:"copy_move_destination"`
	ConvertedTo          string `json:"converted_to"`
	ConvertedDestination string `json:"converted_destination"`
	ActionsRan           string `json:"actions_ran"`
	ActionsFailed        string `json:"actions_failed"`
}

type ScanResults struct {
	ScanDetails        map[string]ScanDetail `json:"scan_details"`
	ScanAllResultI     int                   `json:"scan_all_result_i"`
	StartTime          string                `json:"start_time"`
	TotalTime          int                   `json:"total_time"`
	TotalAvs           int                   `json:"total_avs"`
	TotalDetectedAvs   int                   `json:"total_detected_avs"`
	ProgressPercentage int                   `json:"progress_percentage"`
	ScanAllResultA     string                `json:"scan_all_result_a"`
}

type ScanDetail struct {
	ThreatFound string `json:"threat_found"`
	ScanTime    int    `json:"scan_time"`
	ScanResultI int    `json:"scan_result_i"`
	DefTime     string `json:"def_time"`
}

type FileInfo struct {
	FileSize            int    `json:"file_size"`
	UploadTimestamp     string `json:"upload_timestamp"`
	MD5                 string `json:"md5"`
	SHA1                string `json:"sha1"`
	SHA256              string `json:"sha256"`
	FileTypeCategory    string `json:"file_type_category"`
	FileTypeDescription string `json:"file_type_description"`
	FileTypeExtension   string `json:"file_type_extension"`
	DisplayName         string `json:"display_name"`
}

type Votes struct {
	Up   int `json:"up"`
	Down int `json:"down"`
}

type fileUploadResponse struct {
	DataID        string       `json:"data_id"`
	Status        string       `json:"status"`
	InQueue       int          `json:"in_queue"`
	QueuePriority string       `json:"queue_priority"`
	SHA1          string       `json:"sha1"`
	SHA256        string       `json:"sha256\""`
	SandboxError  SandboxError `json:"sandbox_error"`
}

type SandboxError struct {
	Code     int      `json:"code"`
	Messages []string `json:"messages"`
}
