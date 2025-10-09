package searchx

import (
	"reflect"
	"strconv"
)

func ConvertToInt(data interface{}) int {

	if data == nil {
		return 0
	}

	switch data.(type) {
	default:
		panic("Param " + reflect.TypeOf(data).String() + " to int undefined")
	case string:
		i, _ := strconv.ParseInt(data.(string), 10, 64)
		return int(i)
	case int:
		return data.(int)
	case int8:
		return int(data.(int8))
	case int32:
		return int(data.(int32))
	case int64:
		return int(data.(int64))
	case uint32:
		return int(data.(uint32))
	case float64:
		return int(data.(float64))
	}
}

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
