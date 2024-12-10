package util

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
)

// https://stackoverflow.com/questions/20767724/converting-unknown-interface-to-float64-in-golang

var (
	floatType  = reflect.TypeOf(float64(0))
	stringType = reflect.TypeOf("")
)

func AnyToFloat64(data any) (float64, error) {
	switch d := data.(type) {
	case float64:
		return d, nil
	case float32:
		return float64(d), nil
	case int64:
		return float64(d), nil
	case int32:
		return float64(d), nil
	case int:
		return float64(d), nil
	case uint64:
		return float64(d), nil
	case uint32:
		return float64(d), nil
	case uint:
		return float64(d), nil
	case string:
		return strconv.ParseFloat(d, 64)
	default:
		v := reflect.ValueOf(data)
		v = reflect.Indirect(v)
		if v.Type().ConvertibleTo(floatType) {
			fv := v.Convert(floatType)
			return fv.Float(), nil
		} else if v.Type().ConvertibleTo(stringType) {
			sv := v.Convert(stringType)
			s := sv.String()
			return strconv.ParseFloat(s, 64)
		} else {
			return math.NaN(), fmt.Errorf("can't convert %v to float64", v.Type())
		}
	}
}
