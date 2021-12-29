package prometheus

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rea1shane/basexporter/required/structs"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"strings"
	"sync"
	"time"
)

var (
	Factories              = make(map[string]func(namespace string, logger *log.Entry) (structs.Collector, error)) // Factories records all collector's construction method
	InitiatedCollectorsMtx = sync.Mutex{}                                                                          // InitiatedCollectorsMtx avoid thread conflicts
	InitiatedCollectors    = make(map[string]structs.Collector)                                                    // InitiatedCollectors record the collectors that have been initialized in the method NewTargetCollector (To reduce the collector's construction method call)
	CollectorState         = make(map[string]*bool)                                                                // CollectorState records all collector's default state (enable or disable)
	ForcedCollectors       = map[string]bool{}                                                                     // ForcedCollectors collectors which have been explicitly enabled or disabled
)

type TargetCollector struct {
	Collectors         map[string]structs.Collector
	Logger             *log.Logger
	ScrapeDurationDesc *prometheus.Desc
	ScrapeSuccessDesc  *prometheus.Desc
}

func (t TargetCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- t.ScrapeDurationDesc
	ch <- t.ScrapeSuccessDesc
}

func (t TargetCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	for name, c := range t.Collectors {
		wg.Add(1)
		go func(name string, c structs.Collector) {
			defer wg.Done()
			Execute(name, c, ch, t.Logger, t.ScrapeDurationDesc, t.ScrapeSuccessDesc)
		}(name, c)
	}
	wg.Wait()
}

// NewTargetCollector creates a new TargetCollector.
func NewTargetCollector(exporter structs.Exporter, logger *log.Logger, filters ...string) (*TargetCollector, error) {
	f := make(map[string]bool)
	for _, filter := range filters {
		enabled, exist := CollectorState[filter]
		if !exist {
			return nil, fmt.Errorf("missing collector: %s", filter)
		}
		if !*enabled {
			return nil, fmt.Errorf("disabled collector: %s", filter)
		}
		f[filter] = true
	}
	collectors := make(map[string]structs.Collector)
	InitiatedCollectorsMtx.Lock()
	defer InitiatedCollectorsMtx.Unlock()
	for key, enabled := range CollectorState {
		if !*enabled || (len(f) > 0 && !f[key]) {
			continue
		}
		if collector, ok := InitiatedCollectors[key]; ok {
			collectors[key] = collector
		} else {
			collector, err := Factories[key](exporter.MetricNamespace, logger.WithField("collector", key))
			if err != nil {
				return nil, err
			}
			collectors[key] = collector
			InitiatedCollectors[key] = collector
		}
	}
	return &TargetCollector{
		Collectors: collectors,
		Logger:     logger,
		ScrapeDurationDesc: prometheus.NewDesc(
			prometheus.BuildFQName(exporter.MetricNamespace, "scrape", "collector_duration_seconds"),
			exporter.ExporterName+": Duration of a collector scrape.",
			[]string{"collector"},
			nil,
		),
		ScrapeSuccessDesc: prometheus.NewDesc(
			prometheus.BuildFQName(exporter.MetricNamespace, "scrape", "collector_success"),
			exporter.ExporterName+": Whether a collector succeeded.",
			[]string{"collector"},
			nil,
		),
	}, nil
}

func RegisterCollector(collector string, isDefaultEnabled bool, factory func(namespace string, logger *log.Entry) (structs.Collector, error)) {
	var helpDefaultState string
	if isDefaultEnabled {
		helpDefaultState = "enabled"
	} else {
		helpDefaultState = "disabled"
	}

	flagName := fmt.Sprintf("collector.%s", collector)
	flagHelp := fmt.Sprintf("Enable the %s collector (default: %s).", collector, helpDefaultState)
	defaultValue := fmt.Sprintf("%v", isDefaultEnabled)

	flag := kingpin.Flag(flagName, flagHelp).Default(defaultValue).Action(CollectorFlagAction(collector)).Bool()
	CollectorState[collector] = flag

	Factories[collector] = factory
}

func CollectorFlagAction(collector string) func(ctx *kingpin.ParseContext) error {
	return func(ctx *kingpin.ParseContext) error {
		ForcedCollectors[collector] = true
		return nil
	}
}

func Execute(name string, c structs.Collector, ch chan<- prometheus.Metric, logger *log.Logger, scrapeDurationDesc *prometheus.Desc, scrapeSuccessDesc *prometheus.Desc) {
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
