package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	oauth2proxy "github.com/oauth2-proxy/oauth2-proxy/v7"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/validation"
	"github.com/obot-platform/tools/auth-providers-common/pkg/env"
	"github.com/obot-platform/tools/auth-providers-common/pkg/icon"
	"github.com/obot-platform/tools/auth-providers-common/pkg/state"
	"github.com/obot-platform/tools/github-auth-provider/pkg/profile"
)

type Options struct {
	ClientID                 string  `env:"OBOT_GITHUB_AUTH_PROVIDER_CLIENT_ID"`
	ClientSecret             string  `env:"OBOT_GITHUB_AUTH_PROVIDER_CLIENT_SECRET"`
	ObotServerURL            string  `env:"OBOT_SERVER_URL"`
	PostgresConnectionDSN    string  `env:"OBOT_AUTH_PROVIDER_POSTGRES_CONNECTION_DSN" optional:"true"`
	AuthCookieSecret         string  `usage:"Secret used to encrypt cookie" env:"OBOT_AUTH_PROVIDER_COOKIE_SECRET"`
	AuthEmailDomains         string  `usage:"Email domains allowed for authentication" default:"*" env:"OBOT_AUTH_PROVIDER_EMAIL_DOMAINS"`
	AuthTokenRefreshDuration string  `usage:"Duration to refresh auth token after" optional:"true" default:"1h" env:"OBOT_AUTH_PROVIDER_TOKEN_REFRESH_DURATION"`
	GitHubTeams              *string `usage:"restrict logins to members of any of these GitHub teams (comma-separated list)" optional:"true" env:"OBOT_GITHUB_AUTH_PROVIDER_TEAMS"`
	GitHubOrg                *string `usage:"restrict logins to members of this GitHub organization" optional:"true" env:"OBOT_GITHUB_AUTH_PROVIDER_ORG"`
	GitHubRepo               *string `usage:"restrict logins to collaborators on this GitHub repository (formatted orgname/repo)" optional:"true" env:"OBOT_GITHUB_AUTH_PROVIDER_REPO"`
	GitHubToken              *string `usage:"the token to use when verifying repository collaborators (must have push access to the repository)" optional:"true" env:"OBOT_GITHUB_AUTH_PROVIDER_TOKEN"`
	GitHubAllowUsers         *string `usage:"users allowed to log in, even if they do not belong to the specified org and team or collaborators" optional:"true" env:"OBOT_GITHUB_AUTH_PROVIDER_ALLOW_USERS"`
}

