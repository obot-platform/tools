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
	apiKey           string
	baseURL          string
	failOnFailures   bool
	failOnSuspicious bool
	failOnTimeout    bool
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

	uploadURL, err := f.getUploadURL(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	analysisID, err := f.uploadFile(r.Context(), uploadURL, req.Contents)
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

func (f *fileHandler) getUploadURL(ctx context.Context) (string, error) {
	var resp urlResponse
	return resp.Data, doMultipartRequest(ctx, f.apiKey, f.baseURL+"/files/upload_url", http.MethodGet, nil, &resp)
}

func (f *fileHandler) uploadFile(ctx context.Context, url string, contents []byte) (string, error) {
	var resp fileUploadResponse
	return resp.Data.ID, doMultipartRequest(ctx, f.apiKey, url, http.MethodPost, map[string]io.Reader{
		"file": bytes.NewReader(contents),
	}, &resp)
}

func (f *fileHandler) pollForResult(ctx context.Context, analysisID string) (bool, error) {
	var (
		resp analysisResult
		err  error
	)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-ticker.C:
			if err = doMultipartRequest(ctx, f.apiKey, f.baseURL+"/analyses/"+analysisID, http.MethodGet, nil, &resp); err != nil {
				return false, err
			}

			if resp.Data.Attributes.Status == "completed" {
				if f.failOnFailures && resp.Data.Attributes.Stats.Failure > 0 {
					return false, &maliciousErr{msg: fmt.Sprintf("%d engines failed", resp.Data.Attributes.Stats.Failure)}
				}
				if f.failOnSuspicious && resp.Data.Attributes.Stats.Suspicious > 0 {
					return false, &maliciousErr{msg: fmt.Sprintf("%d engines found this file suspicious", resp.Data.Attributes.Stats.Suspicious)}
				}
				if f.failOnTimeout && resp.Data.Attributes.Stats.Timeout > 0 {
					return false, &maliciousErr{msg: fmt.Sprintf("%d engines timed out", resp.Data.Attributes.Stats.Timeout)}
				}
				return resp.Data.Attributes.Stats.Malicious == 0, nil
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

	req.Header.Set("x-apikey", apiKey)
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
		if err = json.Unmarshal(bodyText, &e); err != nil || e.Error.Message == "" {
			return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(bodyText))
		}
		return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, e.Error.Message)
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
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type urlResponse struct {
	Data string `json:"data"`
}

type fileUploadResponse struct {
	Data Data `json:"data"`
}

type analysisResult struct {
	Data Data `json:"data"`
}

type Data struct {
	Attributes Attributes `json:"attributes"`
	ID         string     `json:"id"`
	Type       string     `json:"type"`
}

type Attributes struct {
	Date    int                    `json:"date"`
	Results map[string]ResultEntry `json:"results"`
	Stats   Stats                  `json:"stats"`
	Status  string                 `json:"status"`
}

type ResultEntry struct {
	Category      string `json:"category"`
	EngineName    string `json:"engine_name"`
	EngineVersion string `json:"engine_version"`
	EngineUpdate  string `json:"engine_update"`
	Method        string `json:"method"`
	Result        string `json:"result"`
}

type Stats struct {
	ConfirmedTimeout int `json:"confirmed-timeout"`
	Failure          int `json:"failure"`
	Harmless         int `json:"harmless"`
	Malicious        int `json:"malicious"`
	Suspicious       int `json:"suspicious"`
	Timeout          int `json:"timeout"`
	TypeUnsupported  int `json:"type-unsupported"`
	Undetected       int `json:"undetected"`
}
