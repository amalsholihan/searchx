package searchx

import (
	"fmt"
	"strings"

	"github.com/xwb1989/sqlparser"
)

func (ks *Searchx) Parse() *Searchx {
	stmt, err := sqlparser.Parse(ks.Raw)
	if err != nil {
		ks.Err = fmt.Errorf("parse error (query: %s): %w", ks.Raw, err)
		return ks
	}

	sel, ok := stmt.(*sqlparser.Select)
	if !ok {
		ks.Err = fmt.Errorf("query must be SELECT statement, got: %T (query: %s)", stmt, ks.Raw)
		return ks
	}

	ks.Parsed = sel
	return ks
}

// ParseSelectMapping mengubah SELECT query jadi mapping alias -> expression
func (ks *Searchx) ParseSelectMapping() *Searchx {
	sel, ok := ks.Parsed.(*sqlparser.Select)
	if !ok {
		ks.Err = fmt.Errorf("not a select query")
		return ks
	}

	mappings := map[string]string{}

	for _, expr := range sel.SelectExprs {
		if ae, ok := expr.(*sqlparser.AliasedExpr); ok {
			alias := ae.As.String()
			if alias == "" {
				alias = sqlparser.String(ae.Expr)
			}
			mappings[alias] = sqlparser.String(ae.Expr)
		}
	}

	if len(mappings) == 0 {
		ks.Err = fmt.Errorf("for save filtering, please don't use * on select query")
		return ks
	}

	ks.MappingSelect = mappings
	return ks
}

// ParseCountQuery mengubah SELECT query menjadi SELECT count(*) AS agg ...
func (ks *Searchx) ParseCountQuery() *Searchx {

	// pastikan statement adalah SELECT
	sel_pointer, ok := ks.Parsed.(*sqlparser.Select)
	if !ok {
		ks.Err = fmt.Errorf("not a select query 1")
		return ks
	}
	sel := *sel_pointer

	if ks.UnionParsed != nil {
		sel_pointer_union, ok := ks.UnionParsed.(*sqlparser.Select)
		if !ok {
			ks.Err = fmt.Errorf("not a select query 2")
			return ks
		}
		sel = *sel_pointer_union
	}

	// ganti SELECT expression jadi count(*)
	sel.SelectExprs = sqlparser.SelectExprs{
		&sqlparser.AliasedExpr{
			Expr: &sqlparser.FuncExpr{
				Name: sqlparser.NewColIdent("count"),
				Exprs: sqlparser.SelectExprs{
					&sqlparser.StarExpr{},
				},
			},
			As: sqlparser.NewColIdent("agg"),
		},
	}

	raw_query := sqlparser.String(&sel)
	ks.RawAgg = ks.normalizeQuery(raw_query)

	return ks
}

// ParseCurrentPageQuery membuat query utk page skarang
func (ks *Searchx) ParseCurrentPageQuery(page, per_page int) *Searchx {

	// pastikan statement adalah SELECT
	sel_pointer, ok := ks.Parsed.(*sqlparser.Select)
	if !ok {
		ks.Err = fmt.Errorf("not a select query 3")
		return ks
	}
	sel := *sel_pointer

	if ks.UnionParsed != nil {
		sel_pointer_union, ok := ks.UnionParsed.(*sqlparser.Select)
		if !ok {
			ks.Err = fmt.Errorf("not a select query 4")
			return ks
		}
		sel = *sel_pointer_union
	}

	offset := (page - 1) * per_page

	// bikin literal untuk limit dan offset
	limitVal := sqlparser.NewIntVal([]byte(fmt.Sprintf("%d", per_page)))
	offsetVal := sqlparser.NewIntVal([]byte(fmt.Sprintf("%d", offset)))

	sel.Limit = &sqlparser.Limit{
		Rowcount: limitVal,
		Offset:   offsetVal,
	}

	raw_query := sqlparser.String(&sel)
	// Normalize for database compatibility
	raw_query = ks.normalizeQuery(raw_query)
	ks.RawCurrentPage = raw_query

	return ks
}

// normalizeQuery normalize query for cross-database compatibility
func (ks *Searchx) normalizeQuery(query string) string {
	// Remove backticks (MySQL), double quotes (PostgreSQL), and brackets (SQLite) for identifiers
	query = strings.ReplaceAll(query, "`", "")
	query = strings.ReplaceAll(query, "\"", "")
	query = strings.ReplaceAll(query, "[", "")
	query = strings.ReplaceAll(query, "]", "")
	// Convert LIMIT syntax from MySQL (LIMIT offset, count) to PostgreSQL standard (LIMIT count OFFSET offset)
	query = ks.normalizeLimitSyntax(query)
	return query
}

// normalizeLimitSyntax converts MySQL LIMIT offset, count to LIMIT count OFFSET offset
func (ks *Searchx) normalizeLimitSyntax(query string) string {
	// Match "LIMIT offset, count" pattern and convert to "LIMIT count OFFSET offset"
	upperQuery := strings.ToUpper(query)
	if !strings.Contains(upperQuery, "LIMIT") {
		return query
	}

	// Use regex-like approach: find "LIMIT <number>, <number>" pattern
	// Convert with case-insensitive matching
	limitIdx := -1
	for i := 0; i < len(upperQuery); i++ {
		if i+5 <= len(upperQuery) && strings.ToUpper(query[i:i+5]) == "LIMIT" {
			limitIdx = i
		}
	}

	if limitIdx == -1 {
		return query
	}

	afterLimit := strings.TrimSpace(query[limitIdx+5:])
	// Check if it has ", " pattern (MySQL style)
	if strings.Contains(afterLimit, ",") {
		parts := strings.Split(afterLimit, ",")
		if len(parts) >= 2 {
			offset := strings.TrimSpace(parts[0])
			remaining := strings.TrimSpace(parts[1])
			// Extract the count (first token after comma)
			countTokens := strings.Fields(remaining)
			if len(countTokens) > 0 {
				count := countTokens[0]
				// Build patterns for replacement (case-insensitive)
				oldPattern := fmt.Sprintf("LIMIT %s, %s", offset, count)
				oldPatternLower := fmt.Sprintf("limit %s, %s", offset, count)
				newPattern := fmt.Sprintf("LIMIT %s OFFSET %s", count, offset)

				// Try uppercase first, then lowercase
				if strings.Contains(query, oldPattern) {
					query = strings.Replace(query, oldPattern, newPattern, 1)
				} else if strings.Contains(query, oldPatternLower) {
					query = strings.Replace(query, oldPatternLower, strings.ToLower(newPattern), 1)
				}
			}
		}
	}
	return query
}
