package http

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/common/version"
	"github.com/rea1shane/basexporter/internal/middleware"
	"github.com/rea1shane/basexporter/internal/prometheus"
	"github.com/rea1shane/basexporter/required/structs"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var srv *http.Server

func Start(logger *log.Logger, e structs.Exporter, args structs.Args) {
	switch args.LogLevel {
	case "debug":
		logger.SetLevel(log.DebugLevel)
	case "info":
		logger.SetLevel(log.InfoLevel)
	case "warn":
		logger.SetLevel(log.WarnLevel)
	case "error":
		logger.SetLevel(log.ErrorLevel)
	default:
		panic("log level unknown: " + args.LogLevel + " (run -h get more information)")
	}

	switch args.GinMode {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		panic("gin mode unknown: " + args.GinMode + " (run -h get more information)")
	}

	displayName := camelString(e.ExporterName)

	logger.Info("Starting "+e.ExporterName, " version", version.Info())
	logger.Info("Build context", version.BuildContext())

	app := gin.New()
	app.Use(
		middleware.ToStdout(logger),
		gin.Recovery(),
	)
	app.GET("/", gin.WrapF(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
			<head><title>` + displayName + `</title></head>
			<body>
			<h1>` + displayName + `</h1>
			<p><a href="` + args.MetricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	}))
	app.GET(args.MetricsPath, gin.WrapH(prometheus.NewHandler(e, !args.DisableExporterMetrics, args.MaxRequests, logger)))

	logger.Info("Listening on address ", args.ListenAddress)
	srv = &http.Server{
		Addr:    args.ListenAddress,
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
			data = append(data, ' ')
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
