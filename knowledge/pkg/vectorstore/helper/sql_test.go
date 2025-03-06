package helper

import (
	"testing"

	cg "github.com/philippgille/chromem-go"
	"github.com/stretchr/testify/assert"
)

func TestBuildWhereDocumentClause_EmptyInput_TRUEClause(t *testing.T) {
	var whereDocs []cg.WhereDocument
	whereClause, err := BuildWhereDocumentClause(whereDocs, "AND")
	assert.NoError(t, err)
	assert.Equal(t, "TRUE", whereClause)
}

func TestBuildWhereDocumentClause_SingleEqualsCondition_ReturnsCorrectClause(t *testing.T) {
	whereDocs := []cg.WhereDocument{
		{Operator: cg.WhereDocumentOperatorEquals, Value: "test"},
	}
	whereClause, err := BuildWhereDocumentClause(whereDocs, "AND")
	assert.NoError(t, err)
	assert.Equal(t, "document = 'test'", whereClause)
}

func TestBuildWhereDocumentClause_SingleContainsCondition_ReturnsCorrectClause(t *testing.T) {
	whereDocs := []cg.WhereDocument{
		{Operator: cg.WhereDocumentOperatorContains, Value: "test"},
	}
	whereClause, err := BuildWhereDocumentClause(whereDocs, "AND")
	assert.NoError(t, err)
	assert.Equal(t, "document LIKE '%test%'", whereClause)
}

func TestBuildWhereDocumentClause_SingleNotContainsCondition_ReturnsCorrectClause(t *testing.T) {
	whereDocs := []cg.WhereDocument{
		{Operator: cg.WhereDocumentOperatorNotContains, Value: "test"},
	}
	whereClause, err := BuildWhereDocumentClause(whereDocs, "AND")
	assert.NoError(t, err)
	assert.Equal(t, "document NOT LIKE '%test%'", whereClause)
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
	whereClause, err := BuildWhereDocumentClause(whereDocs, "AND")
	assert.NoError(t, err)
	assert.Equal(t, "(document = 'test1' AND document = 'test2')", whereClause)
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
	whereClause, err := BuildWhereDocumentClause(whereDocs, "OR")
	assert.NoError(t, err)
	assert.Equal(t, "(document = 'test1' OR document = 'test2')", whereClause)
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
	whereClause, err := BuildWhereDocumentClause(whereDocs, "AND")
	assert.NoError(t, err)
	assert.Equal(t, "(document = 'test1' OR document = 'test2') AND (document = 'test3' AND document = 'test4') AND ((document = 'test5' AND document = 'test6') AND document = 'test7')", whereClause)
}
