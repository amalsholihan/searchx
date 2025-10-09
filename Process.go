package searchx

import (
	"fmt"
	"math"
	"strconv"

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

func (ks *Searchx) Paginate(page, per_page int, result *map[string]any) *Searchx {
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

	aggStr := fmt.Sprintf("%v", total["agg"])
	aggFloat, _ := strconv.ParseFloat(aggStr, 64)

	data := []map[string]any{}
	ks.DB.Session(&gorm.Session{}).Raw(ks.RawCurrentPage).Find(&data)

	result_data := *result
	result_data["total"] = total["agg"]
	result_data["data"] = data
	result_data["page"] = page
	result_data["per_page"] = per_page
	result_data["total_pages"] = int(math.Ceil(float64(aggFloat) / float64(per_page)))

	return ks
}
