package exporter

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

var (
	initiatedCollectorsMtx = sync.Mutex{}               // initiatedCollectorsMtx avoid thread conflicts
	initiatedCollectors    = make(map[string]Collector) // initiatedCollectors record the collectors that have been initialized in the method newCollectorKeeper (To reduce the collector's construction method call)
)

// collectorKeeper implements the prometheus.Collector interface.
type collectorKeeper struct {
	collectors         map[string]Collector
	logger             *logrus.Logger
	scrapeDurationDesc *prometheus.Desc
	scrapeSuccessDesc  *prometheus.Desc
}

// newCollectorKeeper creates a new collectorKeeper.
func newCollectorKeeper(exporterName string, namespace string, logger *logrus.Logger, filters ...string) (*collectorKeeper, error) {
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
			collector, err := factories[key](namespace, logger.WithField("Collector", key))
			if err != nil {
				return nil, err
			}
			collectors[key] = collector
			initiatedCollectors[key] = collector
		}
	}
	return &collectorKeeper{
		collectors: collectors,
		logger:     logger,
		scrapeDurationDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
			fmt.Sprintf("%s: Duration of a collector scrape.", exporterName),
			[]string{"collector"},
			nil,
		),
		scrapeSuccessDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "scrape", "collector_success"),
			fmt.Sprintf("%s: Whether a collector succeeded.", exporterName),
			[]string{"collector"},
			nil,
		),
	}, nil
}

// Describe implements the prometheus.Collector interface.
func (ck collectorKeeper) Describe(ch chan<- *prometheus.Desc) {
	ch <- ck.scrapeDurationDesc
	ch <- ck.scrapeSuccessDesc
}

// Collect implements the prometheus.Collector interface.
func (ck collectorKeeper) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(ck.collectors))
	for name, c := range ck.collectors {
		go func(name string, c Collector) {
			execute(name, c, ch, ck.logger, ck.scrapeDurationDesc, ck.scrapeSuccessDesc)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
}

func execute(name string, c Collector, ch chan<- prometheus.Metric, logger *logrus.Logger, scrapeDurationDesc, scrapeSuccessDesc *prometheus.Desc) {
	begin := time.Now()
	err := c.Update(ch)
	duration := time.Since(begin)
	var success float64

	entry := logger.WithFields(logrus.Fields{
		"Collector": name,
		"Duration":  duration,
	})

	if err != nil {
		if IsNoDataError(err) {
			entry.Debug("collector returned no data: ", err)
		} else {
			entry.Error("collector failed: ", err)
		}
		success = 0
	} else {
		entry.Debug("collector succeeded")
		success = 1
	}
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
}
