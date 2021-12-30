package util

import (
	"fmt"
	"github.com/prometheus/common/version"
	"github.com/rea1shane/basexporter"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

func ParseArgs(e basexporter.Exporter) basexporter.Args {
	var (
		listenAddress = kingpin.Flag(
			"web.listen-address",
			"Address on which to expose metrics and web interface.",
		).Default(fmt.Sprintf(":%d", e.DefaultPort)).String()
		metricsPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.",
		).Default("/metrics").String()
		disableExporterMetrics = kingpin.Flag(
			"web.disable-exporter-metrics",
			"Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).",
		).Default("false").Bool()
		maxRequests = kingpin.Flag(
			"web.max-requests",
			"Maximum number of parallel scrape requests. Use 0 to disable.",
		).Default("40").Int()
		logLevel = kingpin.Flag(
			"log.level",
			"Only log messages with the given severity or above. One of: [debug, info, warn, error]",
		).Default("info").String()
		ginMode = kingpin.Flag(
			"gin.mode",
			"Gin's mode, suggest release mode in production. One of: [debug, release, test]",
		).Default("release").String()
	)

	kingpin.Version(version.Print(e.ExporterName))
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	return basexporter.Args{
		ListenAddress:          *listenAddress,
		MetricsPath:            *metricsPath,
		DisableExporterMetrics: *disableExporterMetrics,
		MaxRequests:            *maxRequests,
		LogLevel:               *logLevel,
		GinMode:                *ginMode,
	}
}
