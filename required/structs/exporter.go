package structs

// Exporter basic info
type Exporter struct {
	MetricNamespace string // MetricNamespace suggest one word and lowcase
	ExporterName    string // ExporterName suggest snake format, like node_exporter
	DefaultPort     int    // DefaultPort is default web listen port of exporter
}
