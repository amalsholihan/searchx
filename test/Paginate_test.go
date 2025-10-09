package test

import (
	"testing"

	"github.com/amalsholihan/searchx"
)

func TestPaginate(t *testing.T) {
	db := SetupTestDB(t)
	result := map[string]any{}
	search_result := searchx.SetDB(*db).Paginate(1, 10, &result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	if result["total"] != 2 {
		t.Fatalf("total data not match 2 : %v", result["total"])
	}
}
