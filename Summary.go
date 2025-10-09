package searchx

import (
	"fmt"
	"sort"
	"strings"

	"github.com/xwb1989/sqlparser"
)

func (ks *Searchx) Summary(select_summary map[string]string) *Searchx {
	ks.SelectSummaries = make(map[string]string)
	for k, v := range select_summary {
		ks.SelectSummaries[k] = v
	}
	return ks
}

func (ks *Searchx) ParseSummaryQuery() *Searchx {

	if len(ks.SelectSummaries) == 0 {
		ks.Err = fmt.Errorf("no select summary defined")
		return ks
	}

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

	raw_query := sqlparser.String(&sel)

	// urut biar hasil stabil (opsional)
	keys := make([]string, 0, len(ks.SelectSummaries))
	for k := range ks.SelectSummaries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// gabung jadi "expr as alias"
	var parts []string
	for _, k := range keys {
		v := ks.SelectSummaries[k]
		parts = append(parts, fmt.Sprintf("%s as %s", v, k))
	}

	select_columns := strings.Join(parts, ", ")

	ks.RawSummary = fmt.Sprintf(`SELECT %v FROM (%v) my_table_summary`, select_columns, raw_query)

	return ks
}
