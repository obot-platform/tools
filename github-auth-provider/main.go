package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	oauth2proxy "github.com/oauth2-proxy/oauth2-proxy/v7"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/validation"
	"github.com/obot-platform/tools/auth-providers-common/pkg/env"
	"github.com/obot-platform/tools/auth-providers-common/pkg/state"
	"github.com/obot-platform/tools/github-auth-provider/pkg/profile"
	"github.com/sahilm/fuzzy"
)

type Options struct {
	ClientID                 string  `env:"OBOT_GITHUB_AUTH_PROVIDER_CLIENT_ID"`
	ClientSecret             string  `env:"OBOT_GITHUB_AUTH_PROVIDER_CLIENT_SECRET"`
	ObotServerURL            string  `env:"OBOT_SERVER_URL"`
	PostgresConnectionDSN    string  `env:"OBOT_AUTH_PROVIDER_POSTGRES_CONNECTION_DSN" optional:"true"`
	AuthCookieSecret         string  `usage:"Secret used to encrypt cookie" env:"OBOT_AUTH_PROVIDER_COOKIE_SECRET"`
	AuthEmailDomains         string  `usage:"Email domains allowed for authentication" default:"*" env:"OBOT_AUTH_PROVIDER_EMAIL_DOMAINS"`
	AuthTokenRefreshDuration string  `usage:"Duration to refresh auth token after" optional:"true" default:"1h" env:"OBOT_AUTH_PROVIDER_TOKEN_REFRESH_DURATION"`
	GitHubOrg                *string `usage:"restrict logins to members of this GitHub organization" optional:"true" env:"OBOT_GITHUB_AUTH_PROVIDER_ORG"`
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
	legacyOpts.LegacyProvider.Scope = "user:email read:org"
	legacyOpts.LegacyProvider.ClientID = opts.ClientID
	legacyOpts.LegacyProvider.ClientSecret = opts.ClientSecret

	// GitHub-specific options
	if opts.GitHubOrg != nil {
		legacyOpts.LegacyProvider.GitHubOrg = *opts.GitHubOrg
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
	mux.HandleFunc("/obot-get-user-info", func(w http.ResponseWriter, r *http.Request) {
		userInfo, err := profile.FetchUserProfile(r.Context(), r.Header.Get("Authorization"))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to fetch user info: %v", err), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(userInfo)
	})
	mux.HandleFunc("/obot-list-auth-groups", listGroups(opts.GitHubOrg))
	mux.HandleFunc("/obot-list-user-auth-groups", listUserGroups)
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
		// Make API requests to get more info about the authenticated user.
		ss.PreferredUsername = ss.User

		// Get user info
		userProfile, err := profile.FetchUserProfile(r.Context(), fmt.Sprintf("token %s", ss.AccessToken))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get user info: %v", err), http.StatusInternalServerError)
			fmt.Printf("ERROR: github-auth-provider: failed to get user info: %v\n", err)
			return
		}
		ss.User = fmt.Sprintf("%d", userProfile.ID)

		if err := json.NewEncoder(w).Encode(ss); err != nil {
			http.Error(w, fmt.Sprintf("failed to encode state: %v", err), http.StatusInternalServerError)
			fmt.Printf("ERROR: github-auth-provider: failed to encode state: %v\n", err)
			return
		}
	}
}

func listGroups(restrictOrg *string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "no authorization token provided", http.StatusUnauthorized)
			return
		}

		groups, err := profile.FetchUserGroupInfos(r.Context(), token)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to fetch user auth groups: %v", err), http.StatusInternalServerError)
			return
		}

		// Handle nil groups slice
		if groups == nil {
			groups = state.GroupInfoList{}
		}

		if restrictOrg != nil && *restrictOrg != "" {
			// Elide all org and team groups that don't match the restrictOrg when set
			groups = slices.DeleteFunc(groups, func(g state.GroupInfo) bool {
				orgLogin, _, _ := strings.Cut(g.Name, "/")
				return orgLogin != *restrictOrg
			})
		}

		// Get the name query parameter for filtering
		nameFilter := r.URL.Query().Get("name")
		if nameFilter != "" && len(groups) > 0 {
			// Create a slice of group names for fuzzy matching
			groupNames := make([]string, len(groups))
			for i, group := range groups {
				groupNames[i] = group.Name
			}

			// Perform fuzzy search - results are automatically ranked by relevance
			matches := fuzzy.Find(nameFilter, groupNames)

			// Filter groups based on fuzzy matches, preserving the relevance order
			var filteredGroups state.GroupInfoList
			for _, match := range matches {
				filteredGroups = append(filteredGroups, groups[match.Index])
			}
			groups = filteredGroups
		}

		if err := json.NewEncoder(w).Encode(groups); err != nil {
			http.Error(w, fmt.Sprintf("failed to encode groups: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func listUserGroups(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "no authorization token provided", http.StatusUnauthorized)
		return
	}

	groups, err := profile.FetchUserGroupInfos(r.Context(), token)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to fetch user auth groups: %v", err), http.StatusInternalServerError)
		return
	}

	// Handle nil groups slice
	if groups == nil {
		groups = state.GroupInfoList{}
	}

	// Note: Don't elide org and team groups because removing them here would cause Obot to drop
	// group membership for the respective user and would cause ACR's to garbage collect MCP servers and
	// server instances. In this case we want admins to manually clean up group based ACRs.
	if err := json.NewEncoder(w).Encode(groups); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode groups: %v", err), http.StatusInternalServerError)
		return
	}
}
