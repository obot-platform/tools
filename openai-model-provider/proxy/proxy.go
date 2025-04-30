package proxy

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

var (
	OpenaiBaseHostName = "api.openai.com"

	ChatCompletionsPath = "/v1/chat/completions"
)

type Config struct {
	URL *url.URL

	// ListenPort is the port the proxy server listens on
	ListenPort string

	// Name is the name of the provider, used for logging
	Name string

	// BaseURL is the upstream model API URL, e.g. "https://api.openai.com/v1" - MUST include the basePath, if any (e.g. /v1)
	BaseURL string

	// APIKey will be used for Bearer Token Auth against the upstream API
	APIKey string

	// PersonalAPIKeyHeader is used to pull personal API keys that will be used when forwarding the request to the model provider
	// Should start with `X-Obot-`
	PersonalAPIKeyHeader string

	// PersonalBaseURLHeader is used to pull a personal base URL from the headers of a request to be used when forwarding the request to the model provider.
	PersonalBaseURLHeader string

	// ValidateFn is a function that can be used to validate the configuration
	ValidateFn func(cfg *Config) error

	// RewriteModelsFn is a function that can be used to rewrite the response from the upstream API on the /models endpoint
	RewriteModelsFn func(*http.Response) error

	// RewriteHeaderFn is a function that can be used to rewrite the request header
	RewriteHeaderFn func(header http.Header)

	// CustomPathHandleFuncs is a map of paths to custom handle funcs to completely override the default reverse proxy behavior for a given path
	CustomPathHandleFuncs map[string]http.HandlerFunc
}

type server struct {
	cfg *Config
}

func (cfg *Config) EnsureURL() error {
	if cfg.URL != nil {
		return nil
	}

	// Remove any trailing slashes from BaseURL
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")

	u, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return fmt.Errorf("failed to parse BaseURL: %w", err)
	}

	if u.Scheme == "" {
		u.Scheme = "https"
		if u.Host == "127.0.0.1" || u.Host == "localhost" {
			u.Scheme = "http"
		}
	}

	cfg.URL = u
	return nil
}

func Run(cfg *Config) error {
	if err := cfg.EnsureURL(); err != nil {
		return fmt.Errorf("failed to ensure URL: %w", err)
	}

	if cfg.RewriteModelsFn == nil {
		cfg.RewriteModelsFn = DefaultRewriteModelsResponse
	}

	if cfg.ValidateFn != nil {
		if err := cfg.ValidateFn(cfg); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	s := &server{cfg: cfg}

	mux := http.NewServeMux()

	if cfg.CustomPathHandleFuncs == nil {
		cfg.CustomPathHandleFuncs = make(map[string]http.HandlerFunc)
	}

	// Register default handlers only if they are not already registered
	if handler := cfg.CustomPathHandleFuncs["/{$}"]; handler == nil {
		cfg.CustomPathHandleFuncs["/{$}"] = s.healthz
	}
	if handler := cfg.CustomPathHandleFuncs["/v1/models"]; handler == nil {
		cfg.CustomPathHandleFuncs["/v1/models"] = (&httputil.ReverseProxy{
			Director:       s.proxyDirector,
			ModifyResponse: cfg.RewriteModelsFn,
		}).ServeHTTP
	}
	if handler := cfg.CustomPathHandleFuncs["/v1/"]; handler == nil {
		cfg.CustomPathHandleFuncs["/v1/"] = (&httputil.ReverseProxy{
			Director: s.proxyDirector,
		}).ServeHTTP
	}

	for path, handler := range cfg.CustomPathHandleFuncs {
		mux.HandleFunc(path, handler)
	}

	httpServer := &http.Server{
		Addr:    "127.0.0.1:" + cfg.ListenPort,
		Handler: mux,
	}

	fmt.Printf("[model-provider: %s] Starting OpenAI-style API proxy on port %s → baseURL=%s\n", cfg.Name, cfg.ListenPort, cfg.BaseURL)
	if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *server) healthz(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("http://127.0.0.1:" + s.cfg.ListenPort))
}

func (s *server) proxyDirector(req *http.Request) {
	u := s.cfg.URL
	if baseURL := req.Header.Get(s.cfg.PersonalBaseURLHeader); baseURL != "" {
		baseU, err := url.Parse(baseURL)
		if err == nil {
			u = baseU
		}
	}
	req.URL.Scheme = u.Scheme
	req.URL.Host = u.Host
	req.URL.Path = u.JoinPath(strings.TrimPrefix(req.URL.Path, "/v1")).Path // join baseURL with request path - /v1 must be part of baseURL if it's needed
	req.Host = req.URL.Host

	apiKey := s.cfg.APIKey
	if requestAPIKey := req.Header.Get(s.cfg.PersonalAPIKeyHeader); requestAPIKey != "" {
		apiKey = requestAPIKey
		req.Header.Del(s.cfg.PersonalAPIKeyHeader)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	if s.cfg.RewriteHeaderFn != nil {
		s.cfg.RewriteHeaderFn(req.Header)
	}
}

func Validate(cfg *Config) error {
	if cfg.ValidateFn == nil {
		return nil
	}
	return cfg.ValidateFn(cfg)
}
