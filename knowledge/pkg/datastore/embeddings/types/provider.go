package types

import (
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
)

type EmbeddingModelProvider interface {
	Name() string
	EmbeddingFunc() (vs.EmbeddingFunc, error)
	Configure() error
	Config() any
	EmbeddingModelName() string
	UseEmbeddingModel(model string)
}
