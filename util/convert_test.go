package util

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func TestAnyToFloat64(t *testing.T) {
	var m1, m2 runtime.MemStats

	runtime.ReadMemStats(&m1)
	start := time.Now()
	for i := 0; i < 1e6; i++ {
		getFloatReflectOnly(37)
	}
	fmt.Println("Reflect-only, 1e6 runs:")
	fmt.Println("Wall time:", time.Now().Sub(start))
	runtime.ReadMemStats(&m2)
	fmt.Println("Bytes allocated:", m2.TotalAlloc-m1.TotalAlloc)

	runtime.ReadMemStats(&m1)
	start = time.Now()
	for i := 0; i < 1e6; i++ {
		AnyToFloat64(37)
	}
	fmt.Println("\nReflect-and-switch, 1e6 runs:")
	fmt.Println("Wall time:", time.Since(start))
	runtime.ReadMemStats(&m2)
	fmt.Println("Bytes allocated:", m2.TotalAlloc-m1.TotalAlloc)

	runtime.ReadMemStats(&m1)
	start = time.Now()
	for i := 0; i < 1e6; i++ {
		getFloatSwitchOnly(37)
	}
	fmt.Println("\nSwitch only, 1e6 runs:")
	fmt.Println("Wall time:", time.Since(start))
	runtime.ReadMemStats(&m2)
	fmt.Println("Bytes allocated:", m2.TotalAlloc-m1.TotalAlloc)
}

func getFloatReflectOnly(unk interface{}) (float64, error) {
	v := reflect.ValueOf(unk)
	v = reflect.Indirect(v)
	if !v.Type().ConvertibleTo(floatType) {
		return math.NaN(), fmt.Errorf("cannot convert %v to float64", v.Type())
	}
	fv := v.Convert(floatType)
	return fv.Float(), nil
}

var errUnexpectedType = errors.New("non-numeric type could not be converted to float")

func getFloatSwitchOnly(unk interface{}) (float64, error) {
	switch i := unk.(type) {
	case float64:
		return i, nil
	case float32:
		return float64(i), nil
	case int64:
		return float64(i), nil
	case int32:
		return float64(i), nil
	case int:
		return float64(i), nil
	case uint64:
		return float64(i), nil
	case uint32:
		return float64(i), nil
	case uint:
		return float64(i), nil
	default:
		return math.NaN(), errUnexpectedType
	}
}
