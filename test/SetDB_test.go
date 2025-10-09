package test

import (
	"testing"

	"github.com/amalsholihan/searchx"
)

func TestGet(t *testing.T) {
	db := SetupTestDB(t)
	result := []map[string]any{}
	search_result := searchx.SetDB(*db).Get(&result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	if search_result.Raw != "SELECT * FROM `test_user`" {
		t.Fatalf("raw query different : %v", search_result.Raw)
	}

	if len(result) != 2 {
		t.Fatalf("data length not 2 : %v", len(result))
	}
}

func TestSummary(t *testing.T) {
	db := SetupTestDB(t)
	result := map[string]any{}
	search_result := searchx.SetDB(*db).Summary(map[string]string{
		"total_sales": "sum(sales)",
	}).GetSummary(&result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	if search_result.RawSummary != "SELECT sum(sales) as total_sales FROM (select * from test_user) my_table_summary" {
		t.Fatalf("raw query different : %v", search_result.Raw)
	}

	if result["total_sales"].(float64) != 600000 {
		t.Fatalf("total sales not 600000 : %v", result["total_sales"])
	}
}

func TestPaginate(t *testing.T) {
	db := SetupTestDB(t)
	result := map[string]any{}
	search_result := searchx.SetDB(*db).Paginate(1, 10, &result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	if result["total"].(int64) != 2 {
		t.Fatalf("total data not match 2 : %v", result["total"])
	}

	t.Logf("data : %v", result)
}
