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

	if search_result.Raw != "SELECT id, name, age, sales FROM test_user" {
		t.Fatalf("raw query different : %v", search_result.Raw)
	}

	if len(result) != 2 {
		t.Fatalf("data length not 2 : %v", len(result))
	}
}
