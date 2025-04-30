package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"

	aproxy "github.com/obot-platform/tools/anthropic-model-provider/proxy"
	"github.com/obot-platform/tools/openai-model-provider/proxy"
)

func main() {
	isValidate := len(os.Args) > 1 && os.Args[1] == "validate"

	cfg := &proxy.Config{
		APIKey:               os.Getenv("OBOT_ANTHROPIC_MODEL_PROVIDER_API_KEY"),
		PersonalAPIKeyHeader: "X-Obot-OBOT_ANTHROPIC_MODEL_PROVIDER_API_KEY",
		ListenPort:           os.Getenv("PORT"),
		BaseURL:              "https://api.anthropic.com/v1/",
		Name:                 "Anthropic",
	}

	prox := aproxy.NewServer(cfg)
	reverseProxy := &httputil.ReverseProxy{
		Director: prox.AnthropicProxyRedirect,
	}
	reverseProxyModels := &httputil.ReverseProxy{
		Director:       prox.AnthropicProxyRedirect,
		ModifyResponse: aproxy.RewriteModelsResponse,
	}
	cfg.CustomPathHandleFuncs = map[string]http.HandlerFunc{
		"/v1/models": reverseProxyModels.ServeHTTP,
		"/v1/":       reverseProxy.ServeHTTP,
	}

	if cfg.APIKey != "" {
		if err := cfg.Validate("/tools/anthropic-model-provider/validate"); err != nil {
			os.Exit(1)
		}
	}

	if isValidate {
		return
	}

	if err := proxy.Run(cfg); err != nil {
		fmt.Printf("failed to run anthropic-model-provider proxy: %v\n", err)
		os.Exit(1)
	}
}
