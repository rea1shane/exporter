package main

import (
	"github.com/rea1shane/basexporter/required/functions"
	"github.com/rea1shane/basexporter/required/structs"
	"github.com/rea1shane/basexporter/util"
)

func main() {
	logger := util.GetLogger()
	azkabanExporter := structs.Exporter{
		MetricNamespace: "basexporter",
		ExporterName:    "basexporter",
		DefaultPort:     9999,
	}
	functions.Start(logger, azkabanExporter, util.ParseArgs(azkabanExporter))
}
