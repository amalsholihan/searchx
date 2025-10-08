package searchx

import (
	"github.com/xwb1989/sqlparser"
	"gorm.io/gorm"
)

type Searchx struct {
	DB             *gorm.DB
	Parsed         sqlparser.SelectStatement
	UnionParsed    sqlparser.SelectStatement
	MappingSelect  map[string]string
	Raw            string
	RawAgg         string
	RawCurrentPage string
	Unions         []*Searchx
	RawUnion       string
	SearchParams   []map[string]string
	Err            error
}
