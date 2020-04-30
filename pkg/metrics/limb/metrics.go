package limb

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	namespace = "limb"

	connections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "connect_connections",
			Help:      "How many connections are connecting adaptor.",
		},
		[]string{"adaptor"},
	)

	connectErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "connect_errors_total",
			Help:      "Total number of connecting adaptor errors.",
		},
		[]string{"adaptor"},
	)

	sendErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "send_errors_total",
			Help:      "Total number of errors of sending device desired to adaptor.",
		},
		[]string{"adaptor"},
	)

	sendLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "send_latency_seconds",
			Help:      "Histogram of the latency of sending device desired to adaptor.",
		},
		[]string{"adaptor"},
	)
)

func RegisterMetrics(registry prometheus.Registerer) error {
	var collectors = []prometheus.Collector{
		connections,
		connectErrors,
		sendErrors,
		sendLatency,
	}

	for _, collector := range collectors {
		if err := registry.Register(collector); err != nil {
			return err
		}
	}
	return nil
}

type MetricsRecorder interface {
	// IncreaseConnections increases the number of active connections.
	IncreaseConnections(adaptorName string)

	// DecreaseConnections decreases the number of active connections.
	DecreaseConnections(adaptorName string)

	// IncreaseConnectErrors increases the error counter when failed to connect to adaptor.
	IncreaseConnectErrors(adaptorName string)

	// ObserveSendLatency observes the histogram metrics when sending device desired to adaptor.
	ObserveSendLatency(adaptorName string, latency time.Duration)

	// IncreaseSendErrors increases the error counter when failed to send to adaptor.
	IncreaseSendErrors(adaptorName string)
}

type metricsRecorder struct{}

func (metricsRecorder) IncreaseConnections(adaptorName string) {
	connections.WithLabelValues(adaptorName).Inc()
}

func (metricsRecorder) DecreaseConnections(adaptorName string) {
	connections.WithLabelValues(adaptorName).Dec()
}

func (metricsRecorder) IncreaseConnectErrors(adaptorName string) {
	connectErrors.WithLabelValues(adaptorName).Inc()
}

func (metricsRecorder) ObserveSendLatency(adaptorName string, latency time.Duration) {
	sendLatency.WithLabelValues(adaptorName).Observe(latency.Seconds())
}

func (metricsRecorder) IncreaseSendErrors(adaptorName string) {
	sendErrors.WithLabelValues(adaptorName).Inc()
}

var recorder = metricsRecorder{}

func GetMetricsRecorder() MetricsRecorder {
	return recorder
}
