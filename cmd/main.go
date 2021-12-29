package main

import (
	"basexporter/required/functions"
	"basexporter/required/structs"
	"basexporter/util"
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
