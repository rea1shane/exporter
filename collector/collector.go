// Package collector includes all individual collectors to gather and export system metrics.
package collector

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/rea1shane/exporter/metric"
)

var (
	initiatedCollectorsMtx = sync.Mutex{}               // initiatedCollectorsMtx avoid thread conflicts
	initiatedCollectors    = make(map[string]Collector) // initiatedCollectors record the collectors that have been initialized in the method NewCollectorCollection (To reduce the collector's construction method call)
)

// collectorCollection implements the prometheus.Collector interface.
type collectorCollection struct {
	collectors         map[string]Collector
	logger             *logrus.Logger
	scrapeDurationDesc metric.TypedDesc
	scrapeSuccessDesc  metric.TypedDesc
}

// NewCollectorCollection creates a new collectorCollection.
// Namespace defines the common namespace to be used by all metrics.
func NewCollectorCollection(exporterName, namespace string, logger *logrus.Logger, filters ...string) (*collectorCollection, error) {
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
			c, err := factories[key](namespace, logger.WithField("collector", key))
			if err != nil {
				return nil, err
			}
			collectors[key] = c
			initiatedCollectors[key] = c
		}
	}
	return &collectorCollection{
		collectors: collectors,
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
func (cc collectorCollection) Describe(ch chan<- *prometheus.Desc) {
	ch <- cc.scrapeDurationDesc.Desc
	ch <- cc.scrapeSuccessDesc.Desc
}

// Collect implements the prometheus.Collector interface.
func (cc collectorCollection) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(cc.collectors))
	for name, c := range cc.collectors {
		go func(name string, c Collector) {
			execute(name, c, ch, cc.logger, cc.scrapeDurationDesc, cc.scrapeSuccessDesc)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
}

func execute(name string, c Collector, ch chan<- prometheus.Metric, logger *logrus.Logger, scrapeDurationDesc, scrapeSuccessDesc metric.TypedDesc) {
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
