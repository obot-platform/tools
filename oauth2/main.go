package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/pkg/browser"
)

type oauthResponse struct {
	TokenType    string            `json:"token_type"`
	Scope        string            `json:"scope"`
	ExpiresIn    int               `json:"expires_in"`
	ExtExpiresIn int               `json:"ext_expires_in"`
	AccessToken  string            `json:"access_token"`
	RefreshToken string            `json:"refresh_token"`
	Extras       map[string]string `json:"extras"`
}

type cred struct {
	Env          map[string]string `json:"env"`
	ExpiresAt    *time.Time        `json:"expiresAt"`
	RefreshToken string            `json:"refreshToken"`
}

var (
	integration   = os.Getenv("INTEGRATION")
	token         = os.Getenv("TOKEN")
	scope         = os.Getenv("SCOPE")
	optionalScope = os.Getenv("OPTIONAL_SCOPE")
	promptTokens  = os.Getenv("PROMPT_TOKENS")
	promptVars    = os.Getenv("PROMPT_VARS")
)

func normalizeForEnv(appName string) string {
	return strings.ToUpper(strings.ReplaceAll(appName, "-", "_"))
}

func getURLs(appName string) (string, string, string) {
	var (
		normalizedAppName = normalizeForEnv(appName)
		authorizeURL      = os.Getenv(fmt.Sprintf("GPTSCRIPT_OAUTH_%s_AUTH_URL", normalizedAppName))
		refreshURL        = os.Getenv(fmt.Sprintf("GPTSCRIPT_OAUTH_%s_REFRESH_URL", normalizedAppName))
		tokenURL          = os.Getenv(fmt.Sprintf("GPTSCRIPT_OAUTH_%s_TOKEN_URL", normalizedAppName))
	)
	return authorizeURL, refreshURL, tokenURL
}

func main() {
	if err := mainErr(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

func mainErr() error {
	authorizeURL, refreshURL, tokenURL := getURLs(integration)
	if authorizeURL == "" || refreshURL == "" || tokenURL == "" {
		// The URLs aren't set for this credential. Check to see if we should prompt the user for other tokens
		if promptTokens == "" {
			fmt.Printf("All the following environment variables must be set: GPTSCRIPT_OAUTH_%s_AUTH_URL, GPTSCRIPT_OAUTH_%[1]s_REFRESH_URL, GPTSCRIPT_OAUTH_%[1]s_TOKEN_URL", normalizeForEnv(integration))
			fmt.Printf("Or the PROMPT_TOKENS environment variable must be provied for token prompting.")
			os.Exit(1)
		}

		return promptForTokens(integration, promptTokens, promptVars)
	}

	// Refresh existing credential if there is one.
	existing := os.Getenv("GPTSCRIPT_EXISTING_CREDENTIAL")
	if existing != "" {
		var c cred
		if err := json.Unmarshal([]byte(existing), &c); err != nil {
			return fmt.Errorf("main: failed to unmarshal existing credential: %w", err)
		}

		u, err := url.Parse(refreshURL)
		if err != nil {
			return fmt.Errorf("main: failed to parse refresh URL: %w", err)
		}

		q := u.Query()
		q.Set("refresh_token", c.RefreshToken)
		if scope != "" {
			q.Set("scope", strings.Join(strings.Fields(scope), " "))
		}
		if optionalScope != "" {
			q.Set("optional_scope", optionalScope)
		}
		u.RawQuery = q.Encode()

		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			return fmt.Errorf("main: failed to create refresh request: %w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("main: failed to send refresh request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("main: unexpected status code from refresh request: %d", resp.StatusCode)
		}

		var oauthResp oauthResponse
		if err := json.NewDecoder(resp.Body).Decode(&oauthResp); err != nil {
			return fmt.Errorf("main: failed to decode refresh response JSON: %w", err)
		}

		envVars := map[string]string{
			token: oauthResp.AccessToken,
		}

		for k, v := range oauthResp.Extras {
			envVars[k] = v
		}

		out := cred{
			Env:          envVars,
			RefreshToken: oauthResp.RefreshToken,
		}

		if oauthResp.ExpiresIn > 0 {
			expiresAt := time.Now().Add(time.Second * time.Duration(oauthResp.ExpiresIn))
			out.ExpiresAt = &expiresAt
		}

		credJSON, err := json.Marshal(out)
		if err != nil {
			return fmt.Errorf("main: failed to marshal refreshed credential: %w", err)
		}

		fmt.Print(string(credJSON))
		return nil
	}

	state, err := generateString()
	if err != nil {
		return fmt.Errorf("main: failed to generate state: %w", err)
	}

	verifier, err := generateString()
	if err != nil {
		return fmt.Errorf("main: failed to generate verifier: %w", err)
	}

	h := sha256.New()
	h.Write([]byte(verifier))
	challenge := hex.EncodeToString(h.Sum(nil))

	u, err := url.Parse(authorizeURL)
	if err != nil {
		return fmt.Errorf("main: failed to parse authorize URL: %w", err)
	}

	q := u.Query()
	q.Set("state", state)
	q.Set("challenge", challenge)
	if scope != "" {
		q.Set("scope", strings.Join(strings.Fields(scope), " "))
	}
	if optionalScope != "" {
		q.Set("optional_scope", optionalScope)
	}
	u.RawQuery = q.Encode()

	gs, err := gptscript.NewGPTScript(gptscript.GlobalOptions{})
	if err != nil {
		return fmt.Errorf("main: failed to create GPTScript: %w", err)
	}

	metadata := map[string]string{
		"authType":        "oauth",
		"toolContext":     "credential",
		"toolDisplayName": fmt.Sprintf("%s%s Integration", strings.ToTitle(integration[:1]), integration[1:]),
		"authURL":         u.String(),
	}

	b, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("main: failed to marshal metadata: %w", err)
	}

	run, err := gs.Run(context.Background(), "sys.prompt", gptscript.Options{
		Input: fmt.Sprintf(`{"metadata":%s,"message":%q}`, b, fmt.Sprintf("To authenticate please open your browser to %s.", u.String())),
	})
	if err != nil {
		return fmt.Errorf("main: failed to run sys.prompt: %w", err)
	}

	out, err := run.Text()
	if err != nil {
		return fmt.Errorf("main: failed to get text from sys.prompt: %w", err)
	}

	var m map[string]string
	_ = json.Unmarshal([]byte(out), &m)

	if m["handled"] != "true" {
		// Don't let the browser library print anything.
		browser.Stdout = io.Discard

		// Open the user's browser so that they can authorize the app.
		_ = browser.OpenURL(u.String())
	}

	t := time.NewTicker(2 * time.Second)
	for range t.C {
		now := time.Now()
		oauthResp, retry, err := makeTokenRequest(tokenURL, state, verifier)
		if err != nil {
			if !retry {
				return err
			}
			_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}

		envVars := map[string]string{
			token: oauthResp.AccessToken,
		}

		for k, v := range oauthResp.Extras {
			envVars[k] = v
		}

		out := cred{
			Env:          envVars,
			RefreshToken: oauthResp.RefreshToken,
		}

		if oauthResp.ExpiresIn > 0 {
			expiresAt := now.Add(time.Second * time.Duration(oauthResp.ExpiresIn))
			out.ExpiresAt = &expiresAt
		}

		credJSON, err := json.Marshal(out)
		if err != nil {
			return fmt.Errorf("main: failed to marshal token credential: %w", err)
		}

		fmt.Print(string(credJSON))
		break
	}

	return nil
}

func makeTokenRequest(tokenURL, state, verifier string) (*oauthResponse, bool, error) {
	// Construct the request to get the token from the gateway.
	req, err := http.NewRequest("GET", tokenURL, nil)
	if err != nil {
		return nil, false, fmt.Errorf("makeTokenRequest: failed to create token request: %w", err)
	}

	q := req.URL.Query()
	q.Set("state", state)
	q.Set("verifier", verifier)
	req.URL.RawQuery = q.Encode()

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, true, fmt.Errorf("makeTokenRequest: failed to send token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, true, fmt.Errorf("makeTokenRequest: unexpected status code from token request: %d", resp.StatusCode)
	}

	// Parse the response from the gateway.
	var oauthResp oauthResponse
	if err = json.NewDecoder(resp.Body).Decode(&oauthResp); err != nil {
		return nil, false, fmt.Errorf("makeTokenRequest: failed to decode token response JSON: %w", err)
	}

	return &oauthResp, false, nil
}

func generateString() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 256)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generateString: failed to read random bytes: %w", err)
	}

	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b), nil
}

