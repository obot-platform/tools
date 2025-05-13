package proxy

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/obot-platform/tools/openai-model-provider/api"
)

func DefaultRewriteModelsResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return nil
	}

	defer resp.Body.Close()

	var body io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		resp.Header.Del("Content-Encoding")
		body = gzReader
	}

	var models api.ModelsResponse
	if err := json.NewDecoder(body).Decode(&models); err != nil {
		return fmt.Errorf("failed to decode models response: %w", err)
	}

	for i, model := range models.Data {
		if model.Metadata == nil {
			model.Metadata = make(map[string]string)
		}
		switch {
		case strings.HasPrefix(model.ID, "gpt-"),
			strings.HasPrefix(model.ID, "ft:gpt-"),
			strings.HasPrefix(model.ID, "o1-"),
			strings.HasPrefix(model.ID, "ft:o1-"):
			model.Metadata["usage"] = "llm"
		case strings.HasPrefix(model.ID, "text-embedding-"),
			strings.HasPrefix(model.ID, "ft:text-embedding-"):
			model.Metadata["usage"] = "text-embedding"
		case strings.HasPrefix(model.ID, "dall-e"),
			strings.HasPrefix(model.ID, "ft:dall-e"):
			model.Metadata["usage"] = "image-generation"
		}
		models.Data[i] = model
	}

	b, err := json.Marshal(models)
	if err != nil {
		return fmt.Errorf("failed to marshal models response: %w", err)
	}

	resp.Body = io.NopCloser(bytes.NewReader(b))
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(b)))
	return nil
}

// RewriteAllModelsWithUsage returns a response modifier that marks all models with the specified usage
func RewriteAllModelsWithUsage(usage string, filters ...func(string) bool) func(*http.Response) error {
	return func(resp *http.Response) error {
		if resp.StatusCode != http.StatusOK {
			return nil
		}

		defer resp.Body.Close()

		var body io.Reader = resp.Body
		if resp.Header.Get("Content-Encoding") == "gzip" {
			gzReader, err := gzip.NewReader(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to create gzip reader: %w", err)
			}
			defer gzReader.Close()
			resp.Header.Del("Content-Encoding")
			body = gzReader
		}

		var models api.ModelsResponse
		if err := json.NewDecoder(body).Decode(&models); err != nil {
			return fmt.Errorf("failed to decode models response: %w", err)
		}

		for i, model := range models.Data {
			if model.Metadata == nil {
				model.Metadata = make(map[string]string)
			}
			if len(filters) == 0 {
				model.Metadata["usage"] = usage
			} else {
				for _, filter := range filters {
					if filter(model.ID) {
						model.Metadata["usage"] = usage
						break
					}
				}
			}
			models.Data[i] = model
		}

		b, err := json.Marshal(models)
		if err != nil {
			return fmt.Errorf("failed to marshal models response: %w", err)
		}

		resp.Body = io.NopCloser(bytes.NewReader(b))
		resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(b)))
		return nil
	}
}

// RewriteAllModelsWithUsageMap returns a response modifier that marks all models with the specified usage
func RewriteAllModelsWithUsageMap(usageMap map[string][]func(string) bool) func(*http.Response) error {
	return func(resp *http.Response) error {
		if resp.StatusCode != http.StatusOK {
			return nil
		}

		defer resp.Body.Close()

		var body io.Reader = resp.Body
		if resp.Header.Get("Content-Encoding") == "gzip" {
			gzReader, err := gzip.NewReader(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to create gzip reader: %w", err)
			}
			defer gzReader.Close()
			resp.Header.Del("Content-Encoding")
			body = gzReader
		}

		var models api.ModelsResponse
		if err := json.NewDecoder(body).Decode(&models); err != nil {
			return fmt.Errorf("failed to decode models response: %w", err)
		}

		for usage, filters := range usageMap {
			for i, model := range models.Data {
				if model.Metadata == nil {
					model.Metadata = make(map[string]string)
				} else if model.Metadata["usage"] != "" {
					continue
				}

				if len(filters) == 0 {
					model.Metadata["usage"] = usage
				} else {
					for _, filter := range filters {
						if filter(model.ID) {
							model.Metadata["usage"] = usage
							break
						}
					}
				}
				models.Data[i] = model
			}
		}

		b, err := json.Marshal(models)
		if err != nil {
			return fmt.Errorf("failed to marshal models response: %w", err)
		}

		resp.Body = io.NopCloser(bytes.NewReader(b))
		resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(b)))
		return nil
	}
}
