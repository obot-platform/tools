package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
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

type authType string

var (
	obotAuthToken = gptscript.GetEnv("OBOT_TOKEN", "")
)

const (
	authTypePAT   authType = "Personal Access Token (PAT)"
	authTypeOAuth authType = "OAuth"
)

type input struct {
	OAuthInfo      oauthInfo   `json:"oauthInfo"`
	PromptInfo     *promptInfo `json:"promptInfo,omitempty"`
	ValidationTool string      `json:"validationTool,omitempty"`
}

type oauthInfo struct {
	Integration   string   `json:"integration"`
	Token         string   `json:"token"`
	Scope         []string `json:"scope"`
	UserScope     []string `json:"userScope"`
	OptionalScope []string `json:"optionalScope"`
}

type promptInfo struct {
	Fields    []field           `json:"fields"`
	Message   string            `json:"message"`
	Sensitive string            `json:"sensitive"`
	Metadata  map[string]string `json:"metadata"`
}

type field struct {
	gptscript.Field
	Env string `json:"env"`
}

type sysPromptInput struct {
	Message   string            `json:"message,omitempty"`
	Fields    gptscript.Fields  `json:"fields,omitempty"`
	Sensitive string            `json:"sensitive,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

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

type urls struct {
	authorizeURL string
	refreshURL   string
	tokenURL     string
}

func normalizeForEnv(appName string) string {
	return strings.ToUpper(strings.ReplaceAll(appName, "-", "_"))
}

func getURLs(appName string) urls {
	var (
		normalizedAppName = normalizeForEnv(appName)
		authorizeURL      = os.Getenv(fmt.Sprintf("GPTSCRIPT_OAUTH_%s_AUTH_URL", normalizedAppName))
		refreshURL        = os.Getenv(fmt.Sprintf("GPTSCRIPT_OAUTH_%s_REFRESH_URL", normalizedAppName))
		tokenURL          = os.Getenv(fmt.Sprintf("GPTSCRIPT_OAUTH_%s_TOKEN_URL", normalizedAppName))
	)
	return urls{
		authorizeURL: authorizeURL,
		refreshURL:   refreshURL,
		tokenURL:     tokenURL,
	}
}

func main() {
	if err := mainErr(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

func mainErr() (err error) {
	var credJSON []byte
	inputStr := os.Getenv("TOOL_CALL_BODY")
	if inputStr == "" {
		return fmt.Errorf("main: TOOL_CALL_BODY environment variable not set")
	}

	var in input
	if err = json.Unmarshal([]byte(inputStr), &in); err != nil {
		return fmt.Errorf("main: error parsing input JSON: %w", err)
	}

	gs, err := gptscript.NewGPTScript(gptscript.GlobalOptions{})
	if err != nil {
		return fmt.Errorf("main: failed to create GPTScript: %w", err)
	}
	defer gs.Close()

	ctx := context.Background()

	defer func() {
		if err == nil {
			if err = validateCredential(ctx, gs, in.ValidationTool, credJSON); err != nil {
				return
			}
			fmt.Print(string(credJSON))
		}
	}()

	urls := getURLs(in.OAuthInfo.Integration)

	if in.PromptInfo != nil && os.Getenv("GPTSCRIPT_EXISTING_CREDENTIAL") == "" {
		if urls.authorizeURL == "" && urls.refreshURL == "" && urls.tokenURL == "" {
			credJSON, err = promptForTokens(ctx, gs, in.OAuthInfo.Integration, in.PromptInfo)
			if err != nil {
				return fmt.Errorf("main: failed to prompt for tokens: %w", err)
			}
			return nil
		}

		authType, err := promptForSelect(ctx, gs)
		if err != nil {
			return fmt.Errorf("main: failed to prompt for auth type: %w", err)
		}

		if authType == authTypePAT {
			credJSON, err = promptForTokens(ctx, gs, in.OAuthInfo.Integration, in.PromptInfo)
			if err != nil {
				return fmt.Errorf("main: failed to prompt for tokens: %w", err)
			}
			return nil
		}
	}

	return promptForOauth(gs, &urls, &in, &credJSON)
}

func promptForOauth(gs *gptscript.GPTScript, urls *urls, in *input, credJSON *[]byte) error {
	// Refresh existing credential if there is one.
	existing := os.Getenv("GPTSCRIPT_EXISTING_CREDENTIAL")
	if existing != "" {
		var c cred
		if err := json.Unmarshal([]byte(existing), &c); err != nil {
			return fmt.Errorf("main: failed to unmarshal existing credential: %w", err)
		}

		u, err := url.Parse(urls.refreshURL)
		if err != nil {
			return fmt.Errorf("main: failed to parse refresh URL: %w", err)
		}

		q := u.Query()
		q.Set("refresh_token", c.RefreshToken)
		if len(in.OAuthInfo.Scope) != 0 {
			q.Set("scope", strings.Join(in.OAuthInfo.Scope, " "))
		}
		if len(in.OAuthInfo.OptionalScope) != 0 {
			q.Set("optional_scope", strings.Join(in.OAuthInfo.OptionalScope, " "))
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
			in.OAuthInfo.Token: oauthResp.AccessToken,
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

		*credJSON, err = json.Marshal(out)
		if err != nil {
			return fmt.Errorf("main: failed to marshal refreshed credential: %w", err)
		}

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

	u, err := url.Parse(urls.authorizeURL)
	if err != nil {
		return fmt.Errorf("main: failed to parse authorize URL: %w", err)
	}

	q := u.Query()
	q.Set("state", state)
	q.Set("challenge", challenge)
	if len(in.OAuthInfo.Scope) != 0 {
		q.Set("scope", strings.Join(in.OAuthInfo.Scope, " "))
	}
	if len(in.OAuthInfo.UserScope) != 0 {
		q.Set("user_scope", strings.Join(in.OAuthInfo.UserScope, " "))
	}
	if len(in.OAuthInfo.OptionalScope) != 0 {
		q.Set("optional_scope", strings.Join(in.OAuthInfo.OptionalScope, " "))
	}
	u.RawQuery = q.Encode()

	metadata := map[string]string{
		"authType":        "oauth",
		"toolContext":     "credential",
		"toolDisplayName": fmt.Sprintf("%s%s Integration", strings.ToTitle(in.OAuthInfo.Integration[:1]), in.OAuthInfo.Integration[1:]),
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
		oauthResp, retry, err := makeTokenRequest(urls.tokenURL, state, verifier)
		if err != nil {
			if !retry {
				return err
			}
			_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}

		envVars := map[string]string{
			in.OAuthInfo.Token: oauthResp.AccessToken,
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

		*credJSON, err = json.Marshal(out)
		if err != nil {
			return fmt.Errorf("main: failed to marshal token credential: %w", err)
		}

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
	if obotAuthToken != "" {
		req.Header.Add("Authorization", "Bearer "+obotAuthToken)
	}

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

func promptForSelect(ctx context.Context, g *gptscript.GPTScript) (authType, error) {
	fieldName := "Authentication Method"

	fields := gptscript.Fields{gptscript.Field{Name: fieldName, Description: "The authentication method to use for this tool.", Options: []string{string(authTypePAT), string(authTypeOAuth)}}}
	sysPromptIn, err := json.Marshal(sysPromptInput{
		Message: "This tool supports two ways to authenticate: personal access tokens (PAT) and OAuth. Select the method you'd like to use.",
		Fields:  fields,
	})
	if err != nil {
		return "", fmt.Errorf("promptForSelect: error marshalling sys prompt input: %w", err)
	}

	run, err := g.Run(ctx, "sys.prompt", gptscript.Options{
		Input: string(sysPromptIn),
	})
	if err != nil {
		return "", fmt.Errorf("promptForSelect: failed to run sys.prompt: %w", err)
	}

	out, err := run.Text()
	if err != nil {
		return "", fmt.Errorf("promptForSelect: failed to get prompt response: %w", err)
	}

	m := make(map[string]string)
	if err = json.Unmarshal([]byte(out), &m); err != nil {
		return "", fmt.Errorf("promptForSelect: failed to unmarshal prompt response: %w", err)
	}

	selectedAuthType := authType(m[fieldName])

	return selectedAuthType, nil
}

func promptForTokens(ctx context.Context, g *gptscript.GPTScript, integration string, prompt *promptInfo) ([]byte, error) {
	if prompt.Metadata == nil {
		prompt.Metadata = make(map[string]string)
	}
	prompt.Metadata["authType"] = "pat"
	prompt.Metadata["toolContext"] = "credential"
	prompt.Metadata["toolDisplayName"] = fmt.Sprintf("%s%s Integration", strings.ToTitle(integration[:1]), integration[1:])

	promptFields := make([]gptscript.Field, 0, len(prompt.Fields))
	for _, field := range prompt.Fields {
		promptFields = append(promptFields, field.Field)
	}

	sysPromptIn, err := json.Marshal(sysPromptInput{
		Message:   prompt.Message,
		Fields:    promptFields,
		Metadata:  prompt.Metadata,
		Sensitive: prompt.Sensitive,
	})
	if err != nil {
		return nil, fmt.Errorf("promptForTokens: error marshalling sys prompt input: %w", err)
	}

	run, err := g.Run(ctx, "sys.prompt", gptscript.Options{
		Input: string(sysPromptIn),
	})
	if err != nil {
		return nil, fmt.Errorf("promptForTokens: failed to run sys.prompt: %w", err)
	}

	out, err := run.Text()
	if err != nil {
		return nil, fmt.Errorf("promptForTokens: failed to get prompt response: %w", err)
	}

	m := make(map[string]string)
	if err = json.Unmarshal([]byte(out), &m); err != nil {
		return nil, fmt.Errorf("promptForTokens: failed to unmarshal prompt response: %w", err)
	}

	envs := make(map[string]string, len(m))
	for _, field := range prompt.Fields {
		envs[field.Env] = m[field.Name]
	}

	b, err := json.Marshal(map[string]any{"env": envs})
	if err != nil {
		return nil, fmt.Errorf("promptForTokens: error marshalling envs: %w", err)
	}

	return b, nil
}

func validateCredential(ctx context.Context, client *gptscript.GPTScript, tool string, envBytes []byte) error {
	if tool == "" {
		return nil
	}

	var envMap map[string]any
	if err := json.Unmarshal(envBytes, &envMap); err != nil {
		return err
	}

	env := make([]string, 0, len(envMap))
	for k, v := range envMap["env"].(map[string]any) {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	run, err := client.Run(ctx, tool, gptscript.Options{
		GlobalOptions: gptscript.GlobalOptions{
			Env: env,
		},
	})
	if err != nil {
		return fmt.Errorf("error running tool: %w", err)
	}

	_, err = run.Text()
	if err != nil {
		errStr, _, _ := strings.Cut(err.Error(), ": exit status ")
		return errors.New(errStr)
	}

	return nil
}
