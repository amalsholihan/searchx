package test

import (
	"testing"

	"github.com/amalsholihan/searchx"
)

func TestPaginate(t *testing.T) {
	db := SetupTestDB(t)
	result := searchx.Paginated{}

	search_result := searchx.SetDB(*db).
		Search([]map[string]any{}).
		Paginate(1, 10, &result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	if len(result.Data) != 2 {
		t.Fatalf("data length not 2 : %v", len(result.Data))
	}
}
