package test

import (
	"testing"

	"github.com/amalsholihan/searchx"
)

func TestSummary(t *testing.T) {
	db := SetupTestDB(t)
	result := map[string]any{}
	search_result := searchx.SetDB(*db).Summary(map[string]string{
		"total_sales": "sum(sales)",
	}).GetSummary(&result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	if search_result.RawSummary != "SELECT sum(sales) as total_sales FROM (select id, name, age, sales from test_user) my_table_summary" {
		t.Fatalf("raw query different : %v", search_result.Raw)
	}

	if result["total_sales"].(float64) != 600000 {
		t.Fatalf("total sales not 600000 : %v", result["total_sales"])
	}
}
