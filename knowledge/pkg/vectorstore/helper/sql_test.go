package helper

import (
	"testing"

	cg "github.com/philippgille/chromem-go"
	"github.com/stretchr/testify/assert"
)

func TestBuildWhereDocumentClause_EmptyInput_TRUEClause(t *testing.T) {
	var whereDocs []cg.WhereDocument
	whereClause, a, err := BuildWhereDocumentClause(whereDocs, "AND")
	assert.NoError(t, err)
	assert.Equal(t, "TRUE", whereClause)
	assert.Empty(t, a)
}

func TestBuildWhereDocumentClause_SingleEqualsCondition_ReturnsCorrectClause(t *testing.T) {
	whereDocs := []cg.WhereDocument{
		{Operator: cg.WhereDocumentOperatorEquals, Value: "test"},
	}
	whereClause, a, err := BuildWhereDocumentClause(whereDocs, "AND")
	assert.NoError(t, err)
	assert.Equal(t, "document = ?", whereClause)
	assert.Equal(t, []any{"test"}, a)
}

func TestBuildWhereDocumentClause_SingleContainsCondition_ReturnsCorrectClause(t *testing.T) {
	whereDocs := []cg.WhereDocument{
		{Operator: cg.WhereDocumentOperatorContains, Value: "test"},
	}
	whereClause, a, err := BuildWhereDocumentClause(whereDocs, "AND")
	assert.NoError(t, err)
	assert.Equal(t, "document LIKE ?", whereClause)
	assert.Equal(t, []any{"%test%"}, a)
}

func TestBuildWhereDocumentClause_SingleNotContainsCondition_ReturnsCorrectClause(t *testing.T) {
	whereDocs := []cg.WhereDocument{
		{Operator: cg.WhereDocumentOperatorNotContains, Value: "test"},
	}
	whereClause, a, err := BuildWhereDocumentClause(whereDocs, "AND")
	assert.NoError(t, err)
	assert.Equal(t, "document NOT LIKE ?", whereClause)
	assert.Equal(t, []any{"%test%"}, a)
}

func TestBuildWhereDocumentClause_AndCondition_ReturnsCorrectClauses(t *testing.T) {
	whereDocs := []cg.WhereDocument{
		{
			Operator: cg.WhereDocumentOperatorAnd,
			WhereDocuments: []cg.WhereDocument{
				{Operator: cg.WhereDocumentOperatorEquals, Value: "test1"},
				{Operator: cg.WhereDocumentOperatorEquals, Value: "test2"},
			},
		},
	}
	whereClause, a, err := BuildWhereDocumentClause(whereDocs, "AND")
	assert.NoError(t, err)
	assert.Equal(t, "(document = ? AND document = ?)", whereClause)
	assert.Equal(t, []any{"test1", "test2"}, a)
}

func TestBuildWhereDocumentClause_OrCondition_ReturnsCorrectClauses(t *testing.T) {
	whereDocs := []cg.WhereDocument{
		{
			Operator: cg.WhereDocumentOperatorOr,
			WhereDocuments: []cg.WhereDocument{
				{Operator: cg.WhereDocumentOperatorEquals, Value: "test1"},
				{Operator: cg.WhereDocumentOperatorEquals, Value: "test2"},
			},
		},
	}
	whereClause, a, err := BuildWhereDocumentClause(whereDocs, "OR")
	assert.NoError(t, err)
	assert.Equal(t, "(document = ? OR document = ?)", whereClause)
	assert.Equal(t, []any{"test1", "test2"}, a)
}

func TestBuildWhereDocumentClause_Nested_ReturnsCorrectClauses(t *testing.T) {
	whereDocs := []cg.WhereDocument{
		{
			Operator: cg.WhereDocumentOperatorOr,
			WhereDocuments: []cg.WhereDocument{
				{Operator: cg.WhereDocumentOperatorEquals, Value: "test1"},
				{Operator: cg.WhereDocumentOperatorEquals, Value: "test2"},
			},
		},
		{
			Operator: cg.WhereDocumentOperatorAnd,
			WhereDocuments: []cg.WhereDocument{
				{Operator: cg.WhereDocumentOperatorEquals, Value: "test3"},
				{Operator: cg.WhereDocumentOperatorEquals, Value: "test4"},
			},
		},
		{
			Operator: cg.WhereDocumentOperatorAnd,
			WhereDocuments: []cg.WhereDocument{
				{
					Operator: cg.WhereDocumentOperatorAnd,
					WhereDocuments: []cg.WhereDocument{
						{Operator: cg.WhereDocumentOperatorEquals, Value: "test5"},
						{Operator: cg.WhereDocumentOperatorEquals, Value: "test6"},
					},
				},
				{Operator: cg.WhereDocumentOperatorEquals, Value: "test7"},
			},
		},
	}
	whereClause, a, err := BuildWhereDocumentClause(whereDocs, "AND")
	assert.NoError(t, err)
	assert.Equal(t, "(document = ? OR document = ?) AND (document = ? AND document = ?) AND ((document = ? AND document = ?) AND document = ?)", whereClause)
	assert.Equal(t, []any{"test1", "test2", "test3", "test4", "test5", "test6", "test7"}, a)
}

func TestBuildWhereDocumentClauseIndexed_Nested_ReturnsCorrectClauses(t *testing.T) {
	whereDocs := []cg.WhereDocument{
		{
			Operator: cg.WhereDocumentOperatorOr,
			WhereDocuments: []cg.WhereDocument{
				{Operator: cg.WhereDocumentOperatorEquals, Value: "test1"},
				{Operator: cg.WhereDocumentOperatorEquals, Value: "test2"},
			},
		},
		{
			Operator: cg.WhereDocumentOperatorAnd,
			WhereDocuments: []cg.WhereDocument{
				{Operator: cg.WhereDocumentOperatorEquals, Value: "test3"},
				{Operator: cg.WhereDocumentOperatorEquals, Value: "test4"},
			},
		},
		{
			Operator: cg.WhereDocumentOperatorAnd,
			WhereDocuments: []cg.WhereDocument{
				{
					Operator: cg.WhereDocumentOperatorAnd,
					WhereDocuments: []cg.WhereDocument{
						{Operator: cg.WhereDocumentOperatorEquals, Value: "test5"},
						{Operator: cg.WhereDocumentOperatorEquals, Value: "test6"},
					},
				},
				{Operator: cg.WhereDocumentOperatorEquals, Value: "test7"},
			},
		},
	}
	whereClause, a, err := BuildWhereDocumentClauseIndexed(whereDocs, "AND", 3)
	assert.NoError(t, err)
	assert.Equal(t, "(document = $3 OR document = $4) AND (document = $5 AND document = $6) AND ((document = $7 AND document = $8) AND document = $9)", whereClause)
	assert.Equal(t, []any{"test1", "test2", "test3", "test4", "test5", "test6", "test7"}, a)
}
