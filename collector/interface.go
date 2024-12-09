package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Collector is the interface a collector has to implement.
type Collector interface {
	Update(ch chan<- prometheus.Metric) error // Update get new metrics and expose them via prometheus registry.
}
