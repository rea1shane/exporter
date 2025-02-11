package collector

import (
	"log/slog"
	"math/rand/v2"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/rea1shane/exporter/collector"
	"github.com/rea1shane/exporter/metric"
)

const (
	aSubsystem = "a"
)

func init() {
	collector.RegisterCollector("a", collector.DefaultEnabled, newCollectorA)
}

type a struct {
	logger *slog.Logger
	m1     metric.TypedDesc
	m2     metric.TypedDesc
}

func newCollectorA(namespace string, logger *slog.Logger) (collector.Collector, error) {
	return &a{
		logger: logger,
		m1: metric.TypedDesc{
			Desc: prometheus.NewDesc(prometheus.BuildFQName(namespace, aSubsystem, "m1"),
				"This is a-m1", []string{"key_x"}, nil),
			ValueType: prometheus.GaugeValue,
		},
		m2: metric.TypedDesc{
			Desc: prometheus.NewDesc(prometheus.BuildFQName(namespace, aSubsystem, "m2"),
				"This is a-m2", []string{"key_x", "key_y"}, map[string]string{"foo": "bar"}),
			ValueType: prometheus.CounterValue,
		},
	}, nil
}

var (
	photoneraAddress = kingpin.Flag(
		"collector.a.test-flag",
		"A user-defined command flag.",
	).Default("undefined").String()
)

func (c a) Update(ch chan<- prometheus.Metric) error {
	c.m1.PushMetric(ch, rand.Float64(), "m")
	c.m2.PushMetric(ch, rand.Float64(), "m", "n")
	c.logger.Info(*photoneraAddress)
	return nil
}
