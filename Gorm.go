package searchx

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// SetDB simpan *gorm.DB ke context
func SetDB(db gorm.DB) *Searchx {
	ks := Searchx{
		DB: &db,
	}
	return &ks
}

// SetDB simpan *gorm.DB ke context
func (ks *Searchx) SetDB(db gorm.DB) *Searchx {
	ks.DB = &db
	return ks
}

// GetDB ambil *gorm.DB dari context
func (ks *Searchx) GetDB() (*gorm.DB, error) {
	if ks.DB == nil {
		return nil, fmt.Errorf("db is empty")
	}
	return ks.DB, nil
}

// SetRawQuery mengembalikan SQL dan parameter dari query GORM
func (ks *Searchx) SetRawQuery() *Searchx {
	db := ks.DB

	stmt := db.Session(&gorm.Session{DryRun: true}).Find(nil).Statement
	return ks.Interpolate(stmt.SQL.String(), stmt.Vars)
}

func (ks *Searchx) Interpolate(query string, vars []interface{}) *Searchx {
	// Convert $N placeholders to ? for compatibility
	paramCount := 1
	for strings.Contains(query, "$") {
		query = strings.Replace(query, "$1", "?", 1)
		paramCount++
	}

	for _, v := range vars {
		var val string
		switch t := v.(type) {
		case string:
			val = fmt.Sprintf("'%s'", t)
		default:
			val = fmt.Sprintf("%v", t)
		}
		query = strings.Replace(query, "?", val, 1)
	}
	// Normalize quoted identifiers for sqlparser compatibility
	query = strings.ReplaceAll(query, "`", "")  // MySQL backticks
	query = strings.ReplaceAll(query, "\"", "") // PostgreSQL double quotes
	query = strings.ReplaceAll(query, "[", "")  // SQLite brackets
	query = strings.ReplaceAll(query, "]", "")

	// Convert PostgreSQL LIMIT syntax to standard: LIMIT offset, count
	// This is only needed if the query came from GORM with PostgreSQL LIMIT format
	// Actually keep it as-is for now since sqlparser understands both formats

	ks.Raw = query
	return ks
}
