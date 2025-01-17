package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	openai "github.com/gptscript-ai/chat-completion-client"
	"google.golang.org/genai"
)

const systemPrompt = `You are a task oriented system.
Be as brief as possible when answering the user.
Only give the required answer.
Do not give your thought process.
Use functions or tools as needed to complete the tasks given to you.
You are referred to as a tool.
Do not call functions or tools unless you need to.
Ensure you are passing the correct arguments to the functions or tools you call.
Do not move on to the next task until the current task is completed.
Do not make up arguments for tools.
Call functions one at a time to make sure you have the correct inputs.`

type server struct {
	port   string
	client *genai.Client
}

func Run(client *genai.Client, port string) error {
	mux := http.NewServeMux()

	s := &server{
		client: client,
		port:   port,
	}

	mux.HandleFunc("/{$}", s.healthz)
	mux.HandleFunc("GET /v1/models", s.listModels)
	mux.HandleFunc("POST /v1/chat/completions", s.chatCompletions)
	mux.HandleFunc("POST /v1/embeddings", s.embeddings)

	httpServer := &http.Server{
		Addr:    "127.0.0.1:" + port,
		Handler: mux,
	}

	if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *server) healthz(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("http://127.0.0.1:" + s.port))
}

func addUsageMetadata(models []map[string]any) []map[string]any {
	for _, m := range models {
		usage := "llm"
		if strings.Contains(m["id"].(string), "embedding") {
			usage = "text-embedding"
		}
		m["metadata"] = map[string]any{"usage": usage}
	}
	return models
}

