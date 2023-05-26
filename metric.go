package exporter

import "github.com/prometheus/client_golang/prometheus"

// TypedDesc contains metric's necessary information
type TypedDesc struct {
	Desc      *prometheus.Desc
	ValueType prometheus.ValueType
}

func (d *TypedDesc) MustNewConstMetric(value float64, labels ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(d.Desc, d.ValueType, value, labels...)
}

// PushTypedDesc is PushMetric TypedDesc version.
func PushTypedDesc(ch chan<- prometheus.Metric, d TypedDesc, value interface{}, labelValues ...string) {
	PushMetric(ch, d.Desc, d.ValueType, value, labelValues...)
}

// PushMetric helps construct and convert a variety of value types into Prometheus float64 metrics.
func PushMetric(ch chan<- prometheus.Metric, fieldDesc *prometheus.Desc, valueType prometheus.ValueType, value interface{}, labelValues ...string) {
	var fVal float64
	switch val := value.(type) {
	case int:
		fVal = float64(val)
	case int8:
		fVal = float64(val)
	case int16:
		fVal = float64(val)
	case int32:
		fVal = float64(val)
	case int64:
		fVal = float64(val)
	case uint:
		fVal = float64(val)
	case uint8:
		fVal = float64(val)
	case uint16:
		fVal = float64(val)
	case uint32:
		fVal = float64(val)
	case uint64:
		fVal = float64(val)
	case uintptr:
		fVal = float64(val)
	case float32:
		fVal = float64(val)
	case float64:
		fVal = val

	case *int:
		if val == nil {
			return
		}
		fVal = float64(*val)
	case *int8:
		if val == nil {
			return
		}
		fVal = float64(*val)
	case *int16:
		if val == nil {
			return
		}
		fVal = float64(*val)
	case *int32:
		if val == nil {
			return
		}
		fVal = float64(*val)
	case *int64:
		if val == nil {
			return
		}
		fVal = float64(*val)
	case *uint:
		if val == nil {
			return
		}
		fVal = float64(*val)
	case *uint8:
		if val == nil {
			return
		}
		fVal = float64(*val)
	case *uint16:
		if val == nil {
			return
		}
		fVal = float64(*val)
	case *uint32:
		if val == nil {
			return
		}
		fVal = float64(*val)
	case *uint64:
		if val == nil {
			return
		}
		fVal = float64(*val)
	case *uintptr:
		if val == nil {
			return
		}
		fVal = float64(*val)
	case *float32:
		if val == nil {
			return
		}
		fVal = float64(*val)
	case *float64:
		if val == nil {
			return
		}
		fVal = *val

	default:
		return
	}

	ch <- prometheus.MustNewConstMetric(fieldDesc, valueType, fVal, labelValues...)
}
