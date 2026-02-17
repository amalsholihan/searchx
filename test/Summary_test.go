package test

import (
	"fmt"
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

	// Check result - handle both float64 and string types from different databases
	totalSales := result["total_sales"]
	var sales float64
	switch v := totalSales.(type) {
	case float64:
		sales = v
	case string:
		fmt.Sscanf(v, "%f", &sales)
	default:
		t.Fatalf("unexpected type: %T", v)
	}

	if sales != 600000 {
		t.Fatalf("total sales not 600000 : %v", sales)
	}
}
