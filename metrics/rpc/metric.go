package rpc

import (
	"github.com/bitlum/hub/metrics"
	"runtime/debug"
	"strings"
	"time"
)

// Metric is an enhancement of Metric backend, which is more suited for this
// package usage.
type Metric struct {
	// backend is an entity which is responsible for collecting and storing
	// the metricsBackend.
	backend MetricsBackend

	// requestName is name used as value for label when adding metricsBackend.
	// Supposed to be set by `NewMetric` request.
	requestName string

	// startTime is a request start time used to measure request
	// duration in `AddMethodDuration` request. Supposed to be set
	// in `NewMetric` request to now.
	startTime time.Time
}

// NewMetric creates new Metric.
//
// Note: we use not pointer type receiver so any changes within request
// do not change original struct fields. Each call creates new `metricsBackend`
// with copied fields.
func NewMetric(request string, backend MetricsBackend) Metric {
	m := Metric{}

	m.backend = backend
	m.requestName = request
	m.startTime = time.Now()

	m.backend.AddRequest(request)
	return m
}

// AddError adds error metric with specified error.
//
// Note: we use not pointer type receiver so any changes within request
// do not change original struct fields. Each call creates new `metricsBackend`
// with copied fields.
func (m Metric) AddError(severity metrics.Severity) {
	m.backend.AddError(m.requestName, severity)
}

// AddPanic adds panic metric
func (m Metric) AddPanic() {
	m.backend.AddPanic(m.requestName)
}

// AddMethodDuration adds request duration metric. Supposed to be
// called after `NewMetric` which defines `startTime`. Calculates
// duration using `startTime` and now as end time.
//
// Note: we use not pointer type receiver so any changes within request
// do not change original struct fields. Each call creates new `metricsBackend`
// with copied fields.
func (m Metric) AddMethodDuration() {
	if m.startTime.Equal(time.Time{}) {
		panic("not initialised request")
	}

	dur := time.Now().Sub(m.startTime)
	m.backend.AddRequestDuration(m.requestName, dur)
}

// Finish used as defer in handlers, to ensure that we track panics and
// measure handler time.
func (m Metric) Finish() {
	m.AddMethodDuration()

	if r := recover(); r != nil {
		m.AddPanic()
		panic(stackTrace())
	}
}

func stackTrace() string {
	s := string(debug.Stack())
	ls := strings.Split(s, "\n")
	for i, l := range ls {
		if strings.Index(l, "src/runtime/panic.go") != -1 && i > 0 &&
			strings.Index(ls[i-1], "panic(") == 0 {
			return strings.TrimSpace(strings.Join(ls[i+2:], "\n"))
		}
	}
	return s
}
