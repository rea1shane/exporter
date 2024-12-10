package exporter

import (
	"fmt"
	"net/http"
	"os"
	"os/user"
	"runtime"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"

	"github.com/rea1shane/exporter/collector"
)

// LandingPageConfig is fork from web.LandingConfig.
// All options are optional, but it is strongly recommended to declare at least TitleCaseName if you need a landing page.
// There will be a landing page if metricsPath is not "/".
type LandingPageConfig struct {
	HeaderColor   string             // HeaderColor used for the landing page header. NOTE: If CSS is not empty, HeaderColor has no effect.
	CSS           string             // CSS style tag for the landing page.
	TitleCaseName string             // TitleCaseName of the exporter. For example: Node Exporter.
	Description   string             // Description about the exporter.
	Form          web.LandingForm    // Form is a POST form.
	Links         []web.LandingLinks // Links you want to show on the landing page OTHER THAN METRICS.
	ExtraHTML     string             // ExtraHTML is additional HTML to be embedded.
	ExtraCSS      string             // ExtraCSS is additional CSS to be embedded.
}

// Run will start the exporter.
// snakeCaseName is exporter name in snake case. For example: node_exporter.
func Run(snakeCaseName, namespace, defaultAddress string, landingPageConfig LandingPageConfig, warningRunAsRoot bool) {
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

	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.Version(version.Print(snakeCaseName))
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promslog.New(promslogConfig)

	if *disableDefaultCollectors {
		collector.DisableDefaultCollectors()
	}
	logger.Info(fmt.Sprintf("Starting %s", snakeCaseName), "version", version.Info())
	logger.Info("Build context", "build_context", version.BuildContext())
	if user, err := user.Current(); warningRunAsRoot && err == nil && user.Uid == "0" {
		logger.Warn(fmt.Sprintf("%s is running as root user. This exporter is designed to run as unprivileged user, root is not required.", snakeCaseName))
	}
	runtime.GOMAXPROCS(*maxProcs)
	logger.Debug("Go MAXPROCS", "procs", runtime.GOMAXPROCS(0))

	http.Handle(*metricsPath, newHandler(snakeCaseName, namespace, !*disableExporterMetrics, *maxRequests, logger))
	if *metricsPath != "/" {
		landingConfig := web.LandingConfig{
			HeaderColor: landingPageConfig.HeaderColor,
			CSS:         landingPageConfig.CSS,
			Name:        landingPageConfig.TitleCaseName,
			Description: landingPageConfig.Description,
			Form:        landingPageConfig.Form,
			ExtraHTML:   landingPageConfig.ExtraHTML,
			ExtraCSS:    landingPageConfig.ExtraCSS,
			Version:     version.Info(),
			Links: append([]web.LandingLinks{
				{
					Address: *metricsPath,
					Text:    "Metrics",
				},
			}, landingPageConfig.Links...),
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
