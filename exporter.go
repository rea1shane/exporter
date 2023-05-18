package exporter

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"github.com/rea1shane/gooooo/http"
	"github.com/sirupsen/logrus"
	"os"
	"runtime"
)

// Exporter basic information.
type Exporter struct {
	Name           string // Name stylized as snake_case, e.g. "node_exporter".
	Description    string // Description
	DefaultAddress string // DefaultAddress
}

// Run start Exporter.
func (e Exporter) Run(logger *logrus.Logger) {
	var (
		metricsPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.",
		).Default("/metrics").String()
		disableExporterMetrics = kingpin.Flag(
			"web.disable-exporter-metrics",
			"Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).",
		).Bool()
		maxRequests = kingpin.Flag(
			"web.max-requests",
			"Maximum number of parallel scrape requests. Use 0 to disable.",
		).Default("40").Int()
		disableDefaultCollectors = kingpin.Flag(
			"collector.disable-defaults",
			"Set all collectors to disabled by default.",
		).Default("false").Bool()
		maxProcs = kingpin.Flag(
			"runtime.gomaxprocs",
			"The target number of CPUs Go will run on (GOMAXPROCS)",
		).Envar("GOMAXPROCS").Default("1").Int()
		toolkitFlags = kingpinflag.AddFlags(kingpin.CommandLine, e.DefaultAddress)

		logLevel = kingpin.Flag(
			"log.level",
			"Only log messages with the given severity or above. One of: [debug, info, warn, error]",
		).Default("info").String()
		latencyThreshold = kingpin.Flag(
			"web.latency_threshold",
			"When the latency exceeds the threshold, the log level will change from INFO to WARN. Use 0 to disable.",
		).Default("0").Duration()
	)
	logLevelMap := make(map[string]logrus.Level)
	logLevelMap["debug"] = logrus.DebugLevel
	logLevelMap["info"] = logrus.InfoLevel
	logLevelMap["warn"] = logrus.WarnLevel
	logLevelMap["error"] = logrus.ErrorLevel
	logger.SetLevel(logLevelMap[*logLevel])

	kingpin.Version(version.Print(e.Name))
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	if *disableDefaultCollectors {
		DisableDefaultCollectors()
	}
	logger.Infof("Starting %s", e.Name)
	logger.Infof("Version: %s", version.Info())
	logger.Infof("Build context: %s", version.BuildContext())

	runtime.GOMAXPROCS(*maxProcs)
	logger.Debugf("Go MAXPROCS: %d", runtime.GOMAXPROCS(0))

	handler := http.NewHandler(logger, *latencyThreshold)
	handler.GET(*metricsPath, gin.WrapH(newHandler(!*disableExporterMetrics, *maxRequests, logger)))
	if *metricsPath != "/" {
		landingConfig := web.LandingConfig{
			Name:        e.Name,
			Description: e.Description,
			Version:     version.Info(),
			Links: []web.LandingLinks{
				{
					Address: *metricsPath,
					Text:    "Metrics",
				},
			},
		}
		landingPage, err := web.NewLandingPage(landingConfig)
		if err != nil {
			logger.Fatal(err)
		}
		handler.GET("/", gin.WrapH(landingPage))
	}

	handler.Run(*toolkitFlags.WebListenAddresses...)
}
