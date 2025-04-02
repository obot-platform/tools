package main

import (
	"encoding/json"
	"fmt"

	"github.com/gptscript-ai/go-gptscript"
)

var inputText = gptscript.GetEnv("WORKFLOW_INPUT", "")

const (
	webhookContext = `This workflow is being called from a webhook. The input is a JSON structure of the webhook payload and any
important headers.`
	emailContext = `This workflow is being called from an email receiver. The input is a JSON structure of the email message body and any
important email headers (from, to, subject, etc).`
	slackContext = `This workflow is being called on an slack event. The input is a JSON structure of incoming slack message.`
)

type workflowInput struct {
	Type string `json:"type"`
}

func main() {
	var structuredInput workflowInput
	if err := json.Unmarshal([]byte(inputText), &structuredInput); err == nil {
		var context string
		switch structuredInput.Type {
		case "email":
			context = emailContext
		case "webhook":
			context = webhookContext
		case "slack":
			context = slackContext
		}
		if context != "" {
			fmt.Printf("START WORKFLOW CONTEXT:\n%s\nEND START WORKFLOW CONTEXT\n\n", context)
		}
	}

	fmt.Printf("START WORKFLOW INPUT:\n%s\nEND WORKFLOW INPUT\n\n", inputText)

	fmt.Printf("START WORKFLOW INSTRUCTIONS:\n%s\nEND WORKFLOW INSTRUCTIONS\n\n", `You are running as part of a headless workflow. Do not ask the user for confirmation. If the given task fails, attempt to determine why.`)
}
