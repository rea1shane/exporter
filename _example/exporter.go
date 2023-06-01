package main

import (
	_ "example/collector"
	"github.com/gin-gonic/gin"
	"github.com/rea1shane/exporter"
	"github.com/rea1shane/gooooo/log"
	"github.com/sirupsen/logrus"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	logger := logrus.New()
	formatter := log.NewFormatter()
	formatter.FieldsOrder = []string{"StatusCode", "Latency", "Collector", "Duration"}
	logger.SetFormatter(formatter)
	exporter.Register("test_exporter", "test", "This is a test exporter.", ":7777", logger)
	exporter.Run()
}
