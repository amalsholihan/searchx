package searchx

import (
	"fmt"
	"strings"
)

func (ks *Searchx) Search(params interface{}) *Searchx {
	// Support both map[string]interface{} and []map[string]any
	switch v := params.(type) {
	case map[string]interface{}:
		// Convert map[string]interface{} to []map[string]any format
		convertedParams := ks.convertSearchMapToParams(v)
		ks.SearchParams = convertedParams
	case []map[string]any:
		// Use []map[string]any directly
		ks.SearchParams = v
	default:
		ks.Err = fmt.Errorf("invalid search params type: %T", params)
	}
	return ks
}

// convertSearchMapToParams converts map[string]interface{} to []map[string]any format
func (ks *Searchx) convertSearchMapToParams(params map[string]interface{}) []map[string]any {
	result := []map[string]any{}

	// Process AND conditions
	if andItems, ok := params["and"].([]interface{}); ok {
		andParams := []map[string]any{}
		for _, item := range andItems {
			if m, ok := item.(map[string]interface{}); ok {
				// Check if this is a nested group
				if _, hasAnd := m["and"]; hasAnd {
					nestedParams := ks.convertSearchMapToParams(m)
					andParams = append(andParams, nestedParams...)
				} else if _, hasOr := m["or"]; hasOr {
					nestedParams := ks.convertSearchMapToParams(m)
					andParams = append(andParams, nestedParams...)
				} else {
					// This is a flat condition
					andParams = append(andParams, map[string]any{
						"search_column":    m["field"],
						"search_condition": m["operator"],
						"search_text":      fmt.Sprintf("%v", m["value"]),
						"search_operator":  "AND",
					})
				}
			}
		}
		if len(andParams) > 0 {
			result = append(result, map[string]any{
				"and": andParams,
			})
		}
	}

	// Process OR conditions
	if orItems, ok := params["or"].([]interface{}); ok {
		orParams := []map[string]any{}
		for _, item := range orItems {
			if m, ok := item.(map[string]interface{}); ok {
				// Check if this is a nested group
				if _, hasAnd := m["and"]; hasAnd {
					nestedParams := ks.convertSearchMapToParams(m)
					orParams = append(orParams, nestedParams...)
				} else if _, hasOr := m["or"]; hasOr {
					nestedParams := ks.convertSearchMapToParams(m)
					orParams = append(orParams, nestedParams...)
				} else {
					// This is a flat condition
					orParams = append(orParams, map[string]any{
						"search_column":    m["field"],
						"search_condition": m["operator"],
						"search_text":      fmt.Sprintf("%v", m["value"]),
						"search_operator":  "OR",
					})
				}
			}
		}
		if len(orParams) > 0 {
			result = append(result, map[string]any{
				"or": orParams,
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
			v.Search(ks.SearchParams)
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
