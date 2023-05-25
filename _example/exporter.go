package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rea1shane/exporter"
	"github.com/rea1shane/gooooo/log"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	logger := log.NewLogger()
	exporter.Register("test_exporter", "test", "This is a test exporter.", ":7777", logger)
	exporter.Run()
}
