package searchx

import (
	"fmt"
	"sort"
	"strings"
)

// pgParse validates raw SQL is a SELECT — no AST needed for postgres
func (ks *Searchx) pgParse() *Searchx {
	upper := strings.TrimSpace(strings.ToUpper(ks.Raw))
	if !strings.HasPrefix(upper, "SELECT") {
		ks.Err = fmt.Errorf("query must be SELECT statement (query: %s)", ks.Raw)
	}
	return ks
}

// pgParseSelectMapping extracts column aliases from raw SQL for postgres
func (ks *Searchx) pgParseSelectMapping() *Searchx {
	selectClause, err := extractSelectClause(ks.Raw)
	if err != nil {
		ks.Err = err
		return ks
	}

	if strings.TrimSpace(selectClause) == "*" {
		ks.Err = fmt.Errorf("for save filtering, please don't use * on select query")
		return ks
	}

	columns := splitRespectingParens(selectClause)
	mappings := map[string]string{}

	for _, col := range columns {
		col = strings.TrimSpace(col)
		if col == "" {
			continue
		}
		upperCol := strings.ToUpper(col)

		asIdx := strings.Index(upperCol, " AS ")
		if asIdx != -1 {
			expr := strings.TrimSpace(col[:asIdx])
			alias := strings.TrimSpace(col[asIdx+4:])
			mappings[alias] = expr
		} else {
			mappings[col] = col
		}
	}

	if len(mappings) == 0 {
		ks.Err = fmt.Errorf("for save filtering, please don't use * on select query")
		return ks
	}

	ks.MappingSelect = mappings
	return ks
}

// pgParseSortQuery appends ORDER BY to the raw SQL
func (ks *Searchx) pgParseSortQuery(orderBy, orderDir string) *Searchx {
	target := &ks.Raw
	if ks.RawUnion != "" {
		target = &ks.RawUnion
	}

	clause := fmt.Sprintf("%s %s", orderBy, strings.ToUpper(orderDir))
	if hasTopLevelOrderBy(*target) {
		*target += ", " + clause
	} else {
		*target += " ORDER BY " + clause
	}
	return ks
}

// hasTopLevelOrderBy detects ORDER BY at paren depth 0, so ORDER BY inside
// subqueries or aggregate functions (e.g. string_agg(..., ', ' ORDER BY col))
// does not trigger a false positive.
func hasTopLevelOrderBy(sql string) bool {
	upper := strings.ToUpper(sql)
	depth := 0
	for i := 0; i < len(upper); i++ {
		ch := upper[i]
		if ch == '(' {
			depth++
		} else if ch == ')' {
			depth--
		} else if depth == 0 && i+9 <= len(upper) && upper[i:i+9] == " ORDER BY" {
			return true
		}
	}
	return false
}

// pgParseCountQuery wraps raw SQL in a COUNT subquery
func (ks *Searchx) pgParseCountQuery() *Searchx {
	base := ks.Raw
	if ks.RawUnion != "" {
		base = ks.RawUnion
	}
	ks.RawAgg = fmt.Sprintf("SELECT count(*) as agg FROM (%s) AS my_table_count", base)
	return ks
}

// pgParseCurrentPageQuery appends LIMIT/OFFSET
func (ks *Searchx) pgParseCurrentPageQuery(page, perPage int) *Searchx {
	base := ks.Raw
	if ks.RawUnion != "" {
		base = ks.RawUnion
	}
	offset := (page - 1) * perPage
	ks.RawCurrentPage = fmt.Sprintf("%s LIMIT %d OFFSET %d", base, perPage, offset)
	return ks
}

// pgProcessUnion builds UNION query via string concatenation
func (ks *Searchx) pgProcessUnion() *Searchx {
	if len(ks.Unions) == 0 {
		return ks
	}

	if ks.Err != nil {
		return ks
	}

	parts := []string{ks.Raw}
	for i, v_union := range ks.Unions {
		v_union.Calc()
		if v_union.Err != nil {
			ks.Err = fmt.Errorf("union %d failed: %w", i, v_union.Err)
			return ks
		}
		parts = append(parts, v_union.Raw)
	}

	unionSQL := strings.Join(parts, " UNION ")
	ks.RawUnion = fmt.Sprintf("SELECT * FROM (%s) AS my_table", unionSQL)
	return ks
}

// pgParseSummaryQuery builds a summary/aggregate query
func (ks *Searchx) pgParseSummaryQuery() *Searchx {
	if len(ks.SelectSummaries) == 0 {
		ks.Err = fmt.Errorf("no select summary defined")
		return ks
	}

	if ks.Err != nil {
		return ks
	}

	base := ks.Raw
	if ks.RawUnion != "" {
		base = ks.RawUnion
	}

	keys := make([]string, 0, len(ks.SelectSummaries))
	for k := range ks.SelectSummaries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		v := ks.SelectSummaries[k]
		parts = append(parts, fmt.Sprintf("%s as %s", v, k))
	}

	selectColumns := strings.Join(parts, ", ")
	ks.RawSummary = fmt.Sprintf("SELECT %s FROM (%s) my_table_summary", selectColumns, base)
	return ks
}

// extractSelectClause extracts the column list from a SELECT statement at depth 0
func extractSelectClause(sql string) (string, error) {
	upper := strings.ToUpper(sql)

	selectIdx := strings.Index(upper, "SELECT ")
	if selectIdx == -1 {
		return "", fmt.Errorf("not a SELECT statement")
	}
	colsStart := selectIdx + 7

	depth := 0
	for i := colsStart; i < len(upper); i++ {
		switch upper[i] {
		case '(':
			depth++
		case ')':
			depth--
		default:
			if depth == 0 && i+6 <= len(upper) && upper[i:i+6] == " FROM " {
				return strings.TrimSpace(sql[colsStart:i]), nil
			}
		}
	}

	return "", fmt.Errorf("FROM keyword not found in query: %s", sql)
}

// splitRespectingParens splits s by commas at paren depth 0
func splitRespectingParens(s string) []string {
	var parts []string
	depth := 0
	start := 0

	for i, ch := range s {
		switch ch {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, s[start:])
	return parts
}
