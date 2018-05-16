package stats

import (
	"time"
	"sync"
	"github.com/bitlum/hub/manager/metrics/network"
	"github.com/bitlum/hub/manager/router"
	"github.com/bitlum/hub/manager/router/registry"
)

type Config struct {
	Router         router.Router
	MetricsBackend network.MetricsBackend
	Storage        router.InfoStorage
}

// NetworkStatsGatherer is an entity which is used for gathering the statistic
// about the router and it local network state.
type NetworkStatsGatherer struct {
	cfg  *Config
	quit chan struct{}
	wg   sync.WaitGroup

	metricsTicker *time.Ticker
}

func NewNetworkStatsGatherer(cfg *Config) *NetworkStatsGatherer {
	return &NetworkStatsGatherer{
		cfg:           cfg,
		metricsTicker: time.NewTicker(time.Second * 10),
		quit:          make(chan struct{}),
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

		receiver := m.cfg.Router.RegisterOnUpdates()
		defer receiver.Stop()

		for {
			select {
			case <-m.metricsTicker.C:
				channels, err := m.cfg.Router.Network()
				if err != nil {
					log.Errorf("unable to fetch network: %v", err)
					continue
				}

				freeBalance, err := m.cfg.Router.FreeBalance()
				if err != nil {
					log.Errorf("unable to fetch free balance: %v", err)
					continue
				}

				var (
					numPendingChannels   = 0
					numActiveChannels    = 0
					numNonActiveChannels = 0
					totalLockedByUsers   = router.BalanceUnit(0)
					totalLockedByRouter  = router.BalanceUnit(0)
				)

				users := make(map[router.UserID]*router.DbPeer)
				for _, channel := range channels {
					if channel.IsPending {
						numPendingChannels++
						continue
					}

					totalLockedByUsers += channel.UserBalance
					totalLockedByRouter += channel.RouterBalance

					if !channel.IsActive {
						numNonActiveChannels++
						continue
					} else {
						numActiveChannels++
					}

					if _, ok := users[channel.UserID]; !ok {
						users[channel.UserID] = &router.DbPeer{
							Alias: registry.GetAlias(channel.UserID),
						}
					}

					users[channel.UserID].LockedByHub += int64(channel.RouterBalance)
					users[channel.UserID].LockedByPeer += int64(channel.UserBalance)
				}

				asset := m.cfg.Router.Asset()

				log.Debugf("Total pending channels: %v", numPendingChannels)
				m.cfg.MetricsBackend.TotalChannels(asset, "pending",
					"inactive", numPendingChannels)

				log.Debugf("Total open active channels: %v", numActiveChannels)
				m.cfg.MetricsBackend.TotalChannels(asset, "opened", "active", numActiveChannels)

				log.Debugf("Total open inactive channels: %v", numNonActiveChannels)
				m.cfg.MetricsBackend.TotalChannels(asset, "opened", "inactive", numNonActiveChannels)

				log.Debugf("Total connected users: %v", len(users))
				m.cfg.MetricsBackend.TotalUsers(asset, len(users))

				log.Debugf("Funds locked by users: %v", totalLockedByUsers)
				m.cfg.MetricsBackend.TotalFundsLockedByUser(asset, uint64(totalLockedByUsers))

				log.Debugf("Funds locked by router: %v", totalLockedByRouter)
				m.cfg.MetricsBackend.TotalFundsLockedByRouter(asset, uint64(totalLockedByRouter))

				log.Debugf("Free funds: %v", freeBalance)
				m.cfg.MetricsBackend.TotalFreeFunds(asset, uint64(freeBalance))

				var peers []*router.DbPeer
				for _, u := range users {
					peers = append(peers, u)
				}

				if err := m.cfg.Storage.UpdatePeers(peers); err != nil {
					log.Errorf("unable to save peers: %v", err)
					continue
				}

			case update := <-receiver.Read():
				switch u := update.(type) {
				case *router.UpdatePayment:
					if u.Type == router.Forward {
						asset := m.cfg.Router.Asset()
						m.cfg.MetricsBackend.AddSuccessfulForwardingPayment(asset)
						m.cfg.MetricsBackend.AddEarnedFunds(asset, uint64(u.Earned))
					}

					if err := m.cfg.Storage.StorePayment(&router.DbPayment{
						FromPeer: registry.GetAlias(u.Sender),
						ToPeer:   registry.GetAlias(u.Receiver),
						Amount:   int64(u.Amount),
						Type:     string(u.Type),
						Status:   string(u.Status),
						Time:     time.Now().Unix(),
					}); err != nil {
						log.Errorf("unable to save the payment: %v", err)
						continue
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
