package test

import (
	"testing"

	"github.com/amalsholihan/searchx"
	"gorm.io/gorm"
)

func TestGet(t *testing.T) {
	db := SetupTestDB(t)
	result := []map[string]any{}
	search_result := searchx.SetDB(*db).Get(&result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	if search_result.Raw != "SELECT id, name, age, sales FROM `test_user`" {
		t.Fatalf("raw query different : %v", search_result.Raw)
	}

	if len(result) != 2 {
		t.Fatalf("data length not 2 : %v", len(result))
	}
}

func TestSearch(t *testing.T) {
	db := SetupTestDB(t)
	result := []map[string]any{}
	search_result := searchx.SetDB(*db).Search([]map[string]string{
		{
			"search_column":    "name",
			"search_condition": "=",
			"search_text":      "Annissa",
		},
	}).Get(&result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	if result[0]["name"] != "Annissa" {
		t.Fatalf("result different : %v", result[0]["name"])
	}

	if len(result) != 1 {
		t.Fatalf("data length not 1 : %v", len(result))
	}
}

func TestSearchWithUnionSort(t *testing.T) {
	db := SetupTestDB(t)
	result := []map[string]any{}

	q_staff := db.Session(&gorm.Session{}).Model(&Staff{}).Select("id, name, age, sales")
	search_result := searchx.SetDB(*db).
		Union(*searchx.SetDB(*q_staff)).
		Search([]map[string]string{
			{
				"search_column":    "name",
				"search_condition": "is not null",
			},
		}).
		Sort([]map[string]string{
			{
				"sort_column": "name",
				"sort_type":   "asc",
			},
			{
				"sort_column": "id",
				"sort_type":   "desc",
			},
		}).
		Get(&result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	if search_result.RawUnion != "select * from (select id, name, age, sales from test_user where name is not null union select id, name, age, sales from test_staff where name is not null) as my_table order by name ASC, id DESC" {
		t.Fatalf("query raw union different : %v", search_result.RawUnion)
	}

	if result[0]["name"] != "Amal" {
		t.Fatalf("result different : %v", result[0]["name"])
	}

	if len(result) != 3 {
		t.Fatalf("data length not 3 : %v", len(result))
	}
}
