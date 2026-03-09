package searchx

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SearchCondition represents a single search condition
type SearchCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// SearchGroup represents a group of search conditions with AND/OR logic
type SearchGroup struct {
	AND []SearchGroupItem `json:"and,omitempty"`
	OR  []SearchGroupItem `json:"or,omitempty"`
}

// SearchGroupItem can be either a SearchCondition or a nested SearchGroup
type SearchGroupItem struct {
	Condition *SearchCondition `json:"-"` // Don't use JSON tag, handle manually
	Group     *SearchGroup     `json:"-"` // Don't use JSON tag, handle manually
}

// SearchRequest represents the complete search request
type SearchRequest struct {
	Search SearchGroup `json:"search"`
}

// UnmarshalJSON for SearchGroupItem to handle both condition and group
func (s *SearchGroupItem) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as SearchCondition first
	var cond SearchCondition
	if err := json.Unmarshal(data, &cond); err == nil && cond.Field != "" {
		s.Condition = &cond
		return nil
	}

	// Try to unmarshal as SearchGroup (with "and" or "or" key)
	var group SearchGroup
	if err := json.Unmarshal(data, &group); err == nil {
		s.Group = &group
		return nil
	}

	return fmt.Errorf("invalid search group item")
}

func (ks *Searchx) Search(params []map[string]any) *Searchx {
	ks.SearchParams = params
	return ks
}

// SearchWithJSON accepts JSON string with nested AND/OR logic
// Example: {"search": {"and": [{"field": "status", "operator": "=", "value": "active"}]}}
func (ks *Searchx) SearchWithJSON(jsonStr string) *Searchx {
	var req SearchRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		ks.Err = fmt.Errorf("failed to parse search JSON: %v", err)
		return ks
	}

	// Convert JSON structure to SearchParams format
	params := ks.convertSearchGroupToParams(&req.Search)
	ks.SearchParams = params
	return ks
}

// convertSearchGroupToParams recursively converts SearchGroup to nested []map[string]any structure
func (ks *Searchx) convertSearchGroupToParams(group *SearchGroup) []map[string]any {
	if group == nil {
		return []map[string]any{}
	}

	// Return the nested structure directly
	result := []map[string]any{}

	// Process AND conditions
	if len(group.AND) > 0 {
		andItems := []map[string]any{}
		for _, item := range group.AND {
			if item.Condition != nil {
				andItems = append(andItems, map[string]any{
					"search_column":    item.Condition.Field,
					"search_condition": item.Condition.Operator,
					"search_text":      fmt.Sprintf("%v", item.Condition.Value),
					"search_operator":  "AND",
				})
			} else if item.Group != nil {
				// Recursively process nested group
				nestedParams := ks.convertSearchGroupToParams(item.Group)
				andItems = append(andItems, nestedParams...)
			}
		}
		if len(andItems) > 0 {
			result = append(result, map[string]any{
				"and": andItems,
			})
		}
	}

	// Process OR conditions
	if len(group.OR) > 0 {
		orItems := []map[string]any{}
		for _, item := range group.OR {
			if item.Condition != nil {
				orItems = append(orItems, map[string]any{
					"search_column":    item.Condition.Field,
					"search_condition": item.Condition.Operator,
					"search_text":      fmt.Sprintf("%v", item.Condition.Value),
					"search_operator":  "OR",
				})
			} else if item.Group != nil {
				// Recursively process nested group
				nestedParams := ks.convertSearchGroupToParams(item.Group)
				orItems = append(orItems, nestedParams...)
			}
		}
		if len(orItems) > 0 {
			result = append(result, map[string]any{
				"or": orItems,
			})
		}
	}

	return result
}

func (ks *Searchx) ProcessSearch() *Searchx {
	params := ks.SearchParams
	query := ks.DB

	// Process nested search params
	whereSQL, whereArgs, havingClauses, havingArgs := ks.processNestedParams(params, 0)

	if whereSQL != "" {
		query = query.Where(whereSQL, whereArgs...)
	}

	// gabung semua HAVING yang terkumpul
	if len(havingClauses) > 0 {
		havingSQL := strings.Join(havingClauses, " ")
		query = query.Having(havingSQL, havingArgs...)
	}

	if len(ks.Unions) > 0 {
		for _, v := range ks.Unions {
			v.Search(params)
		}
	}

	ks.DB = query
	ks.SetRawQuery()
	if ks.Err != nil {
		return ks
	}

	ks.Parse()

	return ks
}

