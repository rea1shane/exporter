package util

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	DefaultEnabled  = true
	DefaultDisabled = false
)

// TypedDesc is suggestion metric's type
type TypedDesc struct {
	Desc      *prometheus.Desc
	ValueType prometheus.ValueType
}

func (t *TypedDesc) MustNewConstMetric(value float64, labels ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(t.Desc, t.ValueType, value, labels...)
}
