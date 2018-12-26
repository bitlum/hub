package stats

import (
	"github.com/bitlum/hub/lightning"
	"github.com/btcsuite/btcutil"
	"github.com/go-errors/errors"
)

// ChannelNodeStats is the channels statistic attached to particular node.
type ChannelNodeStats struct {
	// LockedLocallyActive is number of funds aggregated from all channels,
	// which could be used for send the payments.
	LockedLocallyActive btcutil.Amount

	// LockedRemotelyActive is number of funds aggregated from all channels,
	// which could be used for receiving the payments.
	LockedRemotelyActive btcutil.Amount

	// LockedLocallyOverall is number of funds aggregated from all channels,
	// which are locked on local side. This number include pending, as well
	// as active funds.
	LockedLocallyOverall btcutil.Amount

	// LockedRemotelyOverall is number of funds aggregated from all channels,
	// which are locked on remote side. This number include pending, as well
	// as active funds.
	LockedRemotelyOverall btcutil.Amount
}

func calculateChannelNodeStats(channels []*lightning.Channel) (
	map[lightning.NodeID]ChannelNodeStats, error) {
	nodeStats := make(map[lightning.NodeID]ChannelNodeStats)

	for _, channel := range channels {
		currentState, ok := channel.States[channel.State]
		if !ok {
			return nil, errors.Errorf("unable to get state(%v) for"+
				"channel(%v)", channel.State, channel.ChannelID)
		}

		stat := nodeStats[channel.NodeID]

		// Check last / current channel state
		switch s := currentState.(type) {
		case *lightning.ChannelStateOpened:
			if channel.IsActive() {
				stat.LockedLocallyActive += s.LocalBalance
				stat.LockedRemotelyActive += s.RemoteBalance
			}

			stat.LockedLocallyOverall += s.LocalBalance
			stat.LockedRemotelyOverall += s.RemoteBalance

		case *lightning.ChannelStateOpening:
			stat.LockedLocallyOverall += s.LocalBalance
			stat.LockedRemotelyOverall += s.RemoteBalance
		}

		nodeStats[channel.NodeID] = stat
	}

	return nodeStats, nil
}

// ChannelFeeReport is the report about fee spending system have to make over
// the specified period of time.
type ChannelFeeReport struct {
	// StartTime and endTIme define period for which stats has to be
	// aggregated.
	StartTime int64
	EndTime   int64

	// OpenChannelFee is the number of funds lightning network node has
	// spent on opening channels, during specified period of time.
	OpenChannelFee btcutil.Amount

	// OpenChannels is channels which were included in
	// calculating open fee statistic.
	OpenChannels []*lightning.Channel

	// CloseChannelFee is the number of funds lightning network node has
	// spent on closing channels, during specified period of time.
	CloseChannelFee btcutil.Amount

	// CloseChannels is channels which were included in calculating close fee
	// statistic.
	CloseChannels []*lightning.Channel

	// HtlcSwipeFee number of funds lightning network node spend on
	// blockchain fees to swipe pending htlc aka stuck payments, during
	// specified period of time.
	HtlcSwipeFee btcutil.Amount
}

// ChannelsOverallStats is the overall channel system statistic.
type ChannelsOverallStats struct {
	// CurrentCommitFee is the estimated number of funds we need to spent to
	// cooperatively close all channels.
	CurrentCommitFee btcutil.Amount

	// CurrentLimboBalance is the number of funds which stuck in closing state.
	CurrentLimboBalance btcutil.Amount

	// CurrentStuckBalance is the number of funds which stuck in pending
	// payments.
	CurrentStuckBalance btcutil.Amount
}

// GetChannelFeeSpendingReport returns number of funds spent on channel
// management over specified interval of time.
func GetChannelFeeSpendingReport(startTime, endTime int64,
	channels []*lightning.Channel) (*ChannelFeeReport, error) {

	feeStats := &ChannelFeeReport{
		StartTime: startTime,
		EndTime:   endTime,
	}

	for _, channel := range channels {
		stateName := channel.CurrentState()

		switch stateName {
		case lightning.ChannelOpening,
			lightning.ChannelOpened:

			// We track fee on the moment they were spent from our
			// wallet, which is the time when channel started opening,
			// not on the time when channel has been opened.
			openingTime, err := channel.OpeningTime()
			if err != nil {
				return nil, errors.Errorf("unable get channel(%v) opening "+
					"time: %v", channel.ChannelID, err)
			}

			// Check that open time lies withing inspected period.
			if openingTime > startTime && openingTime < endTime {
				openTime, err := channel.OpenFee()
				if err != nil {
					return nil, errors.Errorf("unable get channel(%v) open "+
						"fee: %v", channel.ChannelID, err)
				}

				feeStats.OpenChannelFee += openTime
				feeStats.OpenChannels = append(feeStats.OpenChannels, channel)
			}

		case lightning.ChannelClosed,
			lightning.ChannelClosing:

			// We track fee on the moment they were spent from our
			// wallet, which is the time when channel started closing,
			// not on the time when channel has been closed.
			openTime, err := channel.OpeningTime()
			if err != nil {
				return nil, errors.Errorf("unable get channel(%v) open "+
					"time: %v", channel.ChannelID, err)
			}

			// Check that open time lies withing inspected period.
			if openTime > startTime && openTime < endTime {
				openFee, err := channel.OpenFee()
				if err != nil {
					return nil, errors.Errorf("unable get channel(%v) open "+
						"fee: %v", channel.ChannelID, err)
				}

				feeStats.OpenChannelFee += openFee
				feeStats.OpenChannels = append(feeStats.OpenChannels, channel)
			}

			closingTime, err := channel.ClosingTime()
			if err != nil {
				return nil, errors.Errorf("unable get channel(%v) close "+
					"time: %v", channel.ChannelID, err)
			}

			// Calculate final fee only after channel has been closed
			if closingTime > startTime && closingTime < endTime {
				closeFee, err := channel.CloseFee()
				if err != nil {
					return nil, errors.Errorf("unable get channel(%v) close "+
						"fee: %v", channel.ChannelID, err)
				}

				feeStats.CloseChannelFee += closeFee

				swipeFee, err := channel.SwipeFee()
				if err != nil {
					return nil, errors.Errorf("unable get channel(%v) close "+
						"swipe fee: %v", channel.ChannelID, err)
				}

				feeStats.HtlcSwipeFee += swipeFee
				feeStats.CloseChannels = append(feeStats.CloseChannels, channel)
			}
		}
	}

	return feeStats, nil
}

// GetChannelOverallStats returns current state of channels not
// attached to particular to node, but in general instead.
func GetChannelOverallStats(channels []*lightning.Channel) (
	*ChannelsOverallStats, error) {

	stats := &ChannelsOverallStats{}

	for _, channel := range channels {
		stateName := channel.CurrentState()

		switch stateName {
		case lightning.ChannelOpening:
			commitFee, err := channel.CommitFee()
			if err != nil {
				return nil, errors.Errorf("unable get channel(%v) commit "+
					"fee: %v", channel.ChannelID, err)
			}

			stats.CurrentCommitFee += commitFee

		case lightning.ChannelOpened:
			commitFee, err := channel.CommitFee()
			if err != nil {
				return nil, errors.Errorf("unable get channel(%v) commit "+
					"fee: %v", channel.ChannelID, err)
			}

			stats.CurrentCommitFee += commitFee

			stuckBalance, err := channel.StuckBalance()
			if err != nil {
				return nil, errors.Errorf("unable get channel(%v) stuck"+
					"balance: %v", channel.ChannelID, err)
			}

			stats.CurrentStuckBalance += stuckBalance

		case lightning.ChannelClosing:
			limboBalance, err := channel.LimboBalance()
			if err != nil {
				return nil, errors.Errorf("unable get channel(%v) "+
					"limbo balance: %v", channel.ChannelID, err)
			}

			stats.CurrentLimboBalance += limboBalance

		case lightning.ChannelClosed:
			// nothing to do because channel was already closed,
			// and limbo balanced is returned back.
		}
	}

	return stats, nil
}