func (s *server) listModels(w http.ResponseWriter, r *http.Request) {
	content := map[string]any{
		"data": addUsageMetadata(
			[]map[string]any{
				// LLMs: https://cloud.google.com/vertex-ai/generative-ai/docs/model-reference/inference#supported-models
				{
					"id":   "gemini-1.5-flash-001",
					"name": "Gemini 1.5 Flash (001)",
				},
				{
					"id":   "gemini-1.5-flash-002",
					"name": "Gemini 1.5 Flash (002)",
				},
				{
					"id":   "gemini-1.5-pro-001",
					"name": "Gemini 1.5 Pro (001)",
				},
				{
					"id":   "gemini-1.5-pro-002",
					"name": "Gemini 1.5 Pro (002)",
				},
				{
					"id":   "gemini-1.0-pro-vision-001",
					"name": "Gemini 1.0 Pro Vision (001)",
				},
				{
					"id":   "gemini-1.0-pro",
					"name": "Gemini 1.0 Pro",
				},
				{
					"id":   "gemini-1.0-pro-001",
					"name": "Gemini 1.0 Pro (001)",
				},
				{
					"id":   "gemini-1.0-pro-002",
					"name": "Gemini 1.0 Pro (002)",
				},
				// Embedding Models: https://cloud.google.com/vertex-ai/generative-ai/docs/embeddings/get-text-embeddings#supported-models
				{
					"id":   "textembedding-gecko@001",
					"name": "Text Embedding Gecko (001) [EN]",
				},
				{
					"id":   "textembedding-gecko@003",
					"name": "Text Embedding Gecko (003) [EN]",
				},
				{
					"id":   "text-embedding-004",
					"name": "Text Embedding 004 [EN]",
				},
				{
					"id":   "text-embedding-005",
					"name": "Text Embedding 005 [EN]",
				},
				{
					"id":   "textembedding-gecko-multilingual@001",
					"name": "Text Embedding Gecko Multilingual (001)",
				},
				{
					"id":   "text-multilingual-embedding-002",
					"name": "Text Multilingual Embedding 002",
				},
			},
		),
	}
	if err := json.NewEncoder(w).Encode(content); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *server) chatCompletions(w http.ResponseWriter, r *http.Request) {
	var cr openai.ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&cr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Tools
	tools, err := mapToolsFromOpenAI(cr.Tools)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Messages
	contents, err := mapMessagesFromOpenAI(cr.Messages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Temperature
	var temperature *float64
	if cr.Temperature != nil {
		t := float64(*cr.Temperature)
		temperature = &t
	}

	// TopP
	var topP *float64
	if cr.TopP > 0 {
		t := float64(cr.TopP)
		topP = &t
	}

	// MaxTokens
	var maxTokens *int64
	if cr.MaxTokens > 0 {
		m := int64(cr.MaxTokens)
		maxTokens = &m
	}

	// Options
	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{
					Text: systemPrompt,
				},
			},
			Role: "user",
		},
		Tools:           tools,
		Temperature:     temperature,
		TopP:            topP,
		MaxOutputTokens: maxTokens,
		CandidateCount:  int64(cr.N),
		SafetySettings: []*genai.SafetySetting{
			{
				Method:    genai.HarmBlockMethodSeverity,
				Category:  genai.HarmCategoryHateSpeech,
				Threshold: genai.HarmBlockThresholdBlockOnlyHigh,
			},
			{
				Method:    genai.HarmBlockMethodSeverity,
				Category:  genai.HarmCategoryDangerousContent,
				Threshold: genai.HarmBlockThresholdBlockOnlyHigh,
			},
			{
				Method:    genai.HarmBlockMethodSeverity,
				Category:  genai.HarmCategorySexuallyExplicit,
				Threshold: genai.HarmBlockThresholdBlockOnlyHigh,
			},
			{
				Method:    genai.HarmBlockMethodSeverity,
				Category:  genai.HarmCategoryHarassment,
				Threshold: genai.HarmBlockThresholdBlockOnlyHigh,
			},
		},
	}

	if cr.Stream {
		for result, err := range s.client.Models.GenerateContentStream(r.Context(), cr.Model, contents, config) {
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			choices, err := mapToOpenAIStreamChoice(result.Candidates)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			resp := openai.ChatCompletionStreamResponse{
				ID:      "0",
				Choices: choices,
				Created: 0,
				Model:   cr.Model,
				Object:  "chat.completion.chunk",
				Usage:   mapUsageToOpenAI(result.UsageMetadata),
			}

			if err := json.NewEncoder(w).Encode(resp); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	} else {
		result, err := s.client.Models.GenerateContent(r.Context(), cr.Model, contents, config)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		choices, err := mapToOpenAIChoice(result.Candidates)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := openai.ChatCompletionResponse{
			ID:      "0",
			Object:  "chat.completion",
			Created: 0,
			Model:   cr.Model,
			Choices: choices,
			Usage:   mapUsageToOpenAI(result.UsageMetadata),
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func mapUsageToOpenAI(usage *genai.GenerateContentResponseUsageMetadata) openai.Usage {
	if usage == nil {
		return openai.Usage{}
	}
	return openai.Usage{
		PromptTokens:     int(usage.PromptTokenCount),
		CompletionTokens: int(usage.CandidatesTokenCount),
		TotalTokens:      int(usage.TotalTokenCount),
	}
}

func mapToOpenAIContentAndToolCalls(parts []*genai.Part) (string, []openai.ToolCall, error) {
	var (
		toolCalls []openai.ToolCall
		content   string
	)
	for idx, p := range parts {
		if p.Text != "" {
			content += "\n" + p.Text
		}
		if p.FunctionCall != nil {
			args, err := json.Marshal(p.FunctionCall.Args)
			if err != nil {
				return "", nil, fmt.Errorf("failed to marshal function arguments: %w", err)
			}
			toolCalls = append(toolCalls, openai.ToolCall{
				Index: &idx,
				ID:    p.FunctionCall.ID,
				Type:  openai.ToolTypeFunction,
				Function: openai.FunctionCall{
					Name:      p.FunctionCall.Name,
					Arguments: string(args),
				},
			})
		}
	}
	return content, toolCalls, nil
}

func mapToOpenAIStreamChoice(candidates []*genai.Candidate) ([]openai.ChatCompletionStreamChoice, error) {
	var choices []openai.ChatCompletionStreamChoice
	for i, c := range candidates {
		content, toolCalls, err := mapToOpenAIContentAndToolCalls(c.Content.Parts)
		if err != nil {
			return nil, fmt.Errorf("failed to map content and tool calls: %w", err)
		}

		var finishReason openai.FinishReason
		if len(toolCalls) > 0 {
			finishReason = openai.FinishReasonFunctionCall
		} else {
			finishReason = mapFinishReasonToOpenAI(c.FinishReason)
		}

		choice := openai.ChatCompletionStreamChoice{
			Index: i,
			Delta: openai.ChatCompletionStreamChoiceDelta{
				Content:   strings.TrimSpace(content),
				ToolCalls: toolCalls,
				Role:      mapRoleToOpenAI(c.Content.Role),
			},
			FinishReason:         finishReason,
			ContentFilterResults: openai.ContentFilterResults{}, // TODO: fill based on Google's finish_reason?
		}
		choices = append(choices, choice)
	}

	return choices, nil
}

func mapToOpenAIChoice(candidates []*genai.Candidate) ([]openai.ChatCompletionChoice, error) {
	var choices []openai.ChatCompletionChoice
	for i, c := range candidates {
		content, toolCalls, err := mapToOpenAIContentAndToolCalls(c.Content.Parts)
		if err != nil {
			return nil, fmt.Errorf("failed to map content and tool calls: %w", err)
		}

		var finishReason openai.FinishReason
		if len(toolCalls) > 0 {
			finishReason = openai.FinishReasonFunctionCall
		} else {
			finishReason = mapFinishReasonToOpenAI(c.FinishReason)
		}

		choice := openai.ChatCompletionChoice{
			Index:        i,
			FinishReason: finishReason,
			Message: openai.ChatCompletionMessage{
				Role:      mapRoleToOpenAI(c.Content.Role),
				Content:   content,
				ToolCalls: toolCalls,
			},
			LogProbs: nil,
		}
		choices = append(choices, choice)
	}

	return choices, nil
}

func mapFinishReasonToOpenAI(reason genai.FinishReason) openai.FinishReason {
	switch reason {
	case genai.FinishReasonStop, genai.FinishReasonUnspecified, genai.FinishReasonOther:
		return openai.FinishReasonStop
	case genai.FinishReasonMaxTokens:
		return openai.FinishReasonLength
	case genai.FinishReasonBlocklist, genai.FinishReasonRecitation, genai.FinishReasonSafety, genai.FinishReasonSPII, genai.FinishReasonProhibitedContent:
		return openai.FinishReasonContentFilter
	default:
		return openai.FinishReasonStop
	}
}

var roleMapFromOpenAI = map[string]string{
	"system":    "user",
	"user":      "user",
	"assistant": "model",
	"model":     "model",
	"tool":      "function",
}

func mapRoleFromOpenAI(role string) string {
	if r, ok := roleMapFromOpenAI[role]; ok {
		return r
	}
	return "user"
}

var roleMapToOpenAI = map[string]string{
	"system":    "user",
	"user":      "user",
	"assistant": "model",
	"model":     "assistant",
	"function":  "tool",
}

func mapRoleToOpenAI(role string) string {
	if r, ok := roleMapToOpenAI[role]; ok {
		return r
	}
	return "user"
}

func mapMessagesFromOpenAI(messages []openai.ChatCompletionMessage) ([]*genai.Content, error) {
	var contents []*genai.Content
	if len(messages) > 0 {
		contents = append(contents, &genai.Content{
			Parts: []*genai.Part{
				{
					Text: systemPrompt,
				},
			},
			Role: "user",
		})
	}

	for _, m := range messages {
		content := &genai.Content{
			Parts: []*genai.Part{},
			Role:  mapRoleFromOpenAI(m.Role),
		}

		if m.ToolCallID != "" {
			// Tool Call Response
			content.Parts = append(content.Parts, &genai.Part{
				FunctionResponse: &genai.FunctionResponse{
					ID:   m.ToolCallID,
					Name: m.Name,
					Response: map[string]any{
						"name":    m.Name,
						"content": m.Content,
					},
				},
			})
		} else if len(m.ToolCalls) > 0 {
			// Tool Calls
			for _, tc := range m.ToolCalls {
				var args map[string]any
				if tc.Function.Arguments != "" {
					err := json.Unmarshal([]byte(tc.Function.Arguments), &args)
					if err != nil {
						return nil, fmt.Errorf("failed to unmarshal function arguments: %w", err)
					}
				}
				content.Parts = append(content.Parts, &genai.Part{
					FunctionCall: &genai.FunctionCall{
						ID:   tc.ID,
						Name: tc.Function.Name,
						Args: args,
					},
				})
			}
		} else if m.Content != "" {
			// Pure text content
			content.Parts = append(content.Parts, &genai.Part{
				Text: m.Content,
			})
		}

		contents = append(contents, content)
	}
	return contents, nil
}

func mapToolsFromOpenAI(oaiTools []openai.Tool) ([]*genai.Tool, error) {
	var tools []*genai.Tool
	for _, t := range oaiTools {
		f, err := mapFunctionDefinitionFromOpenAI(t.Function)
		if err != nil {
			return nil, fmt.Errorf("failed to map functions: %w", err)
		}
		if len(f) > 0 {
			tools = append(tools, &genai.Tool{
				FunctionDeclarations: f,
			})
		}
	}

	return tools, nil
}

func mapFunctionDefinitionFromOpenAI(funcDef *openai.FunctionDefinition) ([]*genai.FunctionDeclaration, error) {
	if funcDef == nil {
		return nil, nil
	}
	var functions []*genai.FunctionDeclaration

	var params *genai.Schema
	if funcDef.Parameters != nil {
		pb, err := json.Marshal(funcDef.Parameters)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal function parameters: %w", err)
		}

		if err := json.Unmarshal(pb, &params); err != nil {
			return nil, fmt.Errorf("failed to unmarshal function parameters: %w", err)
		}
	} else {
		params = &genai.Schema{
			Properties: map[string]*genai.Schema{},
			Type:       genai.TypeObject,
		}
	}

	functions = append(functions, &genai.FunctionDeclaration{
		Description: funcDef.Description,
		Name:        funcDef.Name,
		Parameters:  params,
	})
	return functions, nil
}

// openAIEmbeddingRequest - not (yet) provided by the Chat Completion Client package
type openAIEmbeddingRequest struct {
	Input          string `json:"input"`
	Model          string `json:"model"`
	EncodingFormat string `json:"encoding_format,omitempty"`
	Dimensions     *int   `json:"dimensions,omitempty"`
}

type openAIResponse struct {
	Data []openAIResponseData `json:"data"`
}

type openAIResponseData struct {
	Embedding []float32 `json:"embedding"`
}

type vertexEmbeddingResponse struct {
	Predictions []vertexPrediction `json:"predictions"`
}

type vertexPrediction struct {
	Embeddings vertexEmbeddings `json:"embeddings"`
}

type vertexEmbeddings struct {
	Values []float32 `json:"values"`
	// leaving out what we don't need just yet
}

// embeddings - not (yet) provided by the Google GenAI package
func (s *server) embeddings(w http.ResponseWriter, r *http.Request) {
	var er openAIEmbeddingRequest
	if err := json.NewDecoder(r.Body).Decode(&er); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:predict", s.client.ClientConfig().Location, s.client.ClientConfig().Project, s.client.ClientConfig().Location, er.Model)

	payload := map[string]any{
		"instances": []map[string]any{
			{
				"tast_type":  "QUESTION_ANSWERING",
				"content":    er.Input,
				"parameters": map[string]any{},
			},
		},
	}

	if er.Dimensions != nil {
		payload["parameters"] = map[string]any{
			"outputDimensionality": *er.Dimensions,
		}
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't marshal request body: %v", err), http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't create request: %v", err), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.ClientConfig().HTTPClient.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't make request: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), resp.StatusCode)
		return
	}

	var embeddingResponse vertexEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResponse); err != nil {
		http.Error(w, fmt.Sprintf("couldn't decode response: %v", err), http.StatusInternalServerError)
		return
	}

	if len(embeddingResponse.Predictions) == 0 || len(embeddingResponse.Predictions[0].Embeddings.Values) == 0 {
		http.Error(w, "no embeddings found in the response", http.StatusInternalServerError)
		return
	}

	if len(embeddingResponse.Predictions) > 1 {
		fmt.Println("Info: multiple predictions found in the response - using only the first one")
	}

	oaiResp := openAIResponse{
		Data: []openAIResponseData{
			{
				Embedding: embeddingResponse.Predictions[0].Embeddings.Values,
			},
		},
	}

	if err := json.NewEncoder(w).Encode(oaiResp); err != nil {
		http.Error(w, fmt.Sprintf("couldn't encode response: %v", err), http.StatusInternalServerError)
		return
	}
}
