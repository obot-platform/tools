package main

import (
	"context"
	"net/http"
)

func validate(ctx context.Context, apiKey string) error {
	return doMultipartRequest(ctx, apiKey, "https://api.metadefender.com/v4/apikey/", http.MethodGet, nil, nil)
}
