package proxy

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/obot-platform/tools/openai-model-provider/api"
)

var thinkingModels = []string{
	"claude-3-7-sonnet",
}

// RewriteModelsResponse returns a response modifier that marks all models with the specified usage
func RewriteModelsResponse(resp *http.Response) error {
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

	var extraModels []api.Model
	for i, model := range models.Data {
		if model.Metadata == nil {
			model.Metadata = make(map[string]string)
		}
		model.Metadata["usage"] = "llm"
		models.Data[i] = model

		if slices.ContainsFunc(thinkingModels, func(m string) bool {
			return strings.HasPrefix(model.ID, m)
		}) {
			extraModels = append(extraModels, api.Model{
				ID:       model.ID + "-thinking",
				Metadata: model.Metadata,
				Created:  model.Created,
				Object:   model.Object,
				OwnedBy:  model.OwnedBy,
			})
		}
	}

	models.Data = append(models.Data, extraModels...)

	b, err := json.Marshal(models)
	if err != nil {
		return fmt.Errorf("failed to marshal models response: %w", err)
	}

	resp.Body = io.NopCloser(bytes.NewReader(b))
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(b)))
	return nil
}
