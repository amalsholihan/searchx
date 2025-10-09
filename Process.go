package searchx

import (
	"fmt"
	"math"

	"gorm.io/gorm"
)

func (ks *Searchx) Calc() *Searchx {
	ks.SetRawQuery()
	ks.Parse()
	ks.ParseSelectMapping()
	ks.ProcessSearch()
	ks.ProcessUnion()
	ks.ProcessSort()

	return ks
}

func (ks *Searchx) Get(result *[]map[string]any) *Searchx {

	if len(ks.SelectSummaries) > 0 {
		result_summary := map[string]any{}
		ks = ks.GetSummary(&result_summary)
		*result = append(*result, result_summary)
		return ks
	}

	ks.Calc()
	query_to_execute := ks.Raw
	if ks.RawUnion != "" {
		query_to_execute = ks.RawUnion
	}

	if ks.Err != nil {
		return ks
	}

	ks.DB.Session(&gorm.Session{}).Raw(query_to_execute).Find(result)

	return ks
}

func (ks *Searchx) GetSummary(result *map[string]any) *Searchx {
	ks.Calc()
	ks.ParseSummaryQuery()

	if ks.Err != nil {
		return ks
	}

	ks.DB.Session(&gorm.Session{}).Raw(ks.RawSummary).Take(result)

	return ks
}

func (ks *Searchx) Paginate(page, per_page int, result *Paginated) *Searchx {
	ks.Calc()
	ks.ParseCountQuery()
	ks.ParseCurrentPageQuery(page, per_page)

	if ks.Err != nil {
		return ks
	}

	total := map[string]any{}
	ks.DB.Session(&gorm.Session{}).Raw(ks.RawAgg).Take(&total)

	if total["agg"] == nil {
		ks.Err = fmt.Errorf("query count failed")
		return ks
	}

	data := []map[string]any{}
	ks.DB.Session(&gorm.Session{}).Raw(ks.RawCurrentPage).Find(&data)

	result.Total = ConvertToInt(total["agg"])
	result.Data = data
	result.Page = page
	result.PerPage = per_page
	result.TotalPages = int(math.Ceil(ConvertToFloat(result.Total) / ConvertToFloat(per_page)))

	return ks
}
