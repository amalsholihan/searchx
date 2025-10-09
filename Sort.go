package searchx

import (
	"fmt"
	"strings"

	"github.com/xwb1989/sqlparser"
)

func (ks *Searchx) Sort(params []map[string]string) *Searchx {
	ks.SortParams = params
	return ks
}

// untuk membuat query sort
func (ks *Searchx) ParseSortQuery(orderBy string, orderDir string) *Searchx {
	// pastikan statement adalah SELECT
	selPointer, ok := ks.Parsed.(*sqlparser.Select)
	if !ok {
		ks.Err = fmt.Errorf("not a select query 1")
		return ks
	}
	sel := *selPointer

	// kalau ada UNION, pakai select terakhir
	if ks.UnionParsed != nil {
		selPointerUnion, ok := ks.UnionParsed.(*sqlparser.Select)
		if !ok {
			ks.Err = fmt.Errorf("not a select query 2")
			return ks
		}
		sel = *selPointerUnion
	}

	newOrder := &sqlparser.Order{
		Expr:      &sqlparser.ColName{Name: sqlparser.NewColIdent(orderBy)},
		Direction: strings.ToUpper(orderDir),
	}
	sel.OrderBy = append(sel.OrderBy, newOrder)

	rawQuery := sqlparser.String(&sel)
	if ks.UnionParsed != nil {
		ks.UnionParsed = &sel
		ks.RawUnion = rawQuery
	} else {
		ks.Raw = rawQuery
		ks.Parsed = &sel
	}

	return ks
}

func (ks *Searchx) ProcessSort() *Searchx {
	if len(ks.SortParams) <= 0 {
		return ks
	}

	for _, sort_param := range ks.SortParams {
		if sort_param["sort_column"] == "" {
			ks.Err = fmt.Errorf("sort column is required")
			return ks
		}
		sortColumn := ks.ValidateColumn(sort_param["sort_column"])
		if sortColumn == "" {
			ks.Err = fmt.Errorf("column sort %v not found in select statement", sort_param["sort_column"])
			return ks
		}
		if sort_param["sort_type"] == "" {
			ks.Err = fmt.Errorf("sort type is required")
			return ks
		}
		sortType := ks.ValidateSortType(sort_param["sort_type"])
		if sortType == "" {
			ks.Err = fmt.Errorf("sort type %v is invalid", sort_param["sort_type"])
			return ks
		}
		ks.ParseSortQuery(sortColumn, sortType)
	}

	return ks
}
