package collector

import (
	"log/slog"
	"math/rand/v2"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/rea1shane/exporter/collector"
	"github.com/rea1shane/exporter/metric"
)

const (
	bSubsystem = "b"
)

func init() {
	collector.RegisterCollector("b", collector.DefaultEnabled, newCollectorB)
}

type b struct {
	logger *slog.Logger
	m1     metric.TypedDesc
}

func newCollectorB(namespace string, logger *slog.Logger) (collector.Collector, error) {
	return &b{
		logger: logger,
		m1: metric.TypedDesc{
			Desc: prometheus.NewDesc(prometheus.BuildFQName(namespace, bSubsystem, "m1"),
				"This is b-m1", []string{"key_x"}, nil),
			ValueType: prometheus.CounterValue,
		},
	}, nil
}

func (c b) Update(ch chan<- prometheus.Metric) error {
	c.logger.Info("Updating collector b")
	c.m1.PushMetric(ch, rand.Float64(), "x")
	return nil
}
