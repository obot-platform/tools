package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/obot-platform/obot/apiclient"
)

var (
	taskID   = strings.TrimSpace(os.Getenv("TOOL_CALL_BODY"))
	input    = os.Getenv("GPTSCRIPT_INPUT")
	url      = os.Getenv("OBOT_SERVER_URL") + "/api"
	token    = os.Getenv("OBOT_TOKEN")
	threadID = os.Getenv("OBOT_THREAD_ID")
)

func main() {
	if err := mainErr(); err != nil {
		_ = chatFinish(err.Error())
	}
}

func chatFinish(data string) error {
	cmd := exec.Command(os.ExpandEnv("${GPTSCRIPT_BIN}"), "sys.chat.finish", data)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func mainErr() error {
	if taskID == "" {
		return fmt.Errorf("task ID is empty")
	}

	data := map[string]any{}
	if err := json.Unmarshal([]byte(input), &data); err == nil {
		if s, _ := data["type"]; s == "obotExternalCallResume" {
			result, _ := data["result"].(map[string]any)
			ret, _ := result["data"].(string)
			return chatFinish(ret)
		}
		delete(data, "TOOL_CALL_BODY")
		delete(data, "TOOL_CALL_ARGS")
		if withoutToolCallKeys, err := json.Marshal(data); err == nil {
			input = string(withoutToolCallKeys)
		}
	}

	fmt.Fprintf(os.Stderr, "taskID: %s\n", taskID)
	fmt.Fprintf(os.Stderr, "threadID: %s\n", threadID)
	fmt.Fprintf(os.Stderr, "url: %s\n", url)

	client := &apiclient.Client{
		BaseURL: url,
		Token:   token,
	}

	run, err := client.RunTask(context.Background(), taskID, input, apiclient.TaskRunOptions{
		ThreadID: threadID,
	})
	if err != nil {
		return err
	}

	dataBytes, err := json.Marshal(map[string]any{
		"input":  input,
		"taskID": taskID,
	})
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{
		"id":   run.ID,
		"type": "obotExternalCall",
		"data": string(dataBytes),
	})
}
