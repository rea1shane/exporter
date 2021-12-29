package functions

import (
	"github.com/rea1shane/basexporter/internal/http"
	"github.com/rea1shane/basexporter/internal/prometheus"
	"github.com/rea1shane/basexporter/required/structs"
	log "github.com/sirupsen/logrus"
)

func Start(logger *log.Logger, e structs.Exporter, args structs.Args) {
	http.Start(logger, e, args)
}

// RegisterCollector After you implement the structs.Collector, you should call this func to regist it.
func RegisterCollector(collector string, isDefaultEnabled bool, factory func(namespace string, logger *log.Entry) (structs.Collector, error)) {
	prometheus.RegisterCollector(collector, isDefaultEnabled, factory)
}
