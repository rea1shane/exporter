package basexporter

import (
	"fmt"
	"github.com/rea1shane/basexporter/util"
	"testing"
)

func TestStart(m *testing.T) {
	logger := util.GetLogger()
	defaultPort := 9999
	exporter := BuildExporter("basexporter", "basexporter", defaultPort, "1.0.0")
	// Can't use kingpin in test, reason: https://github.com/alecthomas/kingpin/issues/187
	// Start(logger, exporter, ParseArgs(defaultPort))
	Start(logger, exporter, BuildArgs(
		fmt.Sprintf(":%d", defaultPort),
		"/metrics",
		false,
		40,
		"info",
		"release"),
	)
}
