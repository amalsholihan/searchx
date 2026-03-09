package searchx

import (
	"fmt"
	"strings"

	"github.com/xwb1989/sqlparser"
)

func (ks *Searchx) Sort(params interface{}) *Searchx {
	// Support both map[string]interface{} and []map[string]any
	switch v := params.(type) {
	case map[string]interface{}:
		// Convert map[string]interface{} to []map[string]any format
		convertedParams := ks.convertSortMapToParams(v)
		ks.SortParams = convertedParams
	case []map[string]any:
		// Use []map[string]any directly
		ks.SortParams = v
	case []interface{}:
		// Convert []interface{} to []map[string]any format
		convertedParams := ConvertInterfaceSliceToMapSlice(v)
		ks.SortParams = convertedParams
	default:
		ks.Err = fmt.Errorf("invalid sort params type: %T", params)
	}
	return ks
}

// convertSortMapToParams converts map[string]interface{} to []map[string]any format
func (ks *Searchx) convertSortMapToParams(params map[string]interface{}) []map[string]any {
	result := []map[string]any{}

	// Process sort conditions
	if sortItems, ok := params["sort"].([]interface{}); ok {
		sortParams := []map[string]any{}
		for _, item := range sortItems {
			if m, ok := item.(map[string]interface{}); ok {
				// Check if this is a nested group
				if _, hasSort := m["sort"]; hasSort {
					nestedParams := ks.convertSortMapToParams(m)
					sortParams = append(sortParams, nestedParams...)
				} else {
					// This is a flat condition
					sortParams = append(sortParams, map[string]any{
						"sort_column": m["field"],
						"sort_type":   m["direction"],
					})
				}
			}
		}
		if len(sortParams) > 0 {
			result = append(result, map[string]any{
				"sort": sortParams,
			})
		}
	}

	return result
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
	// Normalize query for cross-database compatibility
	rawQuery = strings.ReplaceAll(rawQuery, "`", "")
	rawQuery = strings.ReplaceAll(rawQuery, "\"", "")
	rawQuery = strings.ReplaceAll(rawQuery, "[", "")
	rawQuery = strings.ReplaceAll(rawQuery, "]", "")

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

	ks.processNestedSortParams(ks.SortParams)

	return ks
}

// processNestedSortParams recursively processes nested sort params
func (ks *Searchx) processNestedSortParams(params []map[string]any) {
	for _, sort_param := range params {
		// Check if this is a nested group with "sort" key
		if sortItems, ok := sort_param["sort"].([]map[string]any); ok {
			// Recursively process nested sort params
			ks.processNestedSortParams(sortItems)
			continue
		}

		// Process flat sort condition
		if sort_param["sort_column"] == "" || sort_param["sort_column"] == nil {
			ks.Err = fmt.Errorf("sort column is required")
			return
		}
		sortColumn := ks.ValidateColumn(fmt.Sprintf("%v", sort_param["sort_column"]))
		if sortColumn == "" {
			ks.Err = fmt.Errorf("column sort %v not found in select statement", sort_param["sort_column"])
			return
		}
		if sort_param["sort_type"] == "" || sort_param["sort_type"] == nil {
			ks.Err = fmt.Errorf("sort type is required")
			return
		}
		sortType := ks.ValidateSortType(fmt.Sprintf("%v", sort_param["sort_type"]))
		if sortType == "" {
			ks.Err = fmt.Errorf("sort type %v is invalid", sort_param["sort_type"])
			return
		}
		ks.ParseSortQuery(sortColumn, sortType)
	}
}
