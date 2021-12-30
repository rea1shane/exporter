package main

import (
	"github.com/rea1shane/basexporter"
	"github.com/rea1shane/basexporter/util"
)

func main() {
	logger := util.GetLogger()
	azkabanExporter := basexporter.Exporter{
		MetricNamespace: "basexporter",
		ExporterName:    "basexporter",
		DefaultPort:     9999,
	}
	basexporter.Start(logger, azkabanExporter, util.ParseArgs(azkabanExporter))
}
