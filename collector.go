package basexporter

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"strings"
	"sync"
	"time"
)

var (
	factories              = make(map[string]func(namespace string, logger *log.Entry) (Collector, error)) // factories records all collector's construction method
	initiatedCollectorsMtx = sync.Mutex{}                                                                  // initiatedCollectorsMtx avoid thread conflicts
	initiatedCollectors    = make(map[string]Collector)                                                    // initiatedCollectors record the collectors that have been initialized in the method newTargetCollector (To reduce the collector's construction method call)
	collectorState         = make(map[string]*bool)                                                        // collectorState records all collector's default state (enable or disable)
	forcedCollectors       = map[string]bool{}                                                             // forcedCollectors collectors which have been explicitly enabled or disabled
)

type targetCollector struct {
	collectors         map[string]Collector
	logger             *log.Logger
	scrapeDurationDesc *prometheus.Desc
	scrapeSuccessDesc  *prometheus.Desc
}

func (t targetCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- t.scrapeDurationDesc
	ch <- t.scrapeSuccessDesc
}

func (t targetCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	for name, c := range t.collectors {
		wg.Add(1)
		go func(name string, c Collector) {
			defer wg.Done()
			execute(name, c, ch, t.logger, t.scrapeDurationDesc, t.scrapeSuccessDesc)
		}(name, c)
	}
	wg.Wait()
}

// newTargetCollector creates a new targetCollector.
func newTargetCollector(exporterName string, namespace string, logger *log.Logger, filters ...string) (*targetCollector, error) {
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
			collector, err := factories[key](namespace, logger.WithField("collector", key))
			if err != nil {
				return nil, err
			}
			collectors[key] = collector
			initiatedCollectors[key] = collector
		}
	}
	return &targetCollector{
		collectors: collectors,
		logger:     logger,
		scrapeDurationDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
			exporterName+": Duration of a collector scrape.",
			[]string{"collector"},
			nil,
		),
		scrapeSuccessDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "scrape", "collector_success"),
			exporterName+": Whether a collector succeeded.",
			[]string{"collector"},
			nil,
		),
	}, nil
}

// Collector is an interface that a collector need to implement.
type Collector interface {
	// Update Get new metrics and expose them via prometheus registry.
	Update(ch chan<- prometheus.Metric) error
}

// RegisterCollector After you implement the structs.Collector, you should call this func to register it.
func RegisterCollector(collector string, isDefaultEnabled bool, factory func(namespace string, logger *log.Entry) (Collector, error)) {
	var helpDefaultState string
	if isDefaultEnabled {
		helpDefaultState = "enabled"
	} else {
		helpDefaultState = "disabled"
	}

	flagName := fmt.Sprintf("collector.%s", collector)
	flagHelp := fmt.Sprintf("Enable the %s collector (default: %s).", collector, helpDefaultState)
	defaultValue := fmt.Sprintf("%v", isDefaultEnabled)

	flag := kingpin.Flag(flagName, flagHelp).Default(defaultValue).Action(collectorFlagAction(collector)).Bool()
	collectorState[collector] = flag

	factories[collector] = factory
}

func collectorFlagAction(collector string) func(ctx *kingpin.ParseContext) error {
	return func(ctx *kingpin.ParseContext) error {
		forcedCollectors[collector] = true
		return nil
	}
}

func execute(name string, c Collector, ch chan<- prometheus.Metric, logger *log.Logger, scrapeDurationDesc *prometheus.Desc, scrapeSuccessDesc *prometheus.Desc) {
	begin := time.Now()
	err := c.Update(ch)
	duration := time.Since(begin)
	var success float64

	if err != nil {
		str1 := fmt.Sprintf("%+v", err)
		str2 := strings.TrimRight(str1, "\n")
		str3 := strings.Replace(str2, "\n}", "\n    }", -1)
		str4 := strings.Replace(str3, "\n  \"", "\n      \"", -1)
		str5 := strings.Replace(str4, "\n", "\n    ", -1)
		logger.
			WithField("name", name).
			WithField("duration_seconds", fmt.Sprintf("%v", duration.Milliseconds())+"ms").
			Errorf("collector failed\n└──>%+v", str5)
		success = 0
	} else {
		logger.
			WithField("name", name).
			WithField("duration_seconds", fmt.Sprintf("%v", duration.Milliseconds())+"ms").
			Debug("collector succeeded")
		success = 1
	}
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
}
