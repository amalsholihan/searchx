package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/amalsholihan/searchx"
)

func TestSortWithJSON(t *testing.T) {
	db := SetupTestDB(t)
	result := []map[string]any{}

	// Test sort with map[string]interface{}
	sortParams := map[string]interface{}{
		"sort": []interface{}{
			map[string]interface{}{
				"field":     "name",
				"direction": "asc",
			},
			map[string]interface{}{
				"field":     "id",
				"direction": "desc",
			},
		},
	}

	search_result := searchx.SetDB(*db).Sort(sortParams).Get(&result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	if len(result) != 2 {
		t.Fatalf("data length not 2 : %v", len(result))
	}

	// Check if sorted by name ASC
	if result[0]["name"] != "Amal" {
		t.Fatalf("first result should be Amal, got: %v", result[0]["name"])
	}

	// Check if raw query contains ORDER BY (case-insensitive)
	if !strings.Contains(strings.ToUpper(search_result.Raw), "ORDER BY") {
		t.Fatalf("raw query should contain ORDER BY: %v", search_result.Raw)
	}
}

func TestSortWithJSONSingle(t *testing.T) {
	db := SetupTestDB(t)
	result := []map[string]any{}

	// Test single sort with map[string]interface{}
	sortParams := map[string]interface{}{
		"sort": []interface{}{
			map[string]interface{}{
				"field":     "age",
				"direction": "desc",
			},
		},
	}

	search_result := searchx.SetDB(*db).Sort(sortParams).Get(&result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	fmt.Printf("%s", search_result.Raw)
	if len(result) != 2 {
		t.Fatalf("data length not 2 : %v", len(result))
	}

	// Check if sorted by age DESC (Amal with age 34 should be first)
	if result[0]["name"] != "Amal" {
		t.Fatalf("first result should be Amal (age 34), got: %v", result[0]["name"])
	}
}

func TestSortWithNestedSortKey(t *testing.T) {
	db := SetupTestDB(t)
	result := []map[string]any{}

	// Test nested sort with "sort" key
	search_result := searchx.SetDB(*db).Sort([]map[string]any{
		{
			"sort_column": "name",
			"sort_type":   "asc",
		},
		{
			"sort": []map[string]any{
				{
					"sort_column": "id",
					"sort_type":   "desc",
				},
			},
		},
	}).Get(&result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	fmt.Printf("%s", search_result.Raw)
	if len(result) != 2 {
		t.Fatalf("data length not 2 : %v", len(result))
	}

	// Check if sorted by name ASC first
	if result[0]["name"] != "Amal" {
		t.Fatalf("first result should be Amal, got: %v", result[0]["name"])
	}

	// Check if raw query contains ORDER BY (case-insensitive)
	if !strings.Contains(strings.ToUpper(search_result.Raw), "ORDER BY") {
		t.Fatalf("raw query should contain ORDER BY: %v", search_result.Raw)
	}
}

func TestSortWithDeeplyNestedSortKey(t *testing.T) {
	db := SetupTestDB(t)
	result := []map[string]any{}

	// Test deeply nested sort with "sort" key
	search_result := searchx.SetDB(*db).Sort([]map[string]any{
		{
			"sort_column": "age",
			"sort_type":   "desc",
		},
		{
			"sort": []map[string]any{
				{
					"sort_column": "name",
					"sort_type":   "asc",
				},
				{
					"sort": []map[string]any{
						{
							"sort_column": "id",
							"sort_type":   "desc",
						},
					},
				},
			},
		},
	}).Get(&result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	fmt.Printf("%s", search_result.Raw)
	if len(result) != 2 {
		t.Fatalf("data length not 2 : %v", len(result))
	}

	// Check if sorted by age DESC (Amal with age 34 should be first)
	if result[0]["name"] != "Amal" {
		t.Fatalf("first result should be Amal (age 34), got: %v", result[0]["name"])
	}

	// Check if raw query contains ORDER BY (case-insensitive)
	if !strings.Contains(strings.ToUpper(search_result.Raw), "ORDER BY") {
		t.Fatalf("raw query should contain ORDER BY: %v", search_result.Raw)
	}
}

func TestSortWithFieldAndDirection(t *testing.T) {
	db := SetupTestDB(t)
	result := []map[string]any{}

	// Test sort with field and direction keys directly (new format)
	search_result := searchx.SetDB(*db).Sort([]map[string]any{
		{
			"field":     "name",
			"direction": "asc",
		},
		{
			"field":     "id",
			"direction": "desc",
		},
	}).Get(&result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	fmt.Printf("%s", search_result.Raw)
	if len(result) != 2 {
		t.Fatalf("data length not 2 : %v", len(result))
	}

	// Check if sorted by name ASC
	if result[0]["name"] != "Amal" {
		t.Fatalf("first result should be Amal, got: %v", result[0]["name"])
	}

	// Check if raw query contains ORDER BY (case-insensitive)
	if !strings.Contains(strings.ToUpper(search_result.Raw), "ORDER BY") {
		t.Fatalf("raw query should contain ORDER BY: %v", search_result.Raw)
	}
}

func TestSortWithNestedFieldAndDirection(t *testing.T) {
	db := SetupTestDB(t)
	result := []map[string]any{}

	// Test nested sort with field and direction keys
	search_result := searchx.SetDB(*db).Sort([]map[string]any{
		{
			"field":     "age",
			"direction": "desc",
		},
		{
			"sort": []map[string]any{
				{
					"field":     "name",
					"direction": "asc",
				},
				{
					"sort": []map[string]any{
						{
							"field":     "id",
							"direction": "desc",
						},
					},
				},
			},
		},
	}).Get(&result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	fmt.Printf("%s", search_result.Raw)
	if len(result) != 2 {
		t.Fatalf("data length not 2 : %v", len(result))
	}

	// Check if sorted by age DESC (Amal with age 34 should be first)
	if result[0]["name"] != "Amal" {
		t.Fatalf("first result should be Amal (age 34), got: %v", result[0]["name"])
	}

	// Check if raw query contains ORDER BY (case-insensitive)
	if !strings.Contains(strings.ToUpper(search_result.Raw), "ORDER BY") {
		t.Fatalf("raw query should contain ORDER BY: %v", search_result.Raw)
	}
}

func TestSortWithMixedFormats(t *testing.T) {
	db := SetupTestDB(t)
	result := []map[string]any{}

	// Test mixed format: old format (sort_column, sort_type) and new format (field, direction)
	search_result := searchx.SetDB(*db).Sort([]map[string]any{
		{
			"sort_column": "age",
			"sort_type":   "desc",
		},
		{
			"field":     "name",
			"direction": "asc",
		},
		{
			"sort_column": "id",
			"sort_type":   "desc",
		},
	}).Get(&result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	fmt.Printf("%s", search_result.Raw)
	if len(result) != 2 {
		t.Fatalf("data length not 2 : %v", len(result))
	}

	// Check if sorted by age DESC (Amal with age 34 should be first)
	if result[0]["name"] != "Amal" {
		t.Fatalf("first result should be Amal (age 34), got: %v", result[0]["name"])
	}

	// Check if raw query contains ORDER BY (case-insensitive)
	if !strings.Contains(strings.ToUpper(search_result.Raw), "ORDER BY") {
		t.Fatalf("raw query should contain ORDER BY: %v", search_result.Raw)
	}
}
