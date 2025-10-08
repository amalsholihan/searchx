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
	ks.Raw = query
	return ks
}
