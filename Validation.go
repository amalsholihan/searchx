package searchx

import "strings"

func (ks *Searchx) ValidateColumn(column string) string {
	if val, ok := ks.MappingSelect[column]; ok {
		return val
	}
	return ""
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
		">=":          true,
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
