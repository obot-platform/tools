package helper

import (
	"fmt"
	"strings"

	cg "github.com/philippgille/chromem-go"
)

func BuildWhereDocumentClause(whereDocs []cg.WhereDocument, joinOperator string) (string, error) {
	if len(whereDocs) == 0 {
		return "TRUE", nil
	}
	if joinOperator == "" {
		joinOperator = "AND"
	}
	joinOperator = fmt.Sprintf(" %s ", strings.TrimSpace(joinOperator)) // ensure space around operator
	var whereClauses []string
	for _, wd := range whereDocs {
		switch wd.Operator {
		case cg.WhereDocumentOperatorAnd:
			wc, err := BuildWhereDocumentClause(wd.WhereDocuments, "AND")
			if err != nil {
				return "", err
			}
			whereClauses = append(whereClauses, fmt.Sprintf("(%s)", wc))
		case cg.WhereDocumentOperatorOr:
			wc, err := BuildWhereDocumentClause(wd.WhereDocuments, "OR")
			if err != nil {
				return "", err
			}
			whereClauses = append(whereClauses, fmt.Sprintf("(%s)", wc))
		case cg.WhereDocumentOperatorEquals:
			whereClauses = append(whereClauses, fmt.Sprintf("document = '%s'", wd.Value))
		case cg.WhereDocumentOperatorContains:
			whereClauses = append(whereClauses, fmt.Sprintf("document LIKE '%%%s%%'", wd.Value))
		case cg.WhereDocumentOperatorNotContains:
			whereClauses = append(whereClauses, fmt.Sprintf("document NOT LIKE '%%%s%%'", wd.Value))
		}
	}
	return strings.Join(whereClauses, joinOperator), nil
}
