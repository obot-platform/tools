module github.com/obot-platform/tools/vllm-model-provider

go 1.23.4

replace github.com/obot-platform/tools/openai-model-provider => ../openai-model-provider

require (
	github.com/gptscript-ai/chat-completion-client v0.0.0-20241127005108-02b41e1cd02e
	github.com/obot-platform/tools/openai-model-provider v0.0.0
)
