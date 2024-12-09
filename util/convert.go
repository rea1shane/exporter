package util

import (
	"fmt"
)

func AnyToFloat64(data any) (float64, error) {
	var f float64
	switch d := data.(type) {
	case uint8:
		f = float64(d)
	case uint16:
		f = float64(d)
	case uint32:
		f = float64(d)
	case uint64:
		f = float64(d)
	case int64:
		f = float64(d)
	case *uint8:
		if d == nil {
			return 0, fmt.Errorf("can't convert %v to float64", data)
		}
		f = float64(*d)
	case *uint16:
		if d == nil {
			return 0, fmt.Errorf("can't convert %v to float64", data)
		}
		f = float64(*d)
	case *uint32:
		if d == nil {
			return 0, fmt.Errorf("can't convert %v to float64", data)
		}
		f = float64(*d)
	case *uint64:
		if d == nil {
			return 0, fmt.Errorf("can't convert %v to float64", data)
		}
		f = float64(*d)
	case *int64:
		if d == nil {
			return 0, fmt.Errorf("can't convert %v to float64", data)
		}
		f = float64(*d)
	default:
		return 0, fmt.Errorf("can't convert %v to float64", data)
	}

	return f, nil
}
