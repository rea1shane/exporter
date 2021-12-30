package basexporter

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	promcollectors "github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	log "github.com/sirupsen/logrus"
	stdlog "log"
	"net/http"
	"sort"
)

// handler is a common handler which implement http.Handler
type handler struct {
	unfilteredHandler       http.Handler
	exporterMetricsRegistry *prometheus.Registry // exporterMetricsRegistry is a separate registry for the metrics about the exporter itself.
	includeExporterMetrics  bool
	maxRequests             int
	logger                  *log.Logger
	exporterName            string
	namespace               string
}

func newHandler(exporterName string, namespace string, includeExporterMetrics bool, maxRequests int, logger *log.Logger) *handler {
	h := &handler{
		exporterMetricsRegistry: prometheus.NewRegistry(),
		includeExporterMetrics:  includeExporterMetrics,
		maxRequests:             maxRequests,
		logger:                  logger,
		exporterName:            exporterName,
		namespace:               namespace,
	}
	if h.includeExporterMetrics {
		h.exporterMetricsRegistry.MustRegister(
			promcollectors.NewProcessCollector(promcollectors.ProcessCollectorOpts{}),
			promcollectors.NewGoCollector(),
		)
	}
	if innerHandler, err := h.innerHandler(); err != nil {
		panic(fmt.Sprintf("Couldn't create metrics handler: %s", err))
	} else {
		h.unfilteredHandler = innerHandler
	}
	return h
}

// ServeHTTP implements http.Handler.
func (h *handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	filters := request.URL.Query()["collect[]"]
	h.logger.Debug("collect query's filters: ", filters)

	if len(filters) == 0 {
		// No filters, use the prepared unfiltered handler.
		h.unfilteredHandler.ServeHTTP(writer, request)
		return
	}
	// To serve filtered metrics, we create a filtering handler on the fly.
	filteredHandler, err := h.innerHandler(filters...)
	if err != nil {
		h.logger.Warnf("Couldn't create filtered metrics handler\n%+v", err)
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte(fmt.Sprintf("Couldn't create filtered metrics handler: %s", err)))
		return
	}
	filteredHandler.ServeHTTP(writer, request)
}

// innerHandler is used to create both the one unfiltered http.Handler to be
// wrapped by the outer handler and also the filtered handlers created on the
// fly. The former is accomplished by calling innerHandler without any arguments
// (in which case it will log all the collectors enabled via command-line flags).
func (h *handler) innerHandler(filters ...string) (http.Handler, error) {
	targetCollector, err := newTargetCollector(h.exporterName, h.namespace, h.logger, filters...)
	if err != nil {
		return nil, fmt.Errorf("couldn't create collector: %s", err)
	}

	// Only log the creation of an unfiltered handler, which should happen
	// only once upon startup.
	if len(filters) == 0 {
		h.logger.Info("Enabled collectors")
		var collectors []string
		for n := range targetCollector.collectors {
			collectors = append(collectors, n)
		}
		sort.Strings(collectors)
		for _, c := range collectors {
			h.logger.Info("collector ", c)
		}
	}

	r := prometheus.NewRegistry()
	// TODO 这句要不要去掉
	r.MustRegister(version.NewCollector(h.exporterName))
	if err := r.Register(targetCollector); err != nil {
		return nil, fmt.Errorf("couldn't register "+h.namespace+" collector: %s", err)
	}
	handler := promhttp.HandlerFor(
		prometheus.Gatherers{h.exporterMetricsRegistry, r},
		promhttp.HandlerOpts{
			ErrorLog:            stdlog.New(h.logger.Out, "", 0),
			ErrorHandling:       promhttp.ContinueOnError,
			MaxRequestsInFlight: h.maxRequests,
			Registry:            h.exporterMetricsRegistry,
		},
	)
	if h.includeExporterMetrics {
		// Note that we have to use h.exporterMetricsRegistry here to use the same promhttp metrics for all expositions.
		handler = promhttp.InstrumentMetricHandler(
			h.exporterMetricsRegistry, handler,
		)
	}
	return handler, nil
}
