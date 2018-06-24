package optimisation

import (
	"github.com/bitlum/hub/manager/router"
	"github.com/bitlum/btcutil"
	"time"
	"strings"
)

type link struct {
	routerFunds router.BalanceUnit
	usersFunds  router.BalanceUnit
	IsActive    bool
}

// EnableChannelBalancing implements primitive channel balance. Is an goroutine
// which listens for change of topology of router and react on it accordingly.
// In current implementation its creating the channel to user if we have less
// than some threshold of number of funds.
func EnableChannelBalancing(r router.Router) {
	go func() {
		notSupportMultipleChannels := make(map[router.UserID]struct{})

		for {
			channels, err := r.Channels()
			if err != nil {
				log.Errorf("unable to fetch network: %v", err)
				<-time.After(time.Second * 10)
				continue
			}

			links := make(map[router.UserID]link)

			for _, channel := range channels {
				if _, ok := notSupportMultipleChannels[channel.UserID]; ok {
					continue
				}

				l, ok := links[channel.UserID]
				if !ok {
					l = link{}
				}

				l.routerFunds += channel.RouterBalance
				l.usersFunds += channel.UserBalance
				l.IsActive = l.IsActive || channel.IsActive()
				links[channel.UserID] = l
			}

			threshold := 0.2
			for userID, l := range links {
				if !l.IsActive {
					continue
				}

				// If ratio of outer funds less than users funds than we
				// should create additional channel.
				ratio := float64(l.routerFunds) / float64(l.usersFunds)

				log.Debugf("Calculate %v ratio with user %v", ratio, userID)

				if ratio < threshold {
					additionalFunds := l.usersFunds - l.routerFunds

					go func(userID router.UserID, additionalFunds router.BalanceUnit) {
						log.Infof("Trying to create additional channel"+
							" with user(%v), additional funds(%v)", userID,
							btcutil.Amount(additionalFunds))

						if err := r.OpenChannel(userID, additionalFunds); err != nil {
							if strings.Contains(err.Error(), "Multiple channels unsupported") {
								// Skip nodes without support of multiple
								// channels.
								log.Infof("Remove link for user(%v) which do not support multiple"+
									" channels", userID)
								notSupportMultipleChannels[userID] = struct{}{}
								return
							}

							log.Errorf("unable to create additional "+
								"channel with user: %v", err)
						}
					}(userID, additionalFunds)
				}
			}

			<-time.After(time.Second * 10)
		}
	}()
}