// processNestedParams recursively processes nested search params and returns WHERE clause
func (ks *Searchx) processNestedParams(params []map[string]any, depth int) (string, []interface{}, []string, []interface{}) {
	if len(params) == 0 {
		return "", nil, nil, nil
	}

	var andClauses []string
	var andArgs []interface{}
	var orClauses []string
	var orArgs []interface{}
	var havingClauses []string
	var havingArgs []interface{}

	for _, p := range params {
		// Check if this is a nested group with "search" key (default to AND)
		if searchItems, ok := p["search"].([]map[string]any); ok {
			nestedSQL, nestedArgs, nestedHaving, nestedHavingArgs := ks.processNestedParams(searchItems, depth+1)
			if nestedSQL != "" {
				if len(andClauses) > 0 || len(orClauses) > 0 {
					andClauses = append(andClauses, "("+nestedSQL+")")
					andArgs = append(andArgs, nestedArgs...)
				} else {
					andClauses = append(andClauses, nestedSQL)
					andArgs = append(andArgs, nestedArgs...)
				}
			}
			havingClauses = append(havingClauses, nestedHaving...)
			havingArgs = append(havingArgs, nestedHavingArgs...)
			continue
		}

		// Check if this is a nested group
		if andItems, ok := p["and"].([]map[string]any); ok {
			nestedSQL, nestedArgs, nestedHaving, nestedHavingArgs := ks.processNestedParams(andItems, depth+1)
			if nestedSQL != "" {
				if len(andClauses) > 0 || len(orClauses) > 0 {
					andClauses = append(andClauses, "("+nestedSQL+")")
					andArgs = append(andArgs, nestedArgs...)
				} else {
					andClauses = append(andClauses, nestedSQL)
					andArgs = append(andArgs, nestedArgs...)
				}
			}
			havingClauses = append(havingClauses, nestedHaving...)
			havingArgs = append(havingArgs, nestedHavingArgs...)
			continue
		}

		if orItems, ok := p["or"].([]map[string]any); ok {
			nestedSQL, nestedArgs, nestedHaving, nestedHavingArgs := ks.processNestedParams(orItems, depth+1)
			if nestedSQL != "" {
				if len(andClauses) > 0 || len(orClauses) > 0 {
					orClauses = append(orClauses, "("+nestedSQL+")")
					orArgs = append(orArgs, nestedArgs...)
				} else {
					orClauses = append(orClauses, nestedSQL)
					orArgs = append(orArgs, nestedArgs...)
				}
			}
			havingClauses = append(havingClauses, nestedHaving...)
			havingArgs = append(havingArgs, nestedHavingArgs...)
			continue
		}

		// Process flat condition
		val := fmt.Sprintf("%v", p["search_text"])
		col := ks.ValidateColumn(fmt.Sprintf("%v", p["search_column"]))
		if col == "" {
			ks.Err = fmt.Errorf("column search %v not found in select statement", p["search_column"])
			return "", nil, nil, nil
		}
		if p["search_operator"] == "" || p["search_operator"] == nil {
			p["search_operator"] = "AND"
		}
		op := ks.ValidateOperator(fmt.Sprintf("%v", p["search_operator"]))
		if op == "" {
			ks.Err = fmt.Errorf("operator %v not allowed", p["search_operator"])
			return "", nil, nil, nil
		}
		if p["search_condition"] == "" || p["search_condition"] == nil {
			p["search_condition"] = "="
		}
		cond := ks.ValidateCondition(fmt.Sprintf("%v", p["search_condition"]))
		if cond == "" {
			ks.Err = fmt.Errorf("condition %v not allowed", p["search_condition"])
			return "", nil, nil, nil
		}

		if col == "" {
			continue
		}
		if strings.ToLower(op) != "or" {
			op = "and"
		}

		q := ""
		var args []interface{}

		switch cond {
		case "=", ">", ">=", "<", "<=":
			q = fmt.Sprintf("%s %s ?", col, cond)
			args = append(args, val)
		case "like":
			q = fmt.Sprintf("%s LIKE ?", col)
			args = append(args, "%"+val+"%")
		case "is null":
			q = fmt.Sprintf("%s IS NULL", col)
		case "is not null":
			q = fmt.Sprintf("%s IS NOT NULL", col)
		default:
			q = fmt.Sprintf("%s LIKE ?", col)
			args = append(args, "%"+val+"%")
		}

		// deteksi fungsi agregat
		isAggregate := strings.Contains(strings.ToUpper(col), "COUNT(") ||
			strings.Contains(strings.ToUpper(col), "SUM(") ||
			strings.Contains(strings.ToUpper(col), "AVG(") ||
			strings.Contains(strings.ToUpper(col), "MIN(") ||
			strings.Contains(strings.ToUpper(col), "MAX(")

		if isAggregate {
			// simpan ke buffer HAVING
			havingClauses = append(havingClauses, strings.ToUpper(op))
			havingClauses = append(havingClauses, q)
			havingArgs = append(havingArgs, args...)
		} else {
			// kondisi biasa -> tetap pakai WHERE
			if strings.ToLower(op) == "and" {
				andClauses = append(andClauses, q)
				andArgs = append(andArgs, args...)
			} else {
				orClauses = append(orClauses, q)
				orArgs = append(orArgs, args...)
			}
		}
	}

	// Combine AND and OR clauses
	var whereSQL string
	var whereArgs []interface{}

	if len(andClauses) > 0 && len(orClauses) > 0 {
		andSQL := strings.Join(andClauses, " AND ")
		orSQL := "(" + strings.Join(orClauses, " OR ") + ")"
		whereSQL = andSQL + " AND " + orSQL
		whereArgs = append(andArgs, orArgs...)
	} else if len(andClauses) > 0 {
		whereSQL = strings.Join(andClauses, " AND ")
		whereArgs = andArgs
	} else if len(orClauses) > 0 {
		whereSQL = "(" + strings.Join(orClauses, " OR ") + ")"
		whereArgs = orArgs
	}

	return whereSQL, whereArgs, havingClauses, havingArgs
}
