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
		"and": true,
		"or":  true,
	}
	if _, ok := allowed_operator[strings.ToLower(operator)]; ok {
		return strings.ToLower(operator)
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
