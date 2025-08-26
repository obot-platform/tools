package config

import (
	"testing"
	"time"

	"github.com/obot-platform/tools/auth-providers-common/pkg/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComplete(t *testing.T) {
	type expected struct {
		completed *Options
		err       error
	}

	tests := []struct {
		name     string
		input    options
		expected expected
	}{
		{
			name: "valid configuration with all fields",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         "example.com,test.org",
				AuthTokenRefreshDuration: "2h",
				GitHubOrg:                ptr("myorg"),
				GitHubAllowUsers:         ptr("user1,user2"),
			},
			expected: expected{
				completed: &Options{
					options: options{
						ClientID:                 "client123",
						ClientSecret:             "secret456",
						ObotServerURL:            "https://example.com",
						AuthCookieSecret:         "cookiesecret",
						AuthEmailDomains:         "example.com,test.org",
						AuthTokenRefreshDuration: "2h",
						GitHubOrg:                ptr("myorg"),
						GitHubAllowUsers:         ptr("user1,user2"),
					},
					AuthEmailDomains:         []string{"example.com", "test.org"},
					AuthTokenRefreshDuration: 2 * time.Hour,
					GitHubOrg:                "myorg",
					GitHubAllowUsers:         []string{"user1", "user2"},
				},
			},
		},
		{
			name: "minimal valid configuration",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         "*",
				AuthTokenRefreshDuration: "1h",
			},
			expected: expected{
				completed: &Options{
					options: options{
						ClientID:                 "client123",
						ClientSecret:             "secret456",
						ObotServerURL:            "https://example.com",
						AuthCookieSecret:         "cookiesecret",
						AuthEmailDomains:         "*",
						AuthTokenRefreshDuration: "1h",
					},
					AuthEmailDomains:         []string{"*"},
					AuthTokenRefreshDuration: 1 * time.Hour,
				},
			},
		},
		{
			name: "empty email domains",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         "example.com,,test.org",
				AuthTokenRefreshDuration: "1h",
			},
			expected: expected{
				err: env.FieldValidationError{
					EnvVar:  "OBOT_AUTH_PROVIDER_EMAIL_DOMAINS",
					Message: "cannot contain empty email domains",
					Value:   "example.com,,test.org",
				},
			},
		},
		{
			name: "wildcard with other domains",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         "*,example.com",
				AuthTokenRefreshDuration: "1h",
			},
			expected: expected{
				err: env.FieldValidationError{
					EnvVar:  "OBOT_AUTH_PROVIDER_EMAIL_DOMAINS",
					Message: "cannot specify multiple email domains when * is provided",
					Value:   "*,example.com",
				},
			},
		},
		{
			name: "invalid email domain",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         "invalid..domain",
				AuthTokenRefreshDuration: "1h",
			},
			expected: expected{
				err: env.FieldValidationError{
					EnvVar:  "OBOT_AUTH_PROVIDER_EMAIL_DOMAINS",
					Message: "'invalid..domain' is not a valid email domain",
					Value:   "invalid..domain",
				},
			},
		},
		{
			name: "invalid duration",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         "*",
				AuthTokenRefreshDuration: "invalid",
			},
			expected: expected{
				err: env.FieldValidationError{
					EnvVar:  "OBOT_AUTH_PROVIDER_TOKEN_REFRESH_DURATION",
					Message: "must be a valid duration string and greater than 0",
					Value:   "invalid",
				},
			},
		},
		{
			name: "zero duration",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         "*",
				AuthTokenRefreshDuration: "0s",
			},
			expected: expected{
				err: env.FieldValidationError{
					EnvVar:  "OBOT_AUTH_PROVIDER_TOKEN_REFRESH_DURATION",
					Message: "must be a valid duration string and greater than 0",
					Value:   "0s",
				},
			},
		},
		{
			name: "negative duration",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         "*",
				AuthTokenRefreshDuration: "-1h",
			},
			expected: expected{
				err: env.FieldValidationError{
					EnvVar:  "OBOT_AUTH_PROVIDER_TOKEN_REFRESH_DURATION",
					Message: "must be a valid duration string and greater than 0",
					Value:   "-1h",
				},
			},
		},
		{
			name: "very short valid duration",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         "*",
				AuthTokenRefreshDuration: "1ns",
			},
			expected: expected{
				completed: &Options{
					options: options{
						ClientID:                 "client123",
						ClientSecret:             "secret456",
						ObotServerURL:            "https://example.com",
						AuthCookieSecret:         "cookiesecret",
						AuthEmailDomains:         "*",
						AuthTokenRefreshDuration: "1ns",
					},
					AuthEmailDomains:         []string{"*"},
					AuthTokenRefreshDuration: 1 * time.Nanosecond,
				},
			},
		},
		{
			name: "GitHub org too long",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         "*",
				AuthTokenRefreshDuration: "1h",
				GitHubOrg:                ptr("this-organization-name-is-way-too-long-to-be-valid"),
			},
			expected: expected{
				err: env.FieldValidationError{
					EnvVar:  "OBOT_GITHUB_AUTH_PROVIDER_ORG",
					Message: "must be 39 characters or less, got 50 characters",
					Value:   "this-organization-name-is-way-too-long-to-be-valid",
				},
			},
		},
		{
			name: "GitHub org invalid characters",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         "*",
				AuthTokenRefreshDuration: "1h",
				GitHubOrg:                ptr("invalid_org"),
			},
			expected: expected{
				err: env.FieldValidationError{
					EnvVar:  "OBOT_GITHUB_AUTH_PROVIDER_ORG",
					Message: "is not a valid GitHub organization login (must contain only alphanumeric characters and single hyphens, cannot start or end with hyphen)",
					Value:   "invalid_org",
				},
			},
		},
		{
			name: "GitHub org starts with hyphen",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         "*",
				AuthTokenRefreshDuration: "1h",
				GitHubOrg:                ptr("-invalid"),
			},
			expected: expected{
				err: env.FieldValidationError{
					EnvVar:  "OBOT_GITHUB_AUTH_PROVIDER_ORG",
					Message: "is not a valid GitHub organization login (must contain only alphanumeric characters and single hyphens, cannot start or end with hyphen)",
					Value:   "-invalid",
				},
			},
		},
		{
			name: "GitHub org ends with hyphen",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         "*",
				AuthTokenRefreshDuration: "1h",
				GitHubOrg:                ptr("invalid-"),
			},
			expected: expected{
				err: env.FieldValidationError{
					EnvVar:  "OBOT_GITHUB_AUTH_PROVIDER_ORG",
					Message: "is not a valid GitHub organization login (must contain only alphanumeric characters and single hyphens, cannot start or end with hyphen)",
					Value:   "invalid-",
				},
			},
		},
		{
			name: "valid GitHub org at max length",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         "*",
				AuthTokenRefreshDuration: "1h",
				GitHubOrg:                ptr("a23456789012345678901234567890123456789"), // exactly 39 chars
			},
			expected: expected{
				completed: &Options{
					options: options{
						ClientID:                 "client123",
						ClientSecret:             "secret456",
						ObotServerURL:            "https://example.com",
						AuthCookieSecret:         "cookiesecret",
						AuthEmailDomains:         "*",
						AuthTokenRefreshDuration: "1h",
						GitHubOrg:                ptr("a23456789012345678901234567890123456789"),
					},
					AuthEmailDomains:         []string{"*"},
					AuthTokenRefreshDuration: 1 * time.Hour,
					GitHubOrg:                "a23456789012345678901234567890123456789",
				},
			},
		},
		{
			name: "empty GitHub users in comma-separated list",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         "*",
				AuthTokenRefreshDuration: "1h",
				GitHubAllowUsers:         ptr("user1,,user2"),
			},
			expected: expected{
				err: env.FieldValidationError{
					EnvVar:  "OBOT_GITHUB_AUTH_PROVIDER_ALLOW_USERS",
					Message: "cannot contain empty users",
					Value:   "user1,,user2",
				},
			},
		},
		{
			name: "GitHub user too long",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         "*",
				AuthTokenRefreshDuration: "1h",
				GitHubAllowUsers:         ptr("this-username-is-way-too-long-to-be-valid"),
			},
			expected: expected{
				err: env.FieldValidationError{
					EnvVar:  "OBOT_GITHUB_AUTH_PROVIDER_ALLOW_USERS",
					Message: "user 'this-username-is-way-too-long-to-be-valid' must be 39 characters or less, got 41 characters",
					Value:   "this-username-is-way-too-long-to-be-valid",
				},
			},
		},
		{
			name: "GitHub user invalid characters",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         "*",
				AuthTokenRefreshDuration: "1h",
				GitHubAllowUsers:         ptr("invalid_user"),
			},
			expected: expected{
				err: env.FieldValidationError{
					EnvVar:  "OBOT_GITHUB_AUTH_PROVIDER_ALLOW_USERS",
					Message: "user 'invalid_user' is not a valid GitHub username (must contain only alphanumeric characters and single hyphens, cannot start or end with hyphen)",
					Value:   "invalid_user",
				},
			},
		},
		{
			name: "valid GitHub user at max length",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         "*",
				AuthTokenRefreshDuration: "1h",
				GitHubAllowUsers:         ptr("u23456789012345678901234567890123456789"), // exactly 39 chars
			},
			expected: expected{
				completed: &Options{
					options: options{
						ClientID:                 "client123",
						ClientSecret:             "secret456",
						ObotServerURL:            "https://example.com",
						AuthCookieSecret:         "cookiesecret",
						AuthEmailDomains:         "*",
						AuthTokenRefreshDuration: "1h",
						GitHubAllowUsers:         ptr("u23456789012345678901234567890123456789"),
					},
					AuthEmailDomains:         []string{"*"},
					AuthTokenRefreshDuration: 1 * time.Hour,
					GitHubAllowUsers:         []string{"u23456789012345678901234567890123456789"},
				},
			},
		},
		{
			name: "whitespace trimming in domains and users",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         " example.com , test.org ",
				AuthTokenRefreshDuration: "1h",
				GitHubOrg:                ptr("myorg"),
				GitHubAllowUsers:         ptr(" user1 , user2 "),
			},
			expected: expected{
				completed: &Options{
					options: options{
						ClientID:                 "client123",
						ClientSecret:             "secret456",
						ObotServerURL:            "https://example.com",
						AuthCookieSecret:         "cookiesecret",
						AuthEmailDomains:         " example.com , test.org ",
						AuthTokenRefreshDuration: "1h",
						GitHubOrg:                ptr("myorg"),
						GitHubAllowUsers:         ptr(" user1 , user2 "),
					},
					AuthEmailDomains:         []string{"example.com", "test.org"},
					AuthTokenRefreshDuration: 1 * time.Hour,
					GitHubOrg:                "myorg",
					GitHubAllowUsers:         []string{"user1", "user2"},
				},
			},
		},
		{
			name: "zero values - nil pointers",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         "",
				AuthTokenRefreshDuration: "30m",
				GitHubOrg:                nil,
				GitHubAllowUsers:         nil,
			},
			expected: expected{
				completed: &Options{
					options: options{
						ClientID:                 "client123",
						ClientSecret:             "secret456",
						ObotServerURL:            "https://example.com",
						AuthCookieSecret:         "cookiesecret",
						AuthEmailDomains:         "",
						AuthTokenRefreshDuration: "30m",
					},
					AuthTokenRefreshDuration: 30 * time.Minute,
				},
			},
		},
		{
			name: "zero values - empty string pointers",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         "*",
				AuthTokenRefreshDuration: "45m",
				GitHubOrg:                ptr(""),
				GitHubAllowUsers:         ptr(""),
			},
			expected: expected{
				completed: &Options{
					options: options{
						ClientID:                 "client123",
						ClientSecret:             "secret456",
						ObotServerURL:            "https://example.com",
						AuthCookieSecret:         "cookiesecret",
						AuthEmailDomains:         "*",
						AuthTokenRefreshDuration: "45m",
						GitHubOrg:                ptr(""),
						GitHubAllowUsers:         ptr(""),
					},
					AuthEmailDomains:         []string{"*"},
					AuthTokenRefreshDuration: 45 * time.Minute,
				},
			},
		},
		{
			name: "complex domain validation with punycode",
			input: options{
				ClientID:                 "client123",
				ClientSecret:             "secret456",
				ObotServerURL:            "https://example.com",
				AuthCookieSecret:         "cookiesecret",
				AuthEmailDomains:         "xn--n3h.com", // valid punycode
				AuthTokenRefreshDuration: "1h",
			},
			expected: expected{
				completed: &Options{
					options: options{
						ClientID:                 "client123",
						ClientSecret:             "secret456",
						ObotServerURL:            "https://example.com",
						AuthCookieSecret:         "cookiesecret",
						AuthEmailDomains:         "xn--n3h.com",
						AuthTokenRefreshDuration: "1h",
					},
					AuthEmailDomains:         []string{"xn--n3h.com"},
					AuthTokenRefreshDuration: 1 * time.Hour,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := complete(tt.input)

			if tt.expected.err != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expected.err.Error())
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.completed, result)
			}
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}
