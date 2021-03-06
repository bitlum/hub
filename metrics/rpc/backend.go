package rpc

import (
	"time"

	"github.com/bitlum/hub/metrics"
	"github.com/go-errors/errors"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// subsystem is used as the second part in the name of the metric,
	// after the metrics namespace.
	subsystem = "rpc"

	// requestLabel is used to distinguish different requests during the
	// process of metric analysis and alert rule constructing.
	requestLabel = "request"

	// severityLabel is used to distinguish different error codes by its
	// level of importance.
	severityLabel = "severity"
)

// MetricsBackend is a system which is responsible for receiving and storing
// the connector metricsBackend.
type MetricsBackend interface {
	AddRequest(request string)
	AddError(request string, severity metrics.Severity)
	AddPanic(request string)
	AddRequestDuration(request string, dur time.Duration)
}

// EmptyBackend is used as an empty metricsBackend backend in order to avoid
type EmptyBackend struct{}

func (b *EmptyBackend) AddRequest(query string)                            {}
func (b *EmptyBackend) AddError(query string, errCode string)              {}
func (b *EmptyBackend) AddPanic(query string)                              {}
func (b *EmptyBackend) AddRequestDuration(query string, dur time.Duration) {}

// PrometheusBackend is the main subsystem metrics implementation. Uses
// prometheus metrics singletons defined above.
//
// WARN: request name should be taken from limited set.
// Don't use dynamic naming, it may cause dramatic increase of the amount of
// data on the metric server.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
type PrometheusBackend struct {
	requestsTotal          *prometheus.CounterVec
	errorsTotal            *prometheus.CounterVec
	panicsTotal            *prometheus.CounterVec
	requestDurationSeconds *prometheus.HistogramVec
}

// AddRequest increases request counter for the given request name.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) AddRequest(request string) {
	m.requestsTotal.With(
		prometheus.Labels{
			requestLabel: request,
		},
	).Add(1)
}

// AddError increases error counter for the given request name.
//
// WARN: Error code name should be taken from limited set.
// Don't use dynamic naming, it may cause dramatic increase of the amount of
// data on the metric server.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) AddError(request string,
	severity metrics.Severity) {
	m.errorsTotal.With(
		prometheus.Labels{
			requestLabel:  request,
			severityLabel: string(severity),
		},
	).Add(1)
}

// AddPanic increases panic counter for the given request name.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) AddPanic(request string) {
	m.panicsTotal.With(
		prometheus.Labels{
			requestLabel: request,
		},
	).Add(1)
}

// AddRequestDuration sends the metric with how much time request has taken
// to proceed.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) AddRequestDuration(request string,
	dur time.Duration) {
	m.requestDurationSeconds.With(
		prometheus.Labels{
			requestLabel: request,
		},
	).Observe(dur.Seconds())
}

// InitMetricsBackend creates subsystem metrics for specified
// net. Creates and tries to register metrics singletons. If register was
// already done, than function not returning error.
func InitMetricsBackend(net string) (MetricsBackend, error) {
	backend := PrometheusBackend{}

	backend.requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "Total requests processed",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			requestLabel,
		},
	)

	if err := prometheus.Register(backend.requestsTotal); err != nil {
		return backend, errors.Errorf(
			"unable to register 'requestsTotal' metric:" +
				err.Error())

	}

	backend.errorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystem,
			Name:      "errors_total",
			Help:      "Total requests which processing ended with error",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			requestLabel,
			severityLabel,
		},
	)

	if err := prometheus.Register(backend.errorsTotal); err != nil {
		return backend, errors.Errorf(
			"unable to register 'errorsTotal' metric: " +
				err.Error())

	}

	backend.panicsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystem,
			Name:      "panics_total",
			Help:      "Total requests which processing ended with panic",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			requestLabel,
		},
	)

	if err := prometheus.Register(backend.panicsTotal); err != nil {
		return backend, errors.Errorf(
			"unable to register 'panicsTotal' metric: " +
				err.Error())
	}

	backend.requestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Help:      "request processing duration in seconds",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			requestLabel,
		},
	)

	if err := prometheus.Register(backend.requestDurationSeconds); err != nil {
		return backend, errors.Errorf(
			"unable to register 'requestDurationSeconds' metric: " +
				err.Error())
	}

	return backend, nil
}
