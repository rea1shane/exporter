package main

import (
	_ "net/http/pprof"

	_ "example/collector"

	"github.com/rea1shane/exporter"
)

func main() {
	exporter.Run("example_exporter", "Example Exporter", "This is a example exporter.", "example", ":7777", true)
}
