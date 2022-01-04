package basexporter

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var srv *http.Server

// Exporter 's basic info
type Exporter struct {
	metricNamespace string // metricNamespace suggest one word and low case
	exporterName    string // exporterName suggest snake format, like node_exporter
	defaultPort     int    // defaultPort is default web listen port of exporter
	version         string // version is exporter's version
}

// Args is basic parameters required by the program
type Args struct {
	listenAddress          string
	metricsPath            string
	disableExporterMetrics bool
	maxRequests            int
	logLevel               string
	ginMode                string
}

// BuildExporter return a new Exporter
func BuildExporter(metricNamespace string, exporterName string, defaultPort int, version string) Exporter {
	return Exporter{
		metricNamespace: metricNamespace,
		exporterName:    exporterName,
		defaultPort:     defaultPort,
		version:         version,
	}
}

// BuildArgs return a new Args
func BuildArgs(listenAddress string, metricsPath string, disableExporterMetrics bool, maxRequests int, logLevel string, ginMode string) Args {
	return Args{
		listenAddress:          listenAddress,
		metricsPath:            metricsPath,
		disableExporterMetrics: disableExporterMetrics,
		maxRequests:            maxRequests,
		logLevel:               logLevel,
		ginMode:                ginMode,
	}
}

func Start(logger *log.Logger, e Exporter, args Args) {
	switch args.logLevel {
	case "debug":
		logger.SetLevel(log.DebugLevel)
	case "info":
		logger.SetLevel(log.InfoLevel)
	case "warn":
		logger.SetLevel(log.WarnLevel)
	case "error":
		logger.SetLevel(log.ErrorLevel)
	default:
		panic("log level unknown: " + args.logLevel + " (run -h get more information)")
	}

	switch args.ginMode {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		panic("gin mode unknown: " + args.ginMode + " (run -h get more information)")
	}

	displayName := camelString(e.exporterName)
	logger.Info("Starting " + displayName + ", version: " + e.version)

	app := gin.New()
	app.Use(
		toStdout(logger),
		gin.Recovery(),
	)
	app.GET("/", gin.WrapF(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
			<head><title>` + displayName + `</title></head>
			<body>
			<h1>` + displayName + `</h1>
			<p><a href="` + args.metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	}))
	app.GET(args.metricsPath, gin.WrapH(newHandler(e.exporterName, e.metricNamespace, e.version, !args.disableExporterMetrics, args.maxRequests, logger)))

	logger.Info("Listening on address ", args.listenAddress)
	srv = &http.Server{
		Addr:    args.listenAddress,
		Handler: app,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	shutdown(logger)
}

func shutdown(logger *log.Logger) {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGQUIT)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("server shutdown\n%+v", err)
		return
	}
}

func camelString(s string) string {
	data := make([]byte, 0, len(s))
	j := false
	k := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if k == false && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || k == false) {
			if i != 0 {
				data = append(data, ' ')
			}
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:])
}

func toStdout(logger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)

		reqMethod := c.Request.Method
		reqUri := c.Request.RequestURI
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		entry := logger.
			WithField("status_code", statusCode)

		if statusCode == 200 {
			entry.Infof("%15v | %15v | %7v %v", latencyTime, clientIP, reqMethod, reqUri)
		} else if statusCode == 404 {
			entry.Warnf("%15v | %15v | %7v %v", latencyTime, clientIP, reqMethod, reqUri)
		} else {
			entry.Errorf("%15v | %15v | %7v %v", latencyTime, clientIP, reqMethod, reqUri)
		}
	}
}

// ParseArgs You can implement another yourself if you need
func ParseArgs(defaultPort int) Args {
	var (
		listenAddress = kingpin.Flag(
			"web.listen-address",
			"Address on which to expose metrics and web interface.",
		).Default(fmt.Sprintf(":%d", defaultPort)).String()
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

	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	return BuildArgs(*listenAddress, *metricsPath, *disableExporterMetrics, *maxRequests, *logLevel, *ginMode)
}
