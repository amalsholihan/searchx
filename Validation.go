package searchx

import "strings"

func (ks *Searchx) ValidateColumn(column string) string {
	val, ok := ks.MappingSelect[column]
	if !ok {
		return ""
	}
	if ks.RawUnion != "" {
		// pgProcessUnion sudah wrap kedua sisi jadi "SELECT * FROM (...) AS my_table"
		// sebelum ProcessSort jalan — di titik ini cuma alias output yang keliatan,
		// bukan ekspresi/tabel aslinya (lihat PostgresParser.go pgProcessUnion).
		return column
	}
	return val
}

func (ks *Searchx) ValidateOperator(operator string) string {
	allowed_operator := map[string]interface{}{
		"AND": true,
		"OR":  true,
	}
	if _, ok := allowed_operator[strings.ToUpper(operator)]; ok {
		return strings.ToUpper(operator)
	}
	return ""
}

func (ks *Searchx) ValidateSortType(sortType string) string {
	allowed_operator := map[string]interface{}{
		"ASC":  true,
		"DESC": true,
	}
	if _, ok := allowed_operator[strings.ToUpper(sortType)]; ok {
		return strings.ToUpper(sortType)
	}
	return ""
}

func (ks *Searchx) ValidateCondition(condition string) string {
	allowed_condition := map[string]interface{}{
		"=":           true,
		">":           true,
		">=":          true,
		"<":           true,
		"<=":          true,
		"!=":          true,
		"like":        true,
		"is not null": true,
		"is null":     true,
	}
	if _, ok := allowed_condition[strings.ToLower(condition)]; ok {
		return strings.ToLower(condition)
	}
	return ""
}
