package stats

import (
	"github.com/bitlum/hub/lightning"
	"sort"
)

type NodeStats struct {
	NodeID lightning.NodeID

	PaymentNodeStats
	ChannelNodeStats
}

type RankedStat struct {
	Rank float64
	NodeStats
}

// RankByPaymentSentNum...
func RankByPaymentSentNum(nodeStats map[lightning.NodeID]NodeStats) []RankedStat {
	var rankedNodes []RankedStat
	for _, stat := range nodeStats {
		rankedNodes = append(rankedNodes, RankedStat{
			Rank:      float64(stat.NumSentPayments),
			NodeStats: stat,
		})
	}

	sort.Slice(rankedNodes, func(i, j int) bool {
		return rankedNodes[i].Rank > rankedNodes[j].Rank
	})

	return rankedNodes
}

// RankByAveragePaymentSentFlow Rank nodes by the amount of funds sent to
// them in average during the over inspected period of time.
// First node in the list is most active one.
func RankByAveragePaymentSentFlow(nodeStats map[lightning.NodeID]NodeStats) []RankedStat {
	var rankedNodes []RankedStat
	for _, stat := range nodeStats {
		rankedNodes = append(rankedNodes, RankedStat{
			Rank:      float64(stat.AverageSentSat),
			NodeStats: stat,
		})
	}

	sort.Slice(rankedNodes, func(i, j int) bool {
		return rankedNodes[i].Rank > rankedNodes[j].Rank
	})

	return rankedNodes
}

// RankByPaymentVolume...
func RankByPaymentVolume(nodeStats map[lightning.NodeID]NodeStats) []RankedStat {
	var rankedNodes []RankedStat
	for _, stat := range nodeStats {
		rankedNodes = append(rankedNodes, RankedStat{
			Rank: float64(stat.AverageSentSat + stat.
				AverageReceivedForwardSat + stat.AverageReceivedForwardSat),
			NodeStats: stat,
		})
	}

	sort.Slice(rankedNodes, func(i, j int) bool {
		return rankedNodes[i].Rank > rankedNodes[j].Rank
	})

	return rankedNodes
}

// RankByIdleFunds sort nodes based on idleness of out funds locked with it,
// if we don't have any activity, and we have local funds,
// than we should release them first. First node in the list is most idle one.
func RankByIdleFunds(nodeStats map[lightning.NodeID]NodeStats) []RankedStat {
	var rankedNodes []RankedStat
	for _, stat := range nodeStats {
		overallFlow := stat.AverageSentSat + stat.AverageSentForwardSat +
			stat.AverageReceivedForwardSat

		if overallFlow == 0 {
			overallFlow = 1
		}

		// Calculate idle rank of node as a ratio of overall locked funds on
		// overall flow of funds.
		idleRank := (stat.LockedLocallyOverall + stat.
			LockedRemotelyOverall) / overallFlow

		rankedNodes = append(rankedNodes, RankedStat{
			Rank:      float64(idleRank),
			NodeStats: stat,
		})
	}

	sort.Slice(rankedNodes, func(i, j int) bool {
		return rankedNodes[i].Rank > rankedNodes[j].Rank
	})

	return rankedNodes
}

// RankByNeededAdditionalCapacity Rank nodes by the number of funds which has
// to be locked with them additional so that payments haven't failed.
func RankByNeededAdditionalCapacity(nodeStats map[lightning.NodeID]NodeStats) []RankedStat {
	var rankedNodes []RankedStat
	for _, stat := range nodeStats {
		sentFlow := stat.AverageSentSat + stat.AverageSentForwardSat -
			stat.AverageReceivedForwardSat

		if sentFlow < 0 {
			sentFlow = 0
		}

		// Calculate number of funds which has to be locked additionally in
		// this channel otherwise payment will start fail.
		neededToLockAdditional := sentFlow - stat.LockedLocallyOverall
		if neededToLockAdditional < 0 {
			neededToLockAdditional = 0
		}

		rankedNodes = append(rankedNodes, RankedStat{
			Rank:      float64(neededToLockAdditional),
			NodeStats: stat,
		})
	}

	sort.Slice(rankedNodes, func(i, j int) bool {
		return rankedNodes[i].Rank > rankedNodes[j].Rank
	})

	return rankedNodes
}
