package test

import (
	"fmt"
	"testing"

	"github.com/amalsholihan/searchx"
)

func TestSearchOnly(t *testing.T) {
	db := SetupTestDB(t)
	result := []map[string]any{}
	search_result := searchx.SetDB(*db).Search([]map[string]any{
		{
			"search_column":    "name",
			"search_condition": "=",
			"search_text":      "Annissa",
		},
	}).Get(&result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	fmt.Printf("%s", search_result.Raw)
	if result[0]["name"] != "Annissa" {
		t.Fatalf("result different : %v", result[0]["name"])
	}

	if len(result) != 1 {
		t.Fatalf("data length not 1 : %v", len(result))
	}
}

func TestSearchWithJSON(t *testing.T) {
	db := SetupTestDB(t)
	result := []map[string]any{}

	// Test simple AND condition
	jsonStr := `{
		"search": {
			"and": [
				{ "field": "name", "operator": "=", "value": "Annissa" }
			]
		}
	}`

	search_result := searchx.SetDB(*db).SearchWithJSON(jsonStr).Get(&result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	if len(result) != 1 {
		t.Fatalf("data length not 1 : %v", len(result))
	}

	if result[0]["name"] != "Annissa" {
		t.Fatalf("result different : %v", result[0]["name"])
	}
}

func TestSearchWithJSONNestedOR(t *testing.T) {
	db := SetupTestDB(t)
	result := []map[string]any{}

	// Test nested OR condition
	jsonStr := `{
		"search": {
			"and": [
				{ "field": "age", "operator": ">=", "value": "32" },
				{
					"or": [
						{ "field": "name", "operator": "like", "value": "Amal" },
						{ "field": "name", "operator": "like", "value": "Annissa" }
					]
				}
			],
			"and": [
				{ "field": "sales", "operator": ">", "value": "1000" }
			]
		}
	}`

	search_result := searchx.SetDB(*db).SearchWithJSON(jsonStr).Get(&result)

	fmt.Printf("%s", search_result.Raw)
	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	if len(result) != 2 {
		t.Fatalf("data length not 2 : %v", len(result))
	}

	// Check that both Amal and Annissa are in results
	names := make(map[string]bool)
	for _, r := range result {
		names[r["name"].(string)] = true
	}

	if !names["Amal"] || !names["Annissa"] {
		t.Fatalf("expected Amal and Annissa in results, got: %v", names)
	}
}

func TestSearchWithJSONMultipleOR(t *testing.T) {
	db := SetupTestDB(t)
	result := []map[string]any{}

	// Test multiple OR conditions
	jsonStr := `{
		"search": {
			"or": [
				{ "field": "name", "operator": "=", "value": "Amal" },
				{ "field": "name", "operator": "=", "value": "Annissa" }
			]
		}
	}`

	search_result := searchx.SetDB(*db).SearchWithJSON(jsonStr).Get(&result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	if len(result) != 2 {
		t.Fatalf("data length not 2 : %v", len(result))
	}
}

func TestSearchWithNestedSearchKey(t *testing.T) {
	db := SetupTestDB(t)
	result := []map[string]any{}

	// Test nested search with "search" key
	search_result := searchx.SetDB(*db).Search([]map[string]any{
		{
			"search_column":    "name",
			"search_condition": "=",
			"search_text":      "Annissa",
		},
		{
			"search": []map[string]any{
				{
					"search_column":    "name",
					"search_condition": "=",
					"search_text":      "Annissa",
				},
			},
		},
	}).Get(&result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	fmt.Printf("%s", search_result.Raw)
	if len(result) != 1 {
		t.Fatalf("data length not 1 : %v", len(result))
	}

	if result[0]["name"] != "Annissa" {
		t.Fatalf("result different : %v", result[0]["name"])
	}
}

func TestSearchWithDeeplyNestedSearchKey(t *testing.T) {
	db := SetupTestDB(t)
	result := []map[string]any{}

	// Test deeply nested search with "search" key
	// This tests that nested "search" keys work correctly
	search_result := searchx.SetDB(*db).Search([]map[string]any{
		{
			"search_column":    "age",
			"search_condition": ">=",
			"search_text":      "32",
		},
		{
			"search": []map[string]any{
				{
					"search_column":    "name",
					"search_condition": "like",
					"search_text":      "Amal",
				},
			},
		},
	}).Get(&result)

	if search_result.Err != nil {
		t.Fatal(search_result.Err)
	}

	fmt.Printf("%s", search_result.Raw)
	// Should return Amal (age >= 32 and name like 'Amal')
	if len(result) != 1 {
		t.Fatalf("data length not 1 : %v", len(result))
	}

	if result[0]["name"] != "Amal" {
		t.Fatalf("result different : %v", result[0]["name"])
	}
}
