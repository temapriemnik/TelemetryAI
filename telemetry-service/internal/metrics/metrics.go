package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	registry = prometheus.NewRegistry()

	ProcessedLogs = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "telemetry_service_logs_processed_total",
		Help: "Total number of logs processed",
	}, []string{"level"})

	ProcessingErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "telemetry_service_processing_errors_total",
		Help: "Total number of processing errors",
	}, []string{"error_type"})

	ProcessingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "telemetry_service_processing_duration_seconds",
		Help:    "Time spent processing logs",
		Buckets: prometheus.DefBuckets,
	})

	KafkaMessagesReceived = promauto.NewCounter(prometheus.CounterOpts{
		Name: "telemetry_service_kafka_messages_received_total",
		Help: "Total number of Kafka messages received",
	})

	NatsMessagesReceived = promauto.NewCounter(prometheus.CounterOpts{
		Name: "telemetry_service_nats_messages_received_total",
		Help: "Total number of NATS messages received",
	})
)

func init() {
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(ProcessedLogs)
	registry.MustRegister(ProcessingErrors)
	registry.MustRegister(ProcessingDuration)
	registry.MustRegister(NatsMessagesReceived)
}

func Handler() http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}