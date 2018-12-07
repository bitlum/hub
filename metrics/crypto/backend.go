package crypto

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/go-errors/errors"
	"github.com/bitlum/hub/metrics"
)

const (
	// subsystem is used as the second part in the name of the metric,
	// after the metrics namespace.
	subsystem = "crypto"

	// methodLabel is used to distinguish different methods during the
	// process of metric analysis and alert rule constructing.
	methodLabel = "method"

	// assetLabel is used to distinguish different currency and daemon on the
	// metrics server.
	assetLabel = "asset"

	// severityLabel is used to distinguish different error codes by its
	// level of importance.
	severityLabel = "severity"
)

// MetricsBackend is a system which is responsible for receiving and storing
// the connector metricsBackend.
type MetricsBackend interface {
	AddMethod(asset, method string)
	AddError(asset, method string, severity metrics.Severity)
	AddPanic(asset, method string)
	AddMethodDuration(asset, method string, dur time.Duration)
}

// EmptyBackend is used as an empty metricsBackend backend in order to avoid
type EmptyBackend struct{}

func (b *EmptyBackend) AddMethod(query string)                            {}
func (b *EmptyBackend) AddError(query string, errCode string)             {}
func (b *EmptyBackend) AddPanic(query string)                             {}
func (b *EmptyBackend) AddMethodDuration(query string, dur time.Duration) {}

// PrometheusBackend is the main subsystem metrics implementation. Uses
// prometheus metrics singletons defined above.
//
// WARN: Method name should be taken from limited set.
// Don't use dynamic naming, it may cause dramatic increase of the amount of
// data on the metric server.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
type PrometheusBackend struct {
	methodsTotal          *prometheus.CounterVec
	errorsTotal           *prometheus.CounterVec
	panicsTotal           *prometheus.CounterVec
	methodDurationSeconds *prometheus.HistogramVec
}

// AddMethod increases method counter for the given method name.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) AddMethod(asset, method string) {
	m.methodsTotal.With(
		prometheus.Labels{
			methodLabel: method,
			assetLabel:  asset,
		},
	).Add(1)
}

// AddError increases error counter for the given method name.
//
// WARN: Error code name should be taken from limited set.
// Don't use dynamic naming, it may cause dramatic increase of the amount of
// data on the metric server.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) AddError(asset, method string,
	severity metrics.Severity) {
	m.errorsTotal.With(
		prometheus.Labels{
			methodLabel:   method,
			severityLabel: string(severity),
			assetLabel:    asset,
		},
	).Add(1)
}

// AddPanic increases panic counter for the given method name.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) AddPanic(asset, method string) {
	m.panicsTotal.With(
		prometheus.Labels{
			methodLabel: method,
			assetLabel:  asset,
		},
	).Add(1)
}

// AddMethodDuration sends the metric with how much time method has taken
// to proceed.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) AddMethodDuration(asset, method string,
	dur time.Duration) {
	m.methodDurationSeconds.With(
		prometheus.Labels{
			methodLabel: method,
			assetLabel:  asset,
		},
	).Observe(dur.Seconds())
}

// InitMetricsBackend creates subsystem metrics for specified
// net. Creates and tries to register metrics singletons. If register was
// already done, than function not returning error.
func InitMetricsBackend(net string) (MetricsBackend, error) {
	backend := PrometheusBackend{}

	backend.methodsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystem,
			Name:      "methods_total",
			Help:      "Total methods processed",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			methodLabel,
			assetLabel,
		},
	)

	if err := prometheus.Register(backend.methodsTotal); err != nil {
		return backend, errors.Errorf(
			"unable to register 'methodsTotal' metric:" +
				err.Error())

	}

	backend.errorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystem,
			Name:      "errors_total",
			Help:      "Total methods which processing ended with error",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			methodLabel,
			assetLabel,
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
			Help:      "Total methods which processing ended with panic",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			methodLabel,
			assetLabel,
		},
	)

	if err := prometheus.Register(backend.panicsTotal); err != nil {
		return backend, errors.Errorf(
			"unable to register 'panicsTotal' metric: " +
				err.Error())
	}

	backend.methodDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystem,
			Name:      "method_duration_seconds",
			Help:      "Method processing duration in seconds",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			methodLabel,
			assetLabel,
		},
	)

	if err := prometheus.Register(backend.methodDurationSeconds); err != nil {
		return backend, errors.Errorf(
			"unable to register 'methodDurationSeconds' metric: " +
				err.Error())
	}

	return backend, nil
}
