package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = w

	outCh := make(chan []byte, 1)
	go func() {
		buf, _ := io.ReadAll(r)
		outCh <- buf
	}()

	fn()

	_ = w.Close()
	os.Stdout = orig
	buf := <-outCh
	_ = r.Close()

	return string(buf)
}

func decodeErrorJSON(t *testing.T, out string) string {
	t.Helper()

	var body map[string]string
	if err := json.NewDecoder(bytes.NewBufferString(out)).Decode(&body); err != nil {
		t.Fatalf("decode stdout JSON: %v (raw: %q)", err, out)
	}
	return body["error"]
}

func captureLogOutput(t *testing.T, fn func()) string {
	t.Helper()

	origWriter := log.Writer()
	origFlags := log.Flags()
	origPrefix := log.Prefix()

	var b bytes.Buffer
	log.SetOutput(&b)
	log.SetFlags(0)
	log.SetPrefix("")

	defer func() {
		log.SetOutput(origWriter)
		log.SetFlags(origFlags)
		log.SetPrefix(origPrefix)
	}()

	fn()
	return b.String()
}

func TestValidate_ErrorResponses(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		responseBody string
		expectedErr  string
		expectedLogs []string
	}{
		{
			name:         "uses upstream invalid key message",
			statusCode:   http.StatusUnauthorized,
			responseBody: `{"error":{"message":"Incorrect API key provided: bad-key"}}`,
			expectedErr:  "Invalid API Key",
			expectedLogs: []string{
				"ERROR Invalid: status 401 Unauthorized",
				"Incorrect API key provided: bad-key",
			},
		},
		{
			name:         "falls back to status when body not parseable",
			statusCode:   http.StatusBadGateway,
			responseBody: "not-json",
			expectedErr:  "status: 502 Bad Gateway",
			expectedLogs: []string{
				"ERROR Invalid: status 502 Bad Gateway: not-json",
			},
		},
		{
			name:         "uses parseable non-auth upstream error message",
			statusCode:   http.StatusTooManyRequests,
			responseBody: `{"error":{"message":"rate limit exceeded"}}`,
			expectedErr:  "rate limit exceeded",
			expectedLogs: []string{
				"ERROR Invalid: status 429 Too Many Requests",
				"rate limit exceeded",
			},
		},
		{
			name:         "maps invalid x-api-key to invalid api key",
			statusCode:   http.StatusUnauthorized,
			responseBody: `{"error":{"message":"invalid x-api-key"}}`,
			expectedErr:  "Invalid API Key",
			expectedLogs: []string{
				"ERROR Invalid: status 401 Unauthorized",
				"invalid x-api-key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			t.Cleanup(ts.Close)

			cfg := &Config{BaseURL: ts.URL, APIKey: "test-key"}

			var err error
			var logs string
			out := captureStdout(t, func() {
				logs = captureLogOutput(t, func() {
					err = cfg.Validate()
				})
			})

			if err == nil {
				t.Fatal("expected validation error")
			}
			if got := err.Error(); got != tt.expectedErr {
				t.Fatalf("unexpected returned error: %q", got)
			}
			if got := decodeErrorJSON(t, out); got != tt.expectedErr {
				t.Fatalf("unexpected stdout error JSON: %q", got)
			}
			for _, expectedLog := range tt.expectedLogs {
				if !strings.Contains(logs, expectedLog) {
					t.Fatalf("expected log to contain %q, got: %q", expectedLog, logs)
				}
			}
		})
	}
}
