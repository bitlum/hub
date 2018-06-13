package processing

import (
	"time"
	"sync"
	"github.com/bitlum/hub/manager/metrics/network"
	"github.com/bitlum/hub/manager/router"
	"github.com/bitlum/hub/manager/router/registry"
	"github.com/bitlum/hub/manager/common/broadcast"
)

type Config struct {
	// Router is the entity which gives us glimpse of information about
	// lightning network.
	Router router.Router

	// MetricsBackend is used to send metrics about state of hub in the
	// monitoring subsystem.
	MetricsBackend network.MetricsBackend

	// InfoStorage is a place where stats gatherer could offload calculated info
	// for farther retrieval.
	InfoStorage router.InfoStorage
}

// Stats is an entity which is used for gathering the statistic
// about the router and it local network state, and
type Stats struct {
	cfg  *Config
	quit chan struct{}
	wg   sync.WaitGroup

	broadcast     *broadcast.Broadcaster
	metricsTicker *time.Ticker
}

func NewStats(cfg *Config) *Stats {
	return &Stats{
		cfg:           cfg,
		metricsTicker: time.NewTicker(time.Second * 10),
		quit:          make(chan struct{}),
	}
}

func (s *Stats) Start() {
	s.wg.Add(1)

	go s.scrapeRouterInfo()
}

func (s *Stats) Stop() {
	s.metricsTicker.Stop()
	close(s.quit)
	s.wg.Wait()
}

// scrapeRouterInfo subscribes on router updates and scrapes some additional
// topology information in order to place in monitoring subsystem.
//
// NOTE: Should run as goroutine.
func (s *Stats) scrapeRouterInfo() {
	defer func() {
		s.wg.Done()
		log.Info("Router network metrics gatherer goroutine stopped")
	}()

	log.Info("Router network metrics gatherer goroutine started")

	receiver := s.cfg.Router.RegisterOnUpdates()
	defer receiver.Stop()

	for {
		select {
		case <-s.metricsTicker.C:
			channels, err := s.cfg.Router.Network()
			if err != nil {
				log.Errorf("unable to fetch network: %v", err)
				continue
			}

			freeBalance, err := s.cfg.Router.FreeBalance()
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
				if channel.IsPending() {
					numPendingChannels++
					continue
				}

				totalLockedByUsers += channel.UserBalance
				totalLockedByRouter += channel.RouterBalance

				if !channel.IsActive() {
					numNonActiveChannels++
					continue
				} else {
					numActiveChannels++
				}

				if _, ok := users[channel.UserID]; !ok {
					users[channel.UserID] = &router.DbPeer{
						PubKey: string(channel.UserID),
						Alias:  registry.GetAlias(channel.UserID),
					}
				}

				users[channel.UserID].LockedByHub += int64(channel.RouterBalance)
				users[channel.UserID].LockedByPeer += int64(channel.UserBalance)
			}

			asset := s.cfg.Router.Asset()

			log.Debugf("Total pending channels: %v", numPendingChannels)
			s.cfg.MetricsBackend.TotalChannels(asset, "pending",
				"inactive", numPendingChannels)

			log.Debugf("Total open active channels: %v", numActiveChannels)
			s.cfg.MetricsBackend.TotalChannels(asset, "opened", "active", numActiveChannels)

			log.Debugf("Total open inactive channels: %v", numNonActiveChannels)
			s.cfg.MetricsBackend.TotalChannels(asset, "opened", "inactive", numNonActiveChannels)

			log.Debugf("Total connected users: %v", len(users))
			s.cfg.MetricsBackend.TotalUsers(asset, len(users))

			log.Debugf("Funds locked by users: %v", totalLockedByUsers)
			s.cfg.MetricsBackend.TotalFundsLockedByUser(asset, uint64(totalLockedByUsers))

			log.Debugf("Funds locked by router: %v", totalLockedByRouter)
			s.cfg.MetricsBackend.TotalFundsLockedByRouter(asset, uint64(totalLockedByRouter))

			log.Debugf("Free funds: %v", freeBalance)
			s.cfg.MetricsBackend.TotalFreeFunds(asset, uint64(freeBalance))

			var peers []*router.DbPeer
			for _, u := range users {
				peers = append(peers, u)
			}

			if err := s.cfg.InfoStorage.UpdatePeers(peers); err != nil {
				log.Errorf("unable to save peers: %v", err)
				continue
			}

		case update := <-receiver.Read():
			switch u := update.(type) {
			case *router.UpdatePayment:
				if u.Type == router.Forward {
					asset := s.cfg.Router.Asset()
					s.cfg.MetricsBackend.AddSuccessfulForwardingPayment(asset)
					s.cfg.MetricsBackend.AddEarnedFunds(asset, uint64(u.Earned))
				}

				if err := s.cfg.InfoStorage.StorePayment(&router.DbPayment{
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

		case <-s.quit:
			return
		}
	}
}
