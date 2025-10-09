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

	if len(ks.Unions) == 0 {
		return ks
	}

	union := ks.Parsed
	for _, v_union := range ks.Unions {
		v_union.Calc()

		sel := v_union.Parsed.(*sqlparser.Select)

		union = &sqlparser.Union{
			Left:  union,
			Right: sel,
			Type:  sqlparser.UnionStr, // atau sqlparser.UnionDistinct
		}
	}

	stmt, err := sqlparser.Parse("SELECT * FROM (" + sqlparser.String(union) + ") as my_table")
	if err != nil {
		ks.Err = fmt.Errorf("parse error: %w  "+"SELECT * FROM ("+sqlparser.String(ks.UnionParsed)+") as my_table", err)
		return ks
	}

	ks.UnionParsed = stmt.(*sqlparser.Select)
	ks.RawUnion = sqlparser.String(stmt)

	return ks
}
