package optimisation

import (
	"github.com/bitlum/hub/manager/router"
	"github.com/bitlum/btcutil"
	"time"
	"strings"
	"sync"
)

// link represent accumulative active funds in the channels with user
type link struct {
	// routerFunds is either funds which resides in open channel or in
	// pending open channel locked from the side of router.
	routerFunds router.BalanceUnit

	// usersFunds is either funds which resides in open channel or in
	// pending open channel locked from the side of usersFunds.
	usersFunds router.BalanceUnit
}

// Balancer implements primitive channel balancing. It listens for change of
// topology of router and react on it accordingly. In current implementation
// its creating the channel to user if we have less than some threshold
// of number of funds.
type Balancer struct {
	threshold                  float64
	router                     router.Router
	notSupportMultipleChannels map[router.UserID]struct{}
	repeateCycleAfter          time.Duration

	wg   sync.WaitGroup
	quit chan struct{}
}

// NewBalancer creates new instance of Balancer.
func NewBalancer(r router.Router, threshold float64,
	repeateCycleAfter time.Duration) *Balancer {
	return &Balancer{
		threshold:                  threshold,
		router:                     r,
		quit:                       make(chan struct{}),
		notSupportMultipleChannels: make(map[router.UserID]struct{}),
		repeateCycleAfter:          repeateCycleAfter,
	}
}

func (b *Balancer) Start() {
	b.wg.Add(1)
	go b.enableChannelBalancing()
}

func (b *Balancer) Stop() {
	close(b.quit)
	b.wg.Wait()
}

// enableChannelBalancing listens for change of topology of router and tries
// to create additional channel with user if our balances with him are
// different in the level of given threshold.
func (b *Balancer) enableChannelBalancing() {
	defer func() {
		b.wg.Done()
		log.Infof("Stop balancing goroutine")
	}()

	log.Infof("Start balancing goroutine")

	for {
		select {
		case <-time.After(b.repeateCycleAfter):
		case <-b.quit:
			return
		}

		// Fetch last router view on our topology
		channels, err := b.router.Channels()
		if err != nil {
			log.Errorf("unable to fetch network: %v", err)
			continue
		}

		links := make(map[router.UserID]link)
		for _, channel := range channels {
			// Skip users which do not support multiple channels in order to
			// avoid bombarding them with open channel requests.
			if _, ok := b.notSupportMultipleChannels[channel.UserID]; ok {
				continue
			}

			// If user is not connected than we wouldn't be able to connect
			// to him, so skipping.
			if !channel.IsUserConnected {
				continue
			}

			l, ok := links[channel.UserID]
			if !ok {
				l = link{}
			}

			// Add pending to avoid repeated channel creation in next balancing
			// cycle.
			switch channel.CurrentState().Name {
			case router.ChannelOpened:
			case router.ChannelUpdating:
			case router.ChannelOpening:
				l.routerFunds += channel.RouterBalance
				l.usersFunds += channel.UserBalance
			}

			links[channel.UserID] = l
		}

		for userID, l := range links {
			// If ratio of outer funds less than users funds than we
			// should create additional channel.
			ratio := float64(l.routerFunds) / float64(l.usersFunds)

			log.Debugf("Calculate %v ratio with user(%v)", ratio, userID)

			if ratio < b.threshold {
				additionalFunds := l.usersFunds - l.routerFunds
				go b.createAdditionalChannel(userID, additionalFunds)
			}
		}
	}
}

func (b *Balancer) createAdditionalChannel(userID router.UserID,
	additionalFunds router.BalanceUnit) {
	log.Infof("Trying to create additional channel"+
		" with user(%v), additional funds(%v)", userID,
		btcutil.Amount(additionalFunds))

	if err := b.router.OpenChannel(userID, additionalFunds); err != nil {
		if strings.Contains(err.Error(), "Multiple channels unsupported") {
			// Skip nodes without support of multiple
			// channels.
			log.Infof("Avoid user(%v) which do not support multiple"+
				" channels", userID)
			b.notSupportMultipleChannels[userID] = struct{}{}
			return
		}

		log.Errorf("unable to create additional "+
			"channel with user: %v", err)
	}
}
