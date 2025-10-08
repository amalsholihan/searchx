package searchx

import (
	"fmt"

	"github.com/xwb1989/sqlparser"
)

func (ks *Searchx) Parse() *Searchx {
	stmt, err := sqlparser.Parse(ks.Raw)
	if err != nil {
		ks.Err = fmt.Errorf("parse error: %w", err)
		return ks
	}

	ks.Parsed = stmt.(*sqlparser.Select)
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
	ks.RawAgg = raw_query

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
	ks.RawCurrentPage = raw_query

	return ks
}
