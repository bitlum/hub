package network

import (
	"github.com/bitlum/hub/metrics"
	"github.com/go-errors/errors"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// subsystem is used as the second part in the name of the metric,
	// after the metrics namespace.
	subsystem = "network"

	// assetLabel is used to distinguish different currency and daemon on the
	// metrics server.
	assetLabel = "asset"

	// channelStatusLabel is used to distinguish different channels statuses,
	// i.e. open, pending.
	channelStatusLabel = "channel_status"

	// channelStateLabel is used to distinguish different channels states,
	// i.e. active, inactive.
	channelStateLabel = "channel_state"

	// sideLabel is used to distinguish different sides, i.e. local node,
	// remote peer.
	sideLabel = "side"
)

// MetricsBackend is a system which is responsible for receiving,
// storing and possible representing the given metrics.
type MetricsBackend interface {
	// TotalChannels accept total number of lightning network payment
	// channels, with the given status and state.
	TotalChannels(asset, status, state string, channels int)

	// TotalUsers accept total number of users that are connected to us with
	// payment channels.
	TotalUsers(asset string, users int)

	// TotalFundsLockedRemotely accept accumulative number of funds locked with
	// us from user side.
	TotalFundsLockedRemotely(asset string, funds uint64)

	// TotalFundsLockedLocally accept accumulative number of funds locked by
	// us, i.e our lightning network node.
	TotalFundsLockedLocally(asset string, funds uint64)

	// TotalFreeFunds accept total number of funds free and under our
	// lightning node control.
	TotalFreeFunds(asset string, funds uint64)

	// AddSuccessfulForwardingPayment increment number of successful transition
	// payments through our lightning network node.
	AddSuccessfulForwardingPayment(asset string)

	// AddEarnedFunds increment number of funds which we earned by forwarding
	// payments.
	AddEarnedFunds(asset string, earned uint64)
}

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
	channelsCurrent *prometheus.GaugeVec
	usersCurrent    *prometheus.GaugeVec

	lockedFundsCurrent *prometheus.GaugeVec
	freeFundsCurrent   *prometheus.GaugeVec
	earnedFundsTotal   *prometheus.CounterVec

	forwardingPaymentTotal *prometheus.CounterVec
}

// TotalChannels accept total number of lightning network payment
// channels, with the given status and state.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) TotalChannels(asset, status, state string,
	channels int) {
	m.channelsCurrent.With(prometheus.Labels{
		assetLabel:         asset,
		channelStatusLabel: status,
		channelStateLabel:  state,
	}).Set(float64(channels))
}

// TotalUsers accept total number of users that are connected to us with
// payment channels.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) TotalUsers(asset string, users int) {
	m.usersCurrent.With(prometheus.Labels{
		assetLabel: asset,
	}).Set(float64(users))
}

// TotalFundsLockedRemotely accept accumulative number of funds locked with
// us from user side.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) TotalFundsLockedRemotely(asset string, funds uint64) {
	m.lockedFundsCurrent.With(prometheus.Labels{
		assetLabel: asset,
		sideLabel:  "remote",
	}).Set(float64(funds))
}

// TotalFundsLockedLocally accept accumulative number of funds locked by
// us, i.e. lightning network node.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) TotalFundsLockedLocally(asset string, funds uint64) {
	m.lockedFundsCurrent.With(prometheus.Labels{
		assetLabel: asset,
		sideLabel:  "local",
	}).Set(float64(funds))
}

// TotalFreeFunds accept total number of funds free and under our lightning
// network node control.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) TotalFreeFunds(asset string, funds uint64) {
	m.freeFundsCurrent.With(prometheus.Labels{
		assetLabel: asset,
	}).Set(float64(funds))
}

// AddSuccessfulForwardingPayment increment number of successful transition
// payments through our lightning network node.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) AddSuccessfulForwardingPayment(asset string) {
	m.forwardingPaymentTotal.With(prometheus.Labels{
		assetLabel: asset,
	}).Add(1)
}

// AddEarnedFunds increment number of funds which we earned by forwarding
// payments.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) AddEarnedFunds(asset string, earned uint64) {
	m.earnedFundsTotal.With(prometheus.Labels{
		assetLabel: asset,
	}).Add(float64(earned))
}

// InitMetricsBackend creates subsystem metrics for specified
// net. Creates and tries to register metrics singletons. If register was
// already done, than return error.
func InitMetricsBackend(net string) (MetricsBackend, error) {
	backend := PrometheusBackend{}

	backend.channelsCurrent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystem,
			Name:      "channels_current",
			Help: "Current channels connected to our lightning network" +
				" node",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			assetLabel,
			channelStatusLabel,
			channelStateLabel,
		},
	)

	if err := prometheus.Register(backend.channelsCurrent); err != nil {
		return nil, errors.Errorf(
			"unable to register 'channelsCurrent' metric:" +
				err.Error())

	}

	backend.usersCurrent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystem,
			Name:      "users_current",
			Help:      "Current users connected to our lightning network node",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			assetLabel,
		},
	)

	if err := prometheus.Register(backend.usersCurrent); err != nil {
		return nil, errors.Errorf(
			"unable to register 'usersCurrent' metric: " +
				err.Error())

	}

	backend.lockedFundsCurrent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystem,
			Name:      "locked_funds_current",
			Help: "Current funds locked in local network of our" +
				" lightning network node",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			assetLabel,
			sideLabel,
		},
	)

	if err := prometheus.Register(backend.lockedFundsCurrent); err != nil {
		return nil, errors.Errorf(
			"unable to register 'lockedFundsCurrent' metric: " +
				err.Error())
	}

	backend.freeFundsCurrent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystem,
			Name:      "free_funds_current",
			Help: "Current funds free under control of our lightning" +
				" network node",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			assetLabel,
		},
	)

	if err := prometheus.Register(backend.freeFundsCurrent); err != nil {
		return nil, errors.Errorf(
			"unable to register 'freeFundsCurrent' metric: " +
				err.Error())
	}

	backend.earnedFundsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystem,
			Name:      "earned_total",
			Help: "Total funds earned under control of our lightning" +
				" network node",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			assetLabel,
		},
	)

	if err := prometheus.Register(backend.earnedFundsTotal); err != nil {
		return nil, errors.Errorf(
			"unable to register 'earnedFundsTotal' metric: " +
				err.Error())
	}

	backend.forwardingPaymentTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystem,
			Name:      "forwards_total",
			Help: "Total payments forwarded through the our lightning" +
				" network node",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			assetLabel,
		},
	)

	if err := prometheus.Register(backend.forwardingPaymentTotal); err != nil {
		return nil, errors.Errorf(
			"unable to register 'forwardingPaymentTotal' metric: " +
				err.Error())
	}

	return backend, nil
}
