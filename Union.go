package searchx

import (
	"fmt"

	"github.com/xwb1989/sqlparser"
)

func (ks *Searchx) Union(uks Searchx) *Searchx {
	ks.Unions = append(ks.Unions, &uks)
	return ks
}

func (ks *Searchx) ProcessUnion() *Searchx {
	if ks.isPostgres() {
		return ks.pgProcessUnion()
	}

	if len(ks.Unions) == 0 {
		return ks
	}

	// Jika ada error sebelumnya, return langsung
	if ks.Err != nil {
		return ks
	}

	// Pastikan query utama adalah SELECT
	if ks.Parsed == nil {
		ks.Err = fmt.Errorf("main query is not parsed")
		return ks
	}

	union := ks.Parsed
	for i, v_union := range ks.Unions {
		v_union.Calc()

		// Cek error dari union calculation
		if v_union.Err != nil {
			ks.Err = fmt.Errorf("union %d failed: %w", i, v_union.Err)
			return ks
		}

		sel, ok := v_union.Parsed.(*sqlparser.Select)
		if !ok {
			ks.Err = fmt.Errorf("union %d is not SELECT statement, got: %T", i, v_union.Parsed)
			return ks
		}

		union = &sqlparser.Union{
			Left:  union,
			Right: sel,
			Type:  sqlparser.UnionStr, // atau sqlparser.UnionDistinct
		}
	}

	unionQuery := "SELECT * FROM (" + sqlparser.String(union) + ") as my_table"
	stmt, err := sqlparser.Parse(unionQuery)
	if err != nil {
		ks.Err = fmt.Errorf("parse union query failed: %w (query: %s)", err, unionQuery)
		return ks
	}

	sel, ok := stmt.(*sqlparser.Select)
	if !ok {
		ks.Err = fmt.Errorf("union query result is not SELECT, got: %T", stmt)
		return ks
	}

	ks.UnionParsed = sel
	ks.RawUnion = ks.normalizeQuery(sqlparser.String(stmt))

	return ks
}
