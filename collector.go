package exporter

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	factories              = make(map[string]func(namespace string, logger *logrus.Entry) (Collector, error)) // factories records all collector's construction method
	initiatedCollectorsMtx = sync.Mutex{}                                                                     // initiatedCollectorsMtx avoid thread conflicts
	initiatedCollectors    = make(map[string]Collector)                                                       // initiatedCollectors record the collectors that have been initialized in the method newTargetCollector (To reduce the collector's construction method call)
	collectorState         = make(map[string]*bool)                                                           // collectorState records all collector's default state (enable or disable)
	forcedCollectors       = map[string]bool{}                                                                // forcedCollectors collectors which have been explicitly enabled or disabled
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

// Collector is the interface a collector has to implement.
type Collector interface {
	Update(ch chan<- prometheus.Metric) error // Update get new metrics and expose them via prometheus registry.
}

const (
	DefaultEnabled  = true
	DefaultDisabled = false
)

// RegisterCollector should be called once you implement the Collector interface.
func RegisterCollector(collector string, isDefaultEnabled bool, factory func(namespace string, logger *logrus.Entry) (Collector, error)) {
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

// collectorFlagAction generates a new action function for the given collector
// to track whether it has been explicitly enabled or disabled from the command line.
// A new action function is needed for each collector flag because the ParseContext
// does not contain information about which flag called the action.
// See: https://github.com/alecthomas/kingpin/issues/294
func collectorFlagAction(collector string) func(ctx *kingpin.ParseContext) error {
	return func(ctx *kingpin.ParseContext) error {
		forcedCollectors[collector] = true
		return nil
	}
}

// DisableDefaultCollectors sets the collector state to false for all collectors which
// have not been explicitly enabled on the command line.
func DisableDefaultCollectors() {
	for c := range collectorState {
		if _, ok := forcedCollectors[c]; !ok {
			*collectorState[c] = false
		}
	}
}

type typedDesc struct {
	desc      *prometheus.Desc
	valueType prometheus.ValueType
}

func (d *typedDesc) mustNewConstMetric(value float64, labels ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(d.desc, d.valueType, value, labels...)
}

// ErrNoData indicates the collector found no data to collect, but had no other error.
var ErrNoData = errors.New("collector returned no data")

func IsNoDataError(err error) bool {
	return err == ErrNoData
}

// pushMetric helps construct and convert a variety of value types into Prometheus float64 metrics.
func pushMetric(ch chan<- prometheus.Metric, fieldDesc *prometheus.Desc, name string, value interface{}, valueType prometheus.ValueType, labelValues ...string) {
	var fVal float64
	switch val := value.(type) {
	case uint8:
		fVal = float64(val)
	case uint16:
		fVal = float64(val)
	case uint32:
		fVal = float64(val)
	case uint64:
		fVal = float64(val)
	case int64:
		fVal = float64(val)
	case *uint8:
		if val == nil {
			return
		}
		fVal = float64(*val)
	case *uint16:
		if val == nil {
			return
		}
		fVal = float64(*val)
	case *uint32:
		if val == nil {
			return
		}
		fVal = float64(*val)
	case *uint64:
		if val == nil {
			return
		}
		fVal = float64(*val)
	case *int64:
		if val == nil {
			return
		}
		fVal = float64(*val)
	default:
		return
	}

	ch <- prometheus.MustNewConstMetric(fieldDesc, valueType, fVal, labelValues...)
}
