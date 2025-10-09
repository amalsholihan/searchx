package searchx

import (
	"reflect"
	"strconv"
)

func ConvertToFloat(data interface{}) float64 {

	switch data.(type) {
	default:
		panic("Param " + reflect.TypeOf(data).String() + " to float undefined")
	case string:
		i, _ := strconv.ParseFloat(data.(string), 10)
		return float64(i)
	case int:
		return float64(data.(int))
	case int32:
		return float64(data.(int32))
	case int64:
		return float64(data.(int64))
	case uint32:
		return float64(data.(uint32))
	case float64:
		return data.(float64)
	}
}
