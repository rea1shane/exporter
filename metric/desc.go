package metric

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/rea1shane/exporter/util"
)

type TypedDesc struct {
	Desc      *prometheus.Desc
	ValueType prometheus.ValueType
}

// PushMetric helps construct and convert a variety of value types into Prometheus float64 metrics.
func (d *TypedDesc) PushMetric(ch chan<- prometheus.Metric, value any, labelValues ...string) {
	fVal, err := util.AnyToFloat64(value)
	if err != nil {
		// TODO handler error
		return
	}

	ch <- prometheus.MustNewConstMetric(d.Desc, d.ValueType, fVal, labelValues...)
}
