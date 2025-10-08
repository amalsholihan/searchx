package test

import (
	"testing"

	"github.com/amalsholihan/searchx"
)

func TestSetDB(t *testing.T) {
	db := SetupTestDB(t)

	search_result := searchx.SetDB(*db).Get()

	if search_result.Raw != "SELECT * FROM `test_user`" {
		t.Fatalf("raw query different : %v", search_result.Raw)
	}

	if search_result.RawAgg != "select count(*) as agg from test_user" {
		t.Fatalf("raw agg query different : %v", search_result.RawAgg)
	}
}
