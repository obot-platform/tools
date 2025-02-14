package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/obot-platform/tools/openai-model-provider/openaiproxy"
	"github.com/obot-platform/tools/openai-model-provider/proxy"
)

func main() {
	apiKey := os.Getenv("OBOT_OPENAI_MODEL_PROVIDER_API_KEY")
	if apiKey == "" {
		fmt.Println("OBOT_OPENAI_MODEL_PROVIDER_API_KEY environment variable not set")
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	cfg := &proxy.Config{
		APIKey:                apiKey,
		ListenPort:            port,
		BaseURL:               "https://api.openai.com/v1",
		RewriteModelsFn:       proxy.DefaultRewriteModelsResponse,
		Name:                  "OpenAI",
		CustomPathHandleFuncs: map[string]http.HandlerFunc{},
	}

	openaiProxy := openaiproxy.NewServer(cfg)
	reverseProxy := &httputil.ReverseProxy{
		Director:       openaiProxy.Openaiv1ProxyRedirect,
		ModifyResponse: openaiProxy.ModifyResponse,
	}
	cfg.CustomPathHandleFuncs["/v1/"] = reverseProxy.ServeHTTP

	if len(os.Args) > 1 && os.Args[1] == "validate" {
		if err := cfg.Validate("/tools/openai-model-provider/validate"); err != nil {
			os.Exit(1)
		}
		return
	}

	if err := proxy.Run(cfg); err != nil {
		panic(err)
	}
}
