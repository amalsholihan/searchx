package searchx

import (
	"fmt"
	"strings"
)

func (ks *Searchx) Search(params []map[string]string) *Searchx {
	ks.SearchParams = params
	return ks
}

func (ks *Searchx) ProcessSearch() *Searchx {
	params := ks.SearchParams
	query := ks.DB
	var havingClauses []string
	var havingArgs []interface{}

	for i, p := range params {
		val := p["search_text"]
		col := ks.ValidateColumn(p["search_column"])
		if col == "" {
			ks.Err = fmt.Errorf("column %v not found in select statement", p["search_column"])
			return ks
		}
		op := ks.ValidateOperator(p["search_operator"])
		if op == "" {
			ks.Err = fmt.Errorf("operator %v not allowed", p["search_operator"])
			return ks
		}
		cond := ks.ValidateCondition(p["search_condition"])
		if cond == "" {
			ks.Err = fmt.Errorf("condition %v not allowed", p["search_condition"])
			return ks
		}

		if col == "" {
			continue
		}
		if op != "or" {
			op = "and"
		}

		q := ""
		var args []interface{}

		switch cond {
		case "=", ">=", "<=":
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
			if i > 0 {
				havingClauses = append(havingClauses, strings.ToUpper(op))
			}
			havingClauses = append(havingClauses, q)
			havingArgs = append(havingArgs, args...)
		} else {
			// kondisi biasa -> tetap pakai WHERE
			if i == 0 || op == "and" {
				query = query.Where(q, args...)
			} else {
				query = query.Or(q, args...)
			}
		}
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
	ks.Parse()

	return ks
}
