package exporter

import (
	"net/http"
	"os"
	"os/user"
	"runtime"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"github.com/sirupsen/logrus"

	"github.com/rea1shane/exporter/collector"
)

type Exporter struct {
}

func New(logger *logrus.Logger, snakeCaseName, titleCaseName, description, namespace, defaultAddress string, warningRunAsRoot bool) {
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
			"runtime.gomaxprocs", "The target number of CPUs Go will run on (GOMAXPROCS)",
		).Envar("GOMAXPROCS").Default("1").Int()
		toolkitFlags = kingpinflag.AddFlags(kingpin.CommandLine, defaultAddress)
	)

	kingpin.Version(version.Print(snakeCaseName))
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	if *disableDefaultCollectors {
		collector.DisableDefaultCollectors()
	}
	logger.Infof("Starting %s %s", snakeCaseName, version.Info())
	logger.Info("Build context", "build_context", version.BuildContext())
	if user, err := user.Current(); warningRunAsRoot && err == nil && user.Uid == "0" {
		logger.Warnf("%s is running as root user. This exporter is designed to run as unprivileged user, root is not required.", titleCaseName)
	}
	runtime.GOMAXPROCS(*maxProcs)
	logger.Debug("Go MAXPROCS", "procs", runtime.GOMAXPROCS(0))

	http.Handle(*metricsPath, newHandler(snakeCaseName, namespace, !*disableExporterMetrics, *maxRequests, logger))
	if *metricsPath != "/" {
		landingConfig := web.LandingConfig{
			Name:        titleCaseName,
			Description: description,
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
			logger.Error(err.Error())
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}

	server := &http.Server{}
	if err := web.ListenAndServe(server, toolkitFlags, logger); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
