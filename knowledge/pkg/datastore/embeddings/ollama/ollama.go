package ollama

import (
	"fmt"
	"strings"

	"dario.cat/mergo"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/load"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/openai"
	cg "github.com/philippgille/chromem-go"
)

type EmbeddingProviderOllama struct {
	BaseURL string `koanf:"baseURL" env:"OLLAMA_BASE_URL"`
	Model   string `koanf:"model" env:"OLLAMA_MODEL" export:"required"`
}

func (p *EmbeddingProviderOllama) UseEmbeddingModel(model string) {
	p.Model = model
}

const EmbeddingProviderOllamaName = "ollama"

func (p *EmbeddingProviderOllama) EmbeddingModelName() string {
	return p.Model
}

func (p *EmbeddingProviderOllama) Name() string {
	return EmbeddingProviderOllamaName
}

func (p *EmbeddingProviderOllama) Configure() error {
	if err := load.FillConfigEnv(strings.ToUpper(EmbeddingProviderOllamaName), &p); err != nil {
		return fmt.Errorf("failed to fill Ollama config from environment: %w", err)
	}

	if err := p.fillDefaults(); err != nil {
		return fmt.Errorf("failed to fill Ollama defaults: %w", err)
	}

	return nil
}

func (p *EmbeddingProviderOllama) fillDefaults() error {
	defaultCfg := EmbeddingProviderOllama{
		Model:   "mxbai-embed-large",
		BaseURL: "http://localhost:11434/v1",
	}

	if err := mergo.Merge(p, defaultCfg); err != nil {
		return fmt.Errorf("failed to merge Ollama config: %w", err)
	}

	return nil
}

func (p *EmbeddingProviderOllama) EmbeddingFunc() (cg.EmbeddingFunc, error) {
	cfg := openai.NewOpenAICompatConfig(p.BaseURL, "", p.Model)
	return openai.NewEmbeddingFuncOpenAICompat(cfg), nil
}

func (p *EmbeddingProviderOllama) Config() any {
	return p
}
