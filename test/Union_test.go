package test

import (
	"testing"

	"github.com/amalsholihan/searchx"
	"gorm.io/gorm"
)

func TestGetUnion(t *testing.T) {
	db := SetupTestDB(t)
	result := []map[string]any{}

	q_staff := db.Session(&gorm.Session{}).Model(&Staff{}).Select("id, name, age, sales")
	search_result := searchx.SetDB(*db).Union(*searchx.SetDB(*q_staff)).Get(&result)

	if search_result.RawUnion != "select * from (select id, name, age, sales from test_user union select id, name, age, sales from test_staff) as my_table" {
		t.Fatalf("raw union query different : %v", search_result.RawUnion)
	}

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

}

func TestPaginateUnion(t *testing.T) {
	db := SetupTestDB(t)
	result := searchx.Paginated{}

	q_staff := db.Session(&gorm.Session{}).Model(&Staff{}).Select("id, name, age, sales")
	search_result := searchx.SetDB(*db).Union(*searchx.SetDB(*q_staff)).Paginate(1, 10, &result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	if result.Total != 3 {
		t.Fatalf("total data not match 3 : %v", result.Total)
	}
}
