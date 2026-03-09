package searchx

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xwb1989/sqlparser"
)

// SortCondition represents a single sort condition
type SortCondition struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

// SortRequest represents the complete sort request
type SortRequest struct {
	Sort []SortCondition `json:"sort"`
}

func (ks *Searchx) Sort(params []map[string]any) *Searchx {
	// Convert all []interface{} to []map[string]any recursively
	convertedParams := ks.convertSortParamsRecursively(params)
	ks.SortParams = convertedParams
	return ks
}

// convertInterfaceSliceToMapSlice converts []interface{} to []map[string]any
func (ks *Searchx) convertInterfaceSliceToMapSlice(items []interface{}) []map[string]any {
	result := []map[string]any{}
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			converted := map[string]any{}
			for k, v := range m {
				converted[k] = v
			}
			result = append(result, converted)
		} else if m, ok := item.(map[string]any); ok {
			result = append(result, m)
		}
	}
	return result
}

// convertSortParamsRecursively converts all []interface{} to []map[string]any recursively
func (ks *Searchx) convertSortParamsRecursively(params []map[string]any) []map[string]any {
	result := []map[string]any{}
	for _, p := range params {
		converted := map[string]any{}
		for k, v := range p {
			if k == "sort" {
				if items, ok := v.([]map[string]any); ok {
					converted[k] = ks.convertSortParamsRecursively(items)
				} else if items, ok := v.([]interface{}); ok {
					converted[k] = ks.convertInterfaceSliceToMapSlice(items)
				} else {
					converted[k] = v
				}
			} else {
				converted[k] = v
			}
		}
		result = append(result, converted)
	}
	return result
}

// SortWithJSON accepts JSON string with sort conditions
// Example: {"sort": [{"field": "name", "direction": "asc"}, {"field": "id", "direction": "desc"}]}
func (ks *Searchx) SortWithJSON(jsonStr string) *Searchx {
	var req SortRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		ks.Err = fmt.Errorf("failed to parse sort JSON: %v", err)
		return ks
	}

	// Convert JSON structure to SortParams format
	params := ks.convertSortConditionsToParams(req.Sort)
	ks.SortParams = params
	return ks
}

// convertSortConditionsToParams converts []SortCondition to []map[string]any
func (ks *Searchx) convertSortConditionsToParams(conditions []SortCondition) []map[string]any {
	var params []map[string]any

	for _, cond := range conditions {
		params = append(params, map[string]any{
			"sort_column": cond.Field,
			"sort_type":   cond.Direction,
		})
	}

	return params
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
