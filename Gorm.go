package searchx

import (
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

// SetDB simpan *gorm.DB ke context
func SetDB(db gorm.DB) *Searchx {
	ks := Searchx{
		DB:      &db,
		Dialect: db.Dialector.Name(),
	}
	return &ks
}

// SetDB simpan *gorm.DB ke context
func (ks *Searchx) SetDB(db gorm.DB) *Searchx {
	ks.DB = &db
	ks.Dialect = db.Dialector.Name()
	return ks
}

func (ks *Searchx) isPostgres() bool {
	return ks.Dialect == "postgres"
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
	// Replace PostgreSQL-style placeholders ($1, $2, ...) with ?
	re := regexp.MustCompile(`\$\d+`)
	query = re.ReplaceAllString(query, "?")

	for _, v := range vars {
		var val string
		switch t := v.(type) {
		case string:
			val = fmt.Sprintf("'%s'", ks.escapeSQLString(t))
		default:
			val = fmt.Sprintf("%v", t)
		}
		query = strings.Replace(query, "?", val, 1)
	}
	// Normalize quoted identifiers for sqlparser compatibility (MySQL/SQLite only).
	// PostgreSQL uses double-quotes to preserve identifier case — must not strip them.
	query = strings.ReplaceAll(query, "`", "") // MySQL backticks
	if !ks.isPostgres() {
		query = strings.ReplaceAll(query, "\"", "") // strip for sqlparser (MySQL/SQLite)
		query = strings.ReplaceAll(query, "[", "")  // SQLite brackets
		query = strings.ReplaceAll(query, "]", "")
	}

	// Convert PostgreSQL LIMIT syntax to standard: LIMIT offset, count
	// This is only needed if the query came from GORM with PostgreSQL LIMIT format
	// Actually keep it as-is for now since sqlparser understands both formats

	ks.Raw = query
	return ks
}

// escapeSQLString escapes a Go string for safe inline use as an SQL string literal.
// Interpolate() manually inlines bound values into the final raw SQL text (needed to wrap
// it in derived count/pagination queries) instead of using driver-level bound parameters,
// so this escaping is the only thing standing between user input (e.g. a search_text
// containing an apostrophe, like "A'mal") and a broken query or SQL injection. Doubling `'`
// is the standard-SQL escape (works on both Postgres and MySQL); MySQL additionally treats
// backslash as an escape character in string literals (Postgres does not, in standard-
// conforming strings), so backslashes are doubled first, MySQL-only, before quote-escaping.
func (ks *Searchx) escapeSQLString(s string) string {
	if !ks.isPostgres() {
		s = strings.ReplaceAll(s, `\`, `\\`)
	}
	return strings.ReplaceAll(s, "'", "''")
}