func main() {
	var opts Options
	if err := env.LoadEnvForStruct(&opts); err != nil {
		fmt.Printf("ERROR: github-auth-provider: failed to load options: %v\n", err)
		os.Exit(1)
	}

	refreshDuration, err := time.ParseDuration(opts.AuthTokenRefreshDuration)
	if err != nil {
		fmt.Printf("ERROR: github-auth-provider: failed to parse token refresh duration: %v\n", err)
		os.Exit(1)
	}

	if refreshDuration < 0 {
		fmt.Printf("ERROR: github-auth-provider: token refresh duration must be greater than 0\n")
		os.Exit(1)
	}

	cookieSecret, err := base64.StdEncoding.DecodeString(opts.AuthCookieSecret)
	if err != nil {
		fmt.Printf("ERROR: github-auth-provider: failed to decode cookie secret: %v\n", err)
		os.Exit(1)
	}

	legacyOpts := options.NewLegacyOptions()
	legacyOpts.LegacyProvider.ProviderType = "github"
	legacyOpts.LegacyProvider.ProviderName = "github"
	legacyOpts.LegacyProvider.ClientID = opts.ClientID
	legacyOpts.LegacyProvider.ClientSecret = opts.ClientSecret

	// GitHub-specific options
	if opts.GitHubTeams != nil {
		legacyOpts.LegacyProvider.GitHubTeam = *opts.GitHubTeams
	}
	if opts.GitHubOrg != nil {
		legacyOpts.LegacyProvider.GitHubOrg = *opts.GitHubOrg
	}
	if opts.GitHubRepo != nil {
		legacyOpts.LegacyProvider.GitHubRepo = *opts.GitHubRepo
	}
	if opts.GitHubToken != nil {
		legacyOpts.LegacyProvider.GitHubToken = *opts.GitHubToken
	}
	if opts.GitHubAllowUsers != nil {
		legacyOpts.LegacyProvider.GitHubUsers = strings.Split(*opts.GitHubAllowUsers, ",")
	}

	oauthProxyOpts, err := legacyOpts.ToOptions()
	if err != nil {
		fmt.Printf("ERROR: github-auth-provider: failed to convert legacy options to new options: %v\n", err)
		os.Exit(1)
	}

	oauthProxyOpts.Server.BindAddress = ""
	oauthProxyOpts.MetricsServer.BindAddress = ""
	if opts.PostgresConnectionDSN != "" {
		oauthProxyOpts.Session.Type = options.PostgresSessionStoreType
		oauthProxyOpts.Session.Postgres.ConnectionDSN = opts.PostgresConnectionDSN
		oauthProxyOpts.Session.Postgres.TableNamePrefix = "github_"
	}
	oauthProxyOpts.Cookie.Refresh = refreshDuration
	oauthProxyOpts.Cookie.Name = "obot_access_token"
	oauthProxyOpts.Cookie.Secret = string(cookieSecret)
	oauthProxyOpts.Cookie.Secure = strings.HasPrefix(opts.ObotServerURL, "https://")
	oauthProxyOpts.Cookie.CSRFExpire = 30 * time.Minute
	oauthProxyOpts.Templates.Path = os.Getenv("GPTSCRIPT_TOOL_DIR") + "/../auth-providers-common/templates"
	oauthProxyOpts.RawRedirectURL = opts.ObotServerURL + "/"
	if opts.AuthEmailDomains != "" {
		emailDomains := strings.Split(opts.AuthEmailDomains, ",")
		for i := range emailDomains {
			emailDomains[i] = strings.TrimSpace(emailDomains[i])
		}
		oauthProxyOpts.EmailDomains = emailDomains
	}
	oauthProxyOpts.Logging.RequestEnabled = false
	oauthProxyOpts.Logging.AuthEnabled = false
	oauthProxyOpts.Logging.StandardEnabled = false

	if err = validation.Validate(oauthProxyOpts); err != nil {
		fmt.Printf("ERROR: github-auth-provider: failed to validate options: %v\n", err)
		os.Exit(1)
	}

	oauthProxy, err := oauth2proxy.NewOAuthProxy(oauthProxyOpts, oauth2proxy.NewValidator(oauthProxyOpts.EmailDomains, oauthProxyOpts.AuthenticatedEmailsFile))
	if err != nil {
		fmt.Printf("ERROR: github-auth-provider: failed to create oauth2 proxy: %v\n", err)
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "9999"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf("http://127.0.0.1:%s", port)))
	})
	mux.HandleFunc("/obot-get-state", getState(oauthProxy))
	mux.HandleFunc("/obot-get-icon-url", icon.ObotGetIconURL(profile.FetchGitHubProfileIconURL))
	mux.HandleFunc("/", oauthProxy.ServeHTTP)

	fmt.Printf("listening on 127.0.0.1:%s\n", port)
	if err := http.ListenAndServe("127.0.0.1:"+port, mux); !errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("ERROR: github-auth-provider: failed to listen and serve: %v\n", err)
		os.Exit(1)
	}
}

func getState(p *oauth2proxy.OAuthProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var sr state.SerializableRequest
		if err := json.NewDecoder(r.Body).Decode(&sr); err != nil {
			http.Error(w, fmt.Sprintf("failed to decode request body: %v", err), http.StatusBadRequest)
			return
		}

		reqObj, err := http.NewRequest(sr.Method, sr.URL, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create request object: %v", err), http.StatusBadRequest)
			return
		}

		reqObj.Header = sr.Header

		ss, err := state.GetSerializableState(p, reqObj)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get state: %v", err), http.StatusInternalServerError)
			fmt.Printf("ERROR: github-auth-provider: failed to get state: %v\n", err)
			return
		}

		// The User on the state, for GitHub, is the GitHub username.
		// This is bad, because we want the user ID instead.
		// Make an API request to get more info about the authenticated user.

		ss.PreferredUsername = ss.User

		var userID struct {
			ID int64 `json:"id"`
		}

		req, err := http.NewRequest("GET", "https://api.github.com/user", nil) // This gets the info for the authenticated user
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create request: %v", err), http.StatusInternalServerError)
			fmt.Printf("ERROR: github-auth-provider: failed to create request: %v\n", err)
			return
		}
		req.Header.Set("Authorization", fmt.Sprintf("token %s", ss.AccessToken))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to make request: %v", err), http.StatusInternalServerError)
			fmt.Printf("ERROR: github-auth-provider: failed to make request: %v\n", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			http.Error(w, fmt.Sprintf("failed to get user: %v", resp.StatusCode), http.StatusInternalServerError)
			fmt.Printf("ERROR: github-auth-provider: failed to get user: %v\n", resp.StatusCode)
			return
		}

		if err := json.NewDecoder(resp.Body).Decode(&userID); err != nil {
			http.Error(w, fmt.Sprintf("failed to decode user: %v", err), http.StatusInternalServerError)
			fmt.Printf("ERROR: github-auth-provider: failed to decode user: %v\n", err)
			return
		}

		ss.User = fmt.Sprintf("%d", userID.ID)

		if err := json.NewEncoder(w).Encode(ss); err != nil {
			http.Error(w, fmt.Sprintf("failed to encode state: %v", err), http.StatusInternalServerError)
			fmt.Printf("ERROR: github-auth-provider: failed to encode state: %v\n", err)
			return
		}
	}
}
