package types

// Disclaimer: a lot of this was taken over from our custom fork of github.com/philippgille/chromem-go which deviated quite a lot from the original,
// and we ended up only still needing some low level abstractions.

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

// EmbeddingFunc must return a normalized vector
type EmbeddingFunc func(ctx context.Context, text string) ([]float32, error)

type Document struct {
	ID              string         `json:"id"`
	Content         string         `json:"content"`
	Metadata        map[string]any `json:"metadata"`
	SimilarityScore float32        `json:"similarity_score"`
	Embedding       []float32      `json:"embedding,omitempty"`
}

const (
	DocMetadataKeyDocIndex  = "docIndex"
	DocMetadataKeyDocsTotal = "docsTotal"
)

func mustInt(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		panic("Unsupported type")
	}
}

func SortDocumentsByMetadata(documents []Document, metadataKey string) {
	// Sort the documents by the metadata field, if present - else we have to assume the order is correct
	slices.SortFunc(documents, func(i, j Document) int {
		iVal, ok := i.Metadata[metadataKey]
		if !ok {
			return 0
		}
		jVal, ok := j.Metadata[metadataKey]
		if !ok {
			return 0
		}

		// Can be either int or float64 (if read from json)
		return mustInt(iVal) - mustInt(jVal)
	})
}

func SortDocumentsByDocIndex(documents []Document) {
	SortDocumentsByMetadata(documents, DocMetadataKeyDocIndex)
}

func SortAndEnsureDocIndex(documents []Document) {
	SortDocumentsByDocIndex(documents)
	l := len(documents)
	for i, doc := range documents {
		doc.Metadata[DocMetadataKeyDocIndex] = i
		doc.Metadata[DocMetadataKeyDocsTotal] = l
	}
}

type WhereDocumentOperator string

const (
	WhereDocumentOperatorEquals      WhereDocumentOperator = "$eq"
	WhereDocumentOperatorContains    WhereDocumentOperator = "$contains"
	WhereDocumentOperatorNotContains WhereDocumentOperator = "$not_contains"
	WhereDocumentOperatorOr          WhereDocumentOperator = "$or"
	WhereDocumentOperatorAnd         WhereDocumentOperator = "$and"
)

type WhereDocument struct {
	Operator       WhereDocumentOperator
	Value          string
	WhereDocuments []WhereDocument
}

func (wd *WhereDocument) Validate() error {
	if !slices.Contains([]WhereDocumentOperator{WhereDocumentOperatorContains, WhereDocumentOperatorNotContains, WhereDocumentOperatorOr, WhereDocumentOperatorAnd}, wd.Operator) {
		return fmt.Errorf("unsupported where document operator %s", wd.Operator)
	}

	if wd.Operator == "" {
		return fmt.Errorf("where document operator is empty")
	}

	// $eq, $contains and $not_contains require a string value
	if slices.Contains([]WhereDocumentOperator{WhereDocumentOperatorEquals, WhereDocumentOperatorContains, WhereDocumentOperatorNotContains}, wd.Operator) {
		if wd.Value == "" {
			return fmt.Errorf("where document operator %s requires a value", wd.Operator)
		}
	}

	// $or requires sub-filters
	if slices.Contains([]WhereDocumentOperator{WhereDocumentOperatorOr, WhereDocumentOperatorAnd}, wd.Operator) {
		if len(wd.WhereDocuments) == 0 {
			return fmt.Errorf("where document operator %s must have at least one sub-filter", wd.Operator)
		}
	}

	for _, wd := range wd.WhereDocuments {
		if err := wd.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Matches checks if a document matches the WhereDocument filter(s)
// There is no error checking on the WhereDocument struct, so it must be validated before calling this function.
func (wd *WhereDocument) Matches(doc *Document) bool {
	switch wd.Operator {
	case WhereDocumentOperatorEquals:
		return doc.Content == wd.Value
	case WhereDocumentOperatorContains:
		return strings.Contains(doc.Content, wd.Value)
	case WhereDocumentOperatorNotContains:
		return !strings.Contains(doc.Content, wd.Value)
	case WhereDocumentOperatorOr:
		for _, subFilter := range wd.WhereDocuments {
			if subFilter.Matches(doc) {
				return true
			}
		}
		return false
	case WhereDocumentOperatorAnd:
		for _, subFilter := range wd.WhereDocuments {
			if !subFilter.Matches(doc) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
