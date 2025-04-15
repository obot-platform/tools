package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/gptscript-ai/go-gptscript"
)

type input struct {
	PromptInfo     promptInfo `json:"promptInfo"`
	ValidationTool string     `json:"validationTool"`
}

type promptInfo struct {
	ToolDisplayName string            `json:"tool_display_name"`
	Sensitive       string            `json:"sensitive"`
	Message         string            `json:"message"`
	Fields          []inputField      `json:"fields"`
	Metadata        map[string]string `json:"metadata"`
}

type inputField struct {
	gptscript.Field
	Env string `json:"env"`
}

type sysPromptInput struct {
	Message   string            `json:"message,omitempty"`
	Fields    gptscript.Fields  `json:"fields,omitempty"`
	Sensitive string            `json:"sensitive,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

func main() {
	if err := mainErr(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

func mainErr() error {
	// Set up signal handler
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	in, err := getInput()
	if err != nil {
		return fmt.Errorf("input error: %w", err)
	}

	client, err := gptscript.NewGPTScript()
	if err != nil {
		return fmt.Errorf("error creating GPTScript client: %w", err)
	}
	defer client.Close()

	credentialValues, err := getCredentials(ctx, client, in)
	if err != nil {
		return fmt.Errorf("error getting credentials: %w", err)
	}

	envs := make(map[string]string, len(credentialValues))
	for _, field := range in.PromptInfo.Fields {
		envs[field.Env] = credentialValues[field.Name]
	}

	if err = validateCredential(ctx, client, in.ValidationTool, envs); err != nil {
		return err
	}

	b, err := json.Marshal(map[string]any{"env": envs})
	if err != nil {
		return fmt.Errorf("error marshalling envs: %w", err)
	}

	fmt.Println(string(b))
	return nil
}

func getCredentials(ctx context.Context, client *gptscript.GPTScript, in input) (map[string]string, error) {
	promptFields := make([]gptscript.Field, 0, len(in.PromptInfo.Fields))
	for _, field := range in.PromptInfo.Fields {
		promptFields = append(promptFields, field.Field)
	}

	sysPromptIn, err := json.Marshal(sysPromptInput{
		Message:   in.PromptInfo.Message,
		Fields:    promptFields,
		Metadata:  in.PromptInfo.Metadata,
		Sensitive: in.PromptInfo.Sensitive,
	})
	if err != nil {
		return nil, fmt.Errorf("error marshalling sys prompt input: %w", err)
	}

	run, err := client.Run(ctx, "sys.prompt", gptscript.Options{
		Input: string(sysPromptIn),
	})
	if err != nil {
		return nil, fmt.Errorf("error running GPTScript prompt: %w", err)
	}

	res, err := run.Text()
	if err != nil {
		return nil, fmt.Errorf("error getting GPTScript response: %w", err)
	}

	var out map[string]string
	return out, json.Unmarshal([]byte(res), &out)
}

func getInput() (input, error) {
	inputStr := os.Getenv("TOOL_CALL_BODY")
	if inputStr == "" {
		return input{}, fmt.Errorf("input not found in environment variable TOOL_CALL_BODY")
	}

	var in input
	if err := json.Unmarshal([]byte(inputStr), &in); err != nil {
		return input{}, fmt.Errorf("error parsing input JSON: %w", err)
	}

	if in.PromptInfo.ToolDisplayName == "" {
		in.PromptInfo.ToolDisplayName = "Generic Credential"
	}

	if in.PromptInfo.Message == "" {
		in.PromptInfo.Message = "Please enter the following credentials."
	}

	if in.PromptInfo.Metadata == nil {
		in.PromptInfo.Metadata = make(map[string]string)
	}
	in.PromptInfo.Metadata["toolContext"] = "credential"
	in.PromptInfo.Metadata["authType"] = "generic"
	in.PromptInfo.Metadata["toolDisplayName"] = in.PromptInfo.ToolDisplayName

	return in, nil
}

func validateCredential(ctx context.Context, client *gptscript.GPTScript, tool string, envMap map[string]string) error {
	if tool == "" {
		return nil
	}

	env := make([]string, 0, len(envMap))
	for k, v := range envMap {
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

	output, err := run.Text()
	if err != nil {
		errStr, _, _ := strings.Cut(err.Error(), ": exit status ")
		return errors.New(errStr)
	}

	var errResp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal([]byte(output), &errResp); err == nil && errResp.Error != "" {
		return errors.New(errResp.Error)
	}

	return nil
}
