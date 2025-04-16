package main

import (
	"context"
	"net/http"
)

func validate(ctx context.Context, apiKey string) error {
	return doMultipartRequest(ctx, apiKey, "https://www.virustotal.com/api/v3/domains/www.google.com/votes", http.MethodGet, nil, nil)
}