func promptForTokens(integration, promptTokens, promptVars string) error {
	metadata := map[string]string{
		"authType":        "pat",
		"toolContext":     "credential",
		"toolDisplayName": fmt.Sprintf("%s%s Integration", strings.ToTitle(integration[:1]), integration[1:]),
	}

	b, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("promptForTokens: failed to marshal metadata: %w", err)
	}

	g, err := gptscript.NewGPTScript(gptscript.GlobalOptions{})
	if err != nil {
		return fmt.Errorf("promptForTokens: failed to create GPTScript client: %w", err)
	}
	defer g.Close()

	run, err := g.Run(context.Background(), "sys.prompt", gptscript.Options{
		Input: fmt.Sprintf(`{"metadata": %s,"message":"Please enter token values for %s.","fields":"%s","sensitive": "true"}`, b, integration, strings.ReplaceAll(promptTokens, ";", ",")),
	})
	if err != nil {
		return fmt.Errorf("promptForTokens: failed to run sys.prompt: %w", err)
	}

	out, err := run.Text()
	if err != nil {
		return fmt.Errorf("promptForTokens: failed to get prompt response: %w", err)
	}

	m := make(map[string]string)
	if err = json.Unmarshal([]byte(out), &m); err != nil {
		return fmt.Errorf("promptForTokens: failed to unmarshal prompt response: %w", err)
	}

	if promptVars != "" {
		run, err = g.Run(context.Background(), "sys.prompt", gptscript.Options{
			Input: fmt.Sprintf(`{"metadata": %s,"message":"Please enter token values for %s.","fields":"%s","sensitive": "false"}`, b, integration, strings.ReplaceAll(promptVars, ";", ",")),
		})
		if err != nil {
			return fmt.Errorf("promptForTokens: failed to run sys.prompt: %w", err)
		}

		out, err = run.Text()
		if err != nil {
			return fmt.Errorf("promptForTokens: failed to get prompt response: %w", err)
		}

		if err = json.Unmarshal([]byte(out), &m); err != nil {
			return fmt.Errorf("promptForTokens: failed to unmarshal prompt response: %w", err)
		}
	}

	b, err = json.Marshal(m)
	if err != nil {
		return fmt.Errorf("promptForTokens: failed to marshal prompt response: %w", err)
	}

	fmt.Printf(`{"env": %s}`, b)
	return nil
}
