package stats

import (
	"time"
	"sync"
	"github.com/bitlum/hub/manager/metrics/network"
	"github.com/bitlum/hub/manager/router"
)

// NetworkStatsGatherer is an entity which is used for gathering the statistic
// about the router and it local network state.
type NetworkStatsGatherer struct {
	router router.Router

	metricsBackend network.MetricsBackend
	metricsTicker  *time.Ticker

	quit chan struct{}
	wg   sync.WaitGroup
}

func NewNetworkStatsGatherer(router router.Router,
	metrics network.MetricsBackend) *NetworkStatsGatherer {

	return &NetworkStatsGatherer{
		router:         router,
		metricsBackend: metrics,
		metricsTicker:  time.NewTicker(time.Second * 10),
		quit:           make(chan struct{}),
	}
}

func (m *NetworkStatsGatherer) Start() {
	m.wg.Add(1)

	go func() {
		defer func() {
			m.wg.Done()
			log.Info("Router network metrics gatherer goroutine stopped")
		}()

		log.Info("Router network metrics gatherer goroutine started")

		receiver := m.router.RegisterOnUpdates()
		defer receiver.Stop()

		for {
			select {
			case <-m.metricsTicker.C:
				channels, err := m.router.Network()
				if err != nil {
					log.Errorf("unable to fetch network: %v", err)
					continue
				}

				freeBalance, err := m.router.FreeBalance()
				if err != nil {
					log.Errorf("unable to fetch free balance: %v", err)
					continue
				}

				var (
					numPendingChannels   = 0
					numActiveChannels    = 0
					numNonActiveChannels = 0
					numLockedByUsers     = router.BalanceUnit(0)
					numLockedByRouter    = router.BalanceUnit(0)
				)

				users := make(map[string]struct{})

				for _, channel := range channels {
					if channel.IsPending {
						numPendingChannels++
						continue
					}

					users[string(channel.UserID)] = struct{}{}
					numLockedByUsers += channel.UserBalance
					numLockedByRouter += channel.RouterBalance

					if channel.IsActive {
						numActiveChannels++
						continue
					}

					numNonActiveChannels++
				}

				asset := m.router.Asset()

				log.Debugf("Total pending channels: %v", numPendingChannels)
				m.metricsBackend.TotalChannels(asset, "pending", "inactive", numPendingChannels)

				log.Debugf("Total open active channels: %v", numActiveChannels)
				m.metricsBackend.TotalChannels(asset, "opened", "active", numActiveChannels)

				log.Debugf("Total open inactive channels: %v", numNonActiveChannels)
				m.metricsBackend.TotalChannels(asset, "opened", "inactive", numNonActiveChannels)

				log.Debugf("Total connected users: %v", len(users))
				m.metricsBackend.TotalUsers(asset, len(users))

				log.Debugf("Funds locked by users: %v", numLockedByUsers)
				m.metricsBackend.TotalFundsLockedByUser(asset, uint64(numLockedByUsers))

				log.Debugf("Funds locked by router: %v", numLockedByRouter)
				m.metricsBackend.TotalFundsLockedByRouter(asset, uint64(numLockedByRouter))

				log.Debugf("Free funds: %v", freeBalance)
				m.metricsBackend.TotalFreeFunds(asset, uint64(freeBalance))

			case update := <-receiver.Read():
				switch u := update.(type) {
				case router.UpdatePayment:
					if u.Type == router.Forward {
						asset := m.router.Asset()
						m.metricsBackend.AddSuccessfulForwardingPayment(asset)
						m.metricsBackend.AddEarnedFunds(asset, uint64(u.Earned))
					}
				default:
					continue
				}

			case <-m.quit:
				return
			}
		}
	}()
}

func (m *NetworkStatsGatherer) Stop() {
	m.metricsTicker.Stop()
	close(m.quit)
	m.wg.Wait()
}
