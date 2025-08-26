package config

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/obot-platform/tools/auth-providers-common/pkg/env"
)

// options is the private struct that holds raw environment variable values
type options struct {
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

// Options is the public struct that holds validated and processed configuration values
type Options struct {
	options
	AuthEmailDomains         []string
	AuthTokenRefreshDuration time.Duration
	GitHubOrg                string
	GitHubAllowUsers         []string
}

var (
	gitHubLoginRegex = regexp.MustCompile(`^[a-zA-Z0-9]+(-[a-zA-Z0-9]+)*$`)
	emailDomainRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
)

// LoadEnv loads environment variables, validates them, and returns the completed configuration
func LoadEnv() (*Options, error) {
	completed, err := loadEnv()
	if err != nil {
		return nil, env.ValidationError{Err: err}
	}

	return completed, nil
}

func loadEnv() (*Options, error) {
	var opts options
	if err := env.LoadEnvForStruct(&opts); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	return complete(opts)
}

func complete(o options) (*Options, error) {
	var (
		validationErrors env.FieldValidationErrors
		completedOptions = Options{
			options: o,
		}
	)

	if o.AuthEmailDomains != "" {
		// TODO(njhale): Add validation for email domains
		var (
			emailDomains = strings.Split(o.AuthEmailDomains, ",")
			errorMsg     string
		)
		for i := range emailDomains {
			switch domain := strings.TrimSpace(emailDomains[i]); {
			case domain == "":
				errorMsg = "cannot contain empty email domains"
			case domain == "*" && len(emailDomains) > 1:
				errorMsg = "cannot specify multiple email domains when * is provided"
			case domain != "*" && !emailDomainRegex.MatchString(domain):
				errorMsg = fmt.Sprintf("'%s' is not a valid email domain", domain)
			default:
				emailDomains[i] = domain
				continue
			}

			// Stop after the first email domain validation error
			break
		}

		if errorMsg != "" {
			validationErrors = append(validationErrors, env.FieldValidationError{
				EnvVar:    "OBOT_AUTH_PROVIDER_EMAIL_DOMAINS",
				Message:   errorMsg,
				Value:     o.AuthEmailDomains,
				Sensitive: false,
			})
		} else {
			completedOptions.AuthEmailDomains = emailDomains
		}

	}

	refreshDuration, err := time.ParseDuration(o.AuthTokenRefreshDuration)
	if err != nil || refreshDuration <= 0 {
		validationErrors = append(validationErrors, env.FieldValidationError{
			EnvVar:    "OBOT_AUTH_PROVIDER_TOKEN_REFRESH_DURATION",
			Message:   "must be a valid duration string and greater than 0",
			Value:     o.AuthTokenRefreshDuration,
			Sensitive: false,
		})
	} else {
		completedOptions.AuthTokenRefreshDuration = refreshDuration
	}

	if o.GitHubOrg != nil && *o.GitHubOrg != "" {
		var (
			org      = *o.GitHubOrg
			errorMsg string
		)
		if len(org) > 39 {
			errorMsg = fmt.Sprintf("must be 39 characters or less, got %d characters", len(org))
		}
		if !gitHubLoginRegex.MatchString(org) {
			errorMsg = "is not a valid GitHub organization login (must contain only alphanumeric characters and single hyphens, cannot start or end with hyphen)"
		}
		if errorMsg != "" {
			validationErrors = append(validationErrors, env.FieldValidationError{
				EnvVar:    "OBOT_GITHUB_AUTH_PROVIDER_ORG",
				Message:   errorMsg,
				Value:     org,
				Sensitive: false,
			})
		} else {
			completedOptions.GitHubOrg = org
		}
	}

	if o.GitHubAllowUsers != nil && *o.GitHubAllowUsers != "" {
		var (
			users    = strings.Split(*o.GitHubAllowUsers, ",")
			errorMsg string
		)
		for i := range users {
			switch user := strings.TrimSpace(users[i]); {
			case user == "":
				errorMsg = "cannot contain empty users"
			case len(user) > 39:
				errorMsg = fmt.Sprintf("user '%s' must be 39 characters or less, got %d characters", user, len(user))
			case !gitHubLoginRegex.MatchString(user):
				errorMsg = fmt.Sprintf("user '%s' is not a valid GitHub username (must contain only alphanumeric characters and single hyphens, cannot start or end with hyphen)", user)
			default:
				users[i] = user
				continue
			}

			// Stop after the first user validation error
			break
		}

		if errorMsg != "" {
			validationErrors = append(validationErrors, env.FieldValidationError{
				EnvVar:    "OBOT_GITHUB_AUTH_PROVIDER_ALLOW_USERS",
				Message:   errorMsg,
				Value:     *o.GitHubAllowUsers,
				Sensitive: false,
			})
		} else {
			completedOptions.GitHubAllowUsers = users
		}
	}

	if len(validationErrors) > 0 {
		return nil, validationErrors
	}

	return &completedOptions, nil
}
