package main

import (
	_ "net/http/pprof"

	"github.com/prometheus/exporter-toolkit/web"

	"github.com/rea1shane/exporter"

	_ "example/collector"
)

func main() {
	landingConfig := exporter.LandingPageConfig{
		HeaderColor:   "#b7999e",
		TitleCaseName: "Example Exporter",
		Description:   "This is an example exporter.",
		Links: []web.LandingLinks{
			{
				Address:     "https://github.com/rea1shane/exporter",
				Text:        "rea1shane/exporter",
				Description: "Example exporter is build use this framework.",
			},
		},
	}
	// Open http://localhost:7777 in browser.
	exporter.Run("example_exporter", "example", ":7777", landingConfig, true)
}
