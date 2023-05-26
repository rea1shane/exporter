package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rea1shane/exporter"
	"github.com/sirupsen/logrus"
)

const (
	aSubsystem = "module_a"
)

func init() {
	exporter.RegisterCollector("a", exporter.DefaultEnabled, newCollectorA)
}

type a struct {
	logger *logrus.Entry
	m1     exporter.TypedDesc
	m2     exporter.TypedDesc
}

func newCollectorA(namespace string, logger *logrus.Entry) (exporter.Collector, error) {
	return &a{
		logger: logger,
		// a's m1 & m2 use common namespace
		m1: exporter.TypedDesc{
			Desc: prometheus.NewDesc(prometheus.BuildFQName(namespace, aSubsystem, "m1"),
				"This is a-m1", labelList1, constLabels),
			ValueType: prometheus.GaugeValue,
		},
		m2: exporter.TypedDesc{
			Desc: prometheus.NewDesc(prometheus.BuildFQName(namespace, aSubsystem, "m2"),
				"This is a-m2", labelList2, nil),
			ValueType: prometheus.CounterValue,
		},
	}, nil
}

func (c a) Update(ch chan<- prometheus.Metric) error {
	exporter.PushTypedDesc(ch, c.m1, 1, "value_a")
	exporter.PushTypedDesc(ch, c.m2, 2, "value_a", "value_b")
	return nil
}
