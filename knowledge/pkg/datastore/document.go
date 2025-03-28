package datastore

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/knowledge/pkg/vectorstore/types"
)

func (s *Datastore) DeleteDocument(ctx context.Context, documentID, datasetID string) error {
	// Remove from Index
	if err := s.Index.DeleteDocument(ctx, documentID, datasetID); err != nil {
		return fmt.Errorf("failed to remove document from Index: %w", err)
	}

	// Remove from VectorStore
	if err := s.Vectorstore.RemoveDocument(ctx, documentID, datasetID, nil, nil); err != nil {
		return fmt.Errorf("failed to remove document from VectorStore: %w", err)
	}

	return nil
}

func (s *Datastore) GetDocuments(ctx context.Context, datasetID string, where map[string]string, whereDocument []types.WhereDocument) ([]types.Document, error) {
	return s.Vectorstore.GetDocuments(ctx, datasetID, where, whereDocument)
}
