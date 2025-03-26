package vectorstore

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	etypes "github.com/obot-platform/tools/knowledge/pkg/datastore/embeddings/types"
	dbtypes "github.com/obot-platform/tools/knowledge/pkg/index/types"
	"github.com/obot-platform/tools/knowledge/pkg/vectorstore/pgvector"
	sqlitevec "github.com/obot-platform/tools/knowledge/pkg/vectorstore/sqlite-vec"
	"github.com/obot-platform/tools/knowledge/pkg/vectorstore/types"
)

type VectorStore interface {
	CreateCollection(ctx context.Context, collection string, opts *dbtypes.DatasetCreateOpts) error
	AddDocuments(ctx context.Context, docs []types.Document, collection string) ([]string, error)                                                                                                                       // @return documentIDs, error
	SimilaritySearch(ctx context.Context, query string, numDocuments int, collection string, where map[string]string, whereDocument []types.WhereDocument, embeddingFunc types.EmbeddingFunc) ([]types.Document, error) //nolint:lll
	RemoveCollection(ctx context.Context, collection string) error
	RemoveDocument(ctx context.Context, documentID string, collection string, where map[string]string, whereDocument []types.WhereDocument) error
	GetDocuments(ctx context.Context, collection string, where map[string]string, whereDocument []types.WhereDocument) ([]types.Document, error)
	GetDocument(ctx context.Context, documentID string, collection string) (types.Document, error)

	ImportCollectionsFromFile(ctx context.Context, path string, collections ...string) error
	ExportCollectionsToFile(ctx context.Context, path string, collections ...string) error

	Close() error
}

func New(ctx context.Context, dsn string, embeddingProvider etypes.EmbeddingModelProvider) (VectorStore, error) {
	embeddingFunc, err := embeddingProvider.EmbeddingFunc()
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding function: %w", err)
	}

	dialect := strings.Split(dsn, "://")[0]

	slog.Debug("vectordb", "dialect", dialect, "dsn", dsn)

	switch dialect {
	case "pgvector":
		return pgvector.New(ctx, dsn, embeddingFunc)
	case "sqlite-vec":
		return sqlitevec.New(ctx, dsn, embeddingFunc)
	default:
		return nil, fmt.Errorf("unsupported dialect: %q", dialect)
	}
}
