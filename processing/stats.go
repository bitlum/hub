package processing

import (
	"github.com/bitlum/hub/common/broadcast"
	"github.com/bitlum/hub/lightning"
	"github.com/bitlum/hub/metrics/network"
	"sync"
	"time"
)

type Config struct {
	// Client is the entity which gives us glimpse of information about
	// lightning network.
	Client lightning.Client

	// MetricsBackend is used to send metrics about state of hub in the
	// monitoring subsystem.
	MetricsBackend network.MetricsBackend

	// Storage is a place where stats gatherer could offload calculated info
	// for farther retrieval.
	Storage lightning.UserStorage
}

// Stats is an entity which is used for gathering the statistic
// about the out lighting network node and it local network state, and
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

	go s.scrapeNodeInfo()
}

func (s *Stats) Stop() {
	s.metricsTicker.Stop()
	close(s.quit)
	s.wg.Wait()
}

// scrapeNodeInfo subscribes on lightning client updates and scrapes some
// additional topology information in order to place in monitoring subsystem.
//
// NOTE: Should run as goroutine.
func (s *Stats) scrapeNodeInfo() {
	defer func() {
		s.wg.Done()
		log.Info("Client network metrics gatherer goroutine stopped")
	}()

	log.Info("Client network metrics gatherer goroutine started")

	receiver := s.cfg.Client.RegisterOnUpdates()
	defer receiver.Stop()

	for {
		select {
		case <-s.metricsTicker.C:
			asset := s.cfg.Client.Asset()

			channels, err := s.cfg.Client.Channels()
			if err != nil {
				log.Errorf("unable to fetch network: %v", err)
				continue
			}

			numPendingChannels := 0
			numActiveChannels := 0
			numNonActiveChannels := 0

			for _, channel := range channels {
				if channel.IsPending() {
					numPendingChannels++
				}

				if channel.IsActive() {
					numActiveChannels++
				} else {
					numNonActiveChannels++
				}
			}

			log.Debugf("Total pending channels: %v", numPendingChannels)
			s.cfg.MetricsBackend.TotalChannels(asset, "pending",
				"inactive", numPendingChannels)

			log.Debugf("Total open active channels: %v", numActiveChannels)
			s.cfg.MetricsBackend.TotalChannels(asset, "opened", "active", numActiveChannels)

			log.Debugf("Total open inactive channels: %v", numNonActiveChannels)
			s.cfg.MetricsBackend.TotalChannels(asset, "opened", "inactive", numNonActiveChannels)

			// TODO(andrew.shvv) Remove when lightning client would return users
			users, err := s.cfg.Storage.Users()
			if err != nil {
				log.Errorf("unable to fetch users: %v", err)
				continue
			}

			totalLockedRemotely := lightning.BalanceUnit(0)
			totalLockedLocally := lightning.BalanceUnit(0)

			activeUsers := 0
			for _, user := range users {
				if user.IsConnected {
					activeUsers += 1
				}

				totalLockedRemotely += user.LockedByUser
				totalLockedLocally += user.LockedByHub
			}

			log.Debugf("Total connected users: %v", activeUsers)
			s.cfg.MetricsBackend.TotalUsers(asset, activeUsers)

			log.Debugf("Funds locked remotely: %v", totalLockedRemotely)
			s.cfg.MetricsBackend.TotalFundsLockedRemotely(asset, uint64(totalLockedRemotely))

			log.Debugf("Funds locked locally: %v", totalLockedLocally)
			s.cfg.MetricsBackend.TotalFundsLockedLocally(asset, uint64(totalLockedLocally))

			freeBalance, err := s.cfg.Client.FreeBalance()
			if err != nil {
				log.Errorf("unable to fetch free balance: %v", err)
				continue
			}

			log.Debugf("Free funds: %v", freeBalance)
			s.cfg.MetricsBackend.TotalFreeFunds(asset, uint64(freeBalance))

		case update := <-receiver.Read():
			switch u := update.(type) {
			case *lightning.UpdatePayment:
				if u.Type == lightning.Forward {
					asset := s.cfg.Client.Asset()
					s.cfg.MetricsBackend.AddSuccessfulForwardingPayment(asset)
					s.cfg.MetricsBackend.AddEarnedFunds(asset, uint64(u.Earned))
				}
			default:
				continue
			}

		case <-s.quit:
			return
		}
	}
}
