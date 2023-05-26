package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rea1shane/exporter"
	"github.com/sirupsen/logrus"
)

const (
	bNamespace = "special"
	bSubsystem = "module_b"
)

func init() {
	exporter.RegisterCollector("b", exporter.DefaultEnabled, newCollectorB)
}

type b struct {
	logger *logrus.Entry
	m1     exporter.TypedDesc
}

func newCollectorB(_ string, logger *logrus.Entry) (exporter.Collector, error) {
	return &b{
		logger: logger,
		// b's m1 use special namespace
		m1: exporter.TypedDesc{
			Desc: prometheus.NewDesc(prometheus.BuildFQName(bNamespace, bSubsystem, "m1"),
				"This is b-m1", labelList1, nil),
			ValueType: prometheus.CounterValue,
		},
	}, nil
}

func (c b) Update(ch chan<- prometheus.Metric) error {
	exporter.PushTypedDesc(ch, c.m1, 3, "value_x")
	return nil
}
