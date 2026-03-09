package test

import (
	"testing"

	"github.com/amalsholihan/searchx"
	"gorm.io/gorm"
)

func TestSearchWithUnionSort(t *testing.T) {
	db := SetupTestDB(t)
	result := []map[string]any{}

	q_staff := db.Session(&gorm.Session{}).Model(&Staff{}).Select("id, name, age, sales")
	search_result := searchx.SetDB(*db).
		Union(*searchx.SetDB(*q_staff)).
		Search([]map[string]any{
			{
				"search_column":    "name",
				"search_condition": "is not null",
			},
		}).
		Sort([]map[string]any{
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

func TestGetUnion(t *testing.T) {
	db := SetupTestDB(t)
	result := []map[string]any{}

	q_staff := db.Session(&gorm.Session{}).Model(&Staff{}).Select("id, name, age, sales")
	search_result := searchx.SetDB(*db).
		Union(*searchx.SetDB(*q_staff)).
		Get(&result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	if search_result.RawUnion != "select * from (select id, name, age, sales from test_user union select id, name, age, sales from test_staff) as my_table" {
		t.Fatalf("query raw union different : %v", search_result.RawUnion)
	}

	if len(result) != 3 {
		t.Fatalf("data length not 3 : %v", len(result))
	}
}
