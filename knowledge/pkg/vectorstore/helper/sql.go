package helper

import (
	"fmt"
	"strings"

	cg "github.com/philippgille/chromem-go"
)

func BuildWhereDocumentClauseIndexed(whereDocs []cg.WhereDocument, joinOperator string, argIndex int) (string, []any, error) {
	if len(whereDocs) == 0 {
		return "TRUE", nil, nil
	}
	if joinOperator == "" {
		joinOperator = "AND"
	}
	joinOperator = fmt.Sprintf(" %s ", strings.TrimSpace(joinOperator)) // ensure space around operator
	var whereClauses []string
	var args []any
	for _, wd := range whereDocs {
		switch wd.Operator {
		case cg.WhereDocumentOperatorAnd:
			wc, a, err := BuildWhereDocumentClauseIndexed(wd.WhereDocuments, "AND", argIndex)
			if err != nil {
				return "", nil, err
			}
			whereClauses = append(whereClauses, fmt.Sprintf("(%s)", wc))
			args = append(args, a...)
			argIndex += len(a)
		case cg.WhereDocumentOperatorOr:
			wc, a, err := BuildWhereDocumentClauseIndexed(wd.WhereDocuments, "OR", argIndex)
			if err != nil {
				return "", nil, err
			}
			whereClauses = append(whereClauses, fmt.Sprintf("(%s)", wc))
			args = append(args, a...)
			argIndex += len(a)
		case cg.WhereDocumentOperatorEquals:
			whereClauses = append(whereClauses, fmt.Sprintf("document = $%d", argIndex))
			args = append(args, wd.Value)
			argIndex += 1
		case cg.WhereDocumentOperatorContains:
			whereClauses = append(whereClauses, fmt.Sprintf("document LIKE $%d", argIndex))
			args = append(args, "%"+wd.Value+"%")
			argIndex += 1
		case cg.WhereDocumentOperatorNotContains:
			whereClauses = append(whereClauses, fmt.Sprintf("document NOT LIKE $%d", argIndex))
			args = append(args, "%"+wd.Value+"%")
			argIndex += 1
		}
	}
	return strings.Join(whereClauses, joinOperator), args, nil
}

func BuildWhereDocumentClause(whereDocs []cg.WhereDocument, joinOperator string) (string, []any, error) {
	if len(whereDocs) == 0 {
		return "TRUE", nil, nil
	}
	if joinOperator == "" {
		joinOperator = "AND"
	}
	joinOperator = fmt.Sprintf(" %s ", strings.TrimSpace(joinOperator)) // ensure space around operator
	var whereClauses []string
	var args []any
	for _, wd := range whereDocs {
		switch wd.Operator {
		case cg.WhereDocumentOperatorAnd:
			wc, a, err := BuildWhereDocumentClause(wd.WhereDocuments, "AND")
			if err != nil {
				return "", nil, err
			}
			whereClauses = append(whereClauses, fmt.Sprintf("(%s)", wc))
			args = append(args, a...)
		case cg.WhereDocumentOperatorOr:
			wc, a, err := BuildWhereDocumentClause(wd.WhereDocuments, "OR")
			if err != nil {
				return "", nil, err
			}
			whereClauses = append(whereClauses, fmt.Sprintf("(%s)", wc))
			args = append(args, a...)
		case cg.WhereDocumentOperatorEquals:
			whereClauses = append(whereClauses, fmt.Sprintf("document = ?"))
			args = append(args, wd.Value)
		case cg.WhereDocumentOperatorContains:
			whereClauses = append(whereClauses, fmt.Sprintf("document LIKE ?"))
			args = append(args, "%"+wd.Value+"%")
		case cg.WhereDocumentOperatorNotContains:
			whereClauses = append(whereClauses, fmt.Sprintf("document NOT LIKE ?"))
			args = append(args, "%"+wd.Value+"%")
		}
	}
	return strings.Join(whereClauses, joinOperator), args, nil
}
