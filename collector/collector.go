// Package collector includes all individual collectors to gather and export system metrics.
package collector

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/rea1shane/exporter/metric"
)

var (
	initiatedCollectorsMtx = sync.Mutex{}               // initiatedCollectorsMtx avoid thread conflicts
	initiatedCollectors    = make(map[string]Collector) // initiatedCollectors record the collectors that have been initialized in the method NewCollection (To reduce the collector's construction method call)
)

// Collection implements the prometheus.Collector interface.
type Collection struct {
	Collectors         map[string]Collector
	logger             *slog.Logger
	scrapeDurationDesc metric.TypedDesc
	scrapeSuccessDesc  metric.TypedDesc
}

// NewCollection creates a new Collection.
// Namespace defines the common namespace to be used by all metrics.
func NewCollection(exporterName, namespace string, logger *slog.Logger, filters ...string) (*Collection, error) {
	f := make(map[string]bool)
	for _, filter := range filters {
		enabled, exist := collectorState[filter]
		if !exist {
			return nil, fmt.Errorf("missing collector: %s", filter)
		}
		if !*enabled {
			return nil, fmt.Errorf("disabled collector: %s", filter)
		}
		f[filter] = true
	}
	collectors := make(map[string]Collector)
	initiatedCollectorsMtx.Lock()
	defer initiatedCollectorsMtx.Unlock()
	for key, enabled := range collectorState {
		if !*enabled || (len(f) > 0 && !f[key]) {
			continue
		}
		if collector, ok := initiatedCollectors[key]; ok {
			collectors[key] = collector
		} else {
			c, err := factories[key](namespace, logger.With("collector", key))
			if err != nil {
				return nil, err
			}
			collectors[key] = c
			initiatedCollectors[key] = c
		}
	}
	return &Collection{
		Collectors: collectors,
		logger:     logger,
		scrapeDurationDesc: metric.TypedDesc{
			Desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
				exporterName+": Duration of a collector scrape.",
				[]string{"collector"},
				nil,
			),
			ValueType: prometheus.GaugeValue,
		},
		scrapeSuccessDesc: metric.TypedDesc{
			Desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "scrape", "collector_success"),
				exporterName+": Whether a collector succeeded.",
				[]string{"collector"},
				nil,
			),
			ValueType: prometheus.GaugeValue,
		},
	}, nil
}

// Describe implements the prometheus.Collector interface.
func (c Collection) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.scrapeDurationDesc.Desc
	ch <- c.scrapeSuccessDesc.Desc
}

// Collect implements the prometheus.Collector interface.
func (c Collection) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(c.Collectors))
	for name, collector := range c.Collectors {
		go func(name string, collector Collector) {
			execute(name, collector, ch, c.logger, c.scrapeDurationDesc, c.scrapeSuccessDesc)
			wg.Done()
		}(name, collector)
	}
	wg.Wait()
}

func execute(name string, c Collector, ch chan<- prometheus.Metric, logger *slog.Logger, scrapeDurationDesc, scrapeSuccessDesc metric.TypedDesc) {
	begin := time.Now()
	err := c.Update(ch)
	duration := time.Since(begin)
	var success float64

	if err != nil {
		if isNoDataError(err) {
			logger.Debug("collector returned no data", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		} else {
			logger.Error("collector failed", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		}
		success = 0
	} else {
		logger.Debug("collector succeeded", "name", name, "duration_seconds", duration.Seconds())
		success = 1
	}
	scrapeDurationDesc.PushMetric(ch, duration.Seconds(), name)
	scrapeSuccessDesc.PushMetric(ch, success, name)
}
