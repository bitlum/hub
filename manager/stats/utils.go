package stats

import "github.com/bitlum/hub/lightning"

func GetNodeStats(period string, payments []*lightning.Payment,
	forwardPayments []*lightning.ForwardPayment,
	channels []*lightning.Channel) (map[lightning.NodeID]NodeStats, error) {

	nodeChannelStats, err := calculateChannelNodeStats(channels)
	if err != nil {
		return nil, err
	}

	paymentChannelStats, err := calculatePaymentsStats(period, payments,
		forwardPayments)
	if err != nil {
		return nil, err
	}

	nodeStats := make(map[lightning.NodeID]NodeStats)

	for nodeID, stat := range nodeChannelStats {
		obj := nodeStats[nodeID]
		obj.ChannelNodeStats = stat
		obj.NodeID = nodeID
		nodeStats[nodeID] = obj
	}

	for nodeID, stat := range paymentChannelStats {
		obj := nodeStats[nodeID]
		obj.PaymentNodeStats = stat
		obj.NodeID = nodeID
		nodeStats[nodeID] = obj
	}

	return nodeStats, nil
}
