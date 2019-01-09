package manager

import (
	"github.com/bitlum/hub/common"
	"github.com/bitlum/hub/lightning"
	"github.com/bitlum/hub/manager/stats"
	"github.com/bitlum/hub/metrics"
	"github.com/bitlum/hub/metrics/crypto"
	"github.com/btcsuite/btcutil"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-errors/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	// Client is the entity which gives us glimpse of information about
	// lightning network.
	Client lightning.Client

	// MetricsBackend is used to send metrics about state of hub in the
	// monitoring subsystem.
	MetricsBackend crypto.MetricsBackend

	// BitconPriceUSD is current bitcoin price which is used for calculation of
	// minimum and maximum channel size in bitcoin.
	GetBitcoinPriceUSD func() (float64, error)

	// MaxChannelSizeUSD represent maximum channel size we expect to create with
	// important nodes.
	MaxChannelSizeUSD float64

	// MinChannelSizeUSD represent minimal channel size in dollars,
	// if chanel which has to be created to important node is less in size
	// than minimum, that channel size will be bumped up to minimal amount.
	// This amount is needed to avoid creation of dust channels, during
	// channel management control. Lots of small channel could result in a lot
	// of fees we need to pay later for close of this channels.
	MinChannelSizeUSD float64

	// MaxCloseSpendingPerDayUSD is the expected amount of funds we could
	// allow to spent on channel close.
	MaxCloseSpendingPerDayUSD float64

	// MaxOpenSpendingPerDayUSD is the expected amount of funds we could
	// allow to spent on channel close.
	MaxOpenSpendingPerDayUSD float64

	// MaxCommitFeeUSD is the maximum number of fee we expect to pay for
	// closing all channels.
	MaxCommitFeeUSD float64

	// MaxTotalLimbo is the maximum number of balance which we might accept
	// as being in limbo.
	MaxLimboUSD float64

	// MaxStuckBalance is the maximum number of balance which we could accept
	// on being stuck as pending htlcs in channels.
	MaxStuckBalanceUSD float64

	OurNodeID lightning.NodeID
	OurName   string

	Asset string
}

// validate check that config is valid.
func (c *Config) validate() error {
	if c.Client == nil {
		return errors.New("lightning client should be specified")
	}

	if c.MetricsBackend == nil {
		return errors.New("metric backend should be specified")
	}

	if c.GetBitcoinPriceUSD == nil {
		return errors.New("bitcoin usd price func should be specified")
	}

	if c.Asset == "" {
		return errors.New("asset should be specified")
	}

	if c.OurNodeID == "" {
		return errors.New("our node id should be specified")
	}

	if c.OurName == "" {
		return errors.New("our node name should be specified")
	}

	if c.GetBitcoinPriceUSD == nil {
		return errors.New("get bitcoin price func should be specified")
	}

	if c.MaxChannelSizeUSD == 0 {
		return errors.New("max channel size should be specified")
	}

	if c.MinChannelSizeUSD == 0 {
		return errors.New("min channel size should be specified")
	}

	if c.MaxCloseSpendingPerDayUSD == 0 {
		return errors.New("max close spending should be specified")
	}

	if c.MaxOpenSpendingPerDayUSD == 0 {
		return errors.New("max open spending should be specified")
	}

	if c.MaxCommitFeeUSD == 0 {
		return errors.New("max commit fee should be specified")
	}

	if c.MaxLimboUSD == 0 {
		return errors.New("max limbo should be specified")
	}

	if c.MaxStuckBalanceUSD == 0 {
		return errors.New("max stuck balance be specified")
	}

	return nil
}

// Responsibilities:
// 1. Keep track of known lightning network nodes:
//	1.1 Allow third-party to add / remove this nodes.
//	1.2 If node was removed than we should remove channel.
//
// 2. Keep channel to them:
//	2.1 Predict number of funds which has to be locked and end send alerts if
//	we don't have them.
// 	2.2. Send alerts if we are not connected to known node with channel.
//
// 3. Expose info about connectivity to known nodes:
//	3.1 connected or not.
//	3.2 number of local / remote locked funds.
//	3.3 number of forward payment to them.
//	3.4 number of payments to them.
//	3.5 average amount flow to / from them per day.
//
// 4. Alert about strange changes:
// 	4.1 Channel with known node is closing for some unknown reason.
type NodeManager struct {
	started  int32
	shutdown int32
	quit     chan struct{}
	wg       sync.WaitGroup

	cfg *Config

	// importantNodes is a map of nodes to which we always have to keep
	// channel with, as well as execute additional checks which would ensure the
	// validity of cooperation between us and this node.
	importantNodes      map[lightning.NodeID]string
	importantNodesMutex sync.Mutex
}

// NewNodeManager creates new instance.
func NewNodeManager(cfg *Config) (*NodeManager, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &NodeManager{
		quit:           make(chan struct{}),
		cfg:            cfg,
		importantNodes: make(map[lightning.NodeID]string),
	}, nil
}

func (nm *NodeManager) Start() {
	if !atomic.CompareAndSwapInt32(&nm.started, 0, 1) {
		log.Warn("Node manager already started")
		return
	}

	nm.wg.Add(1)
	go func() {
		checkNodesTicker := time.NewTicker(time.Second * 25)
		reportDailyStatsTicker := time.NewTicker(time.Second * 25)

		defer func() {
			log.Info("Stopped checking connection with important nodes goroutine")
			nm.wg.Done()

			checkNodesTicker.Stop()
			reportDailyStatsTicker.Stop()
		}()

		log.Info("Started checking connection with important nodes goroutine")

		for {
			select {
			case <-checkNodesTicker.C:
				if err := nm.checkNodesAvailability(); err != nil {
					log.Errorf("unable to check nodes availability: %v", err)
					continue
				}
			case <-reportDailyStatsTicker.C:
				if err := nm.reportDailyStats(); err != nil {
					log.Errorf("unable to report daily stats: %v", err)
					continue
				}
			case <-nm.quit:
				return
			}
		}
	}()
}

// Stop gracefully stops the node manager.
func (nm *NodeManager) Stop(reason string) {
	if !atomic.CompareAndSwapInt32(&nm.shutdown, 0, 1) {
		log.Warn("Node manager already shutdown")
		return
	}

	close(nm.quit)
	nm.wg.Wait()

	log.Infof("Node manager shutdown, reason(%v)", reason)
}

func checkChannelsCountChange() {
	// Algorithm count of channels drop check:
	// Per 3 hours:
	// 1. Check count of channels.
	// 2. Check previous count of active channels.
	// 3. Send metric in prometheus
	// 4. If drop is 20% than alert.
	// 5. Who closed channels and why?
	// 6. Save current count of channel, and prev check time.
	// TODO(andrew.svv) implement

	// Algorithm capacity drop check:
	// Per 3 hours:
	// 1. Check overall capacity.
	// 2. Check prev overall capacity.
	// 3. Send metric in prometheus
	// 4. If drop 20% than alert.
	// 5. Who closed channels and why?
	// 6. Save current overall capacity, and prev check time.
	// TODO(andrew.svv) implement
}

// reportSuccessMetric reports ration of how many payment to important nodes
// goes over 1 hope versus multihop, which is good metric of success of
// node manager algo. Because if payment are going not through direct channels,
//	that means that something is wrong.
func (nm *NodeManager) reportSuccessMetric() {
	// TODO(andrew.svv) implement
}

// ...
func (nm *NodeManager) reportDailyStats() error {
	m := crypto.NewMetric(nm.cfg.Asset, common.GetFunctionName(),
		nm.cfg.MetricsBackend)
	defer m.Finish()

	day := int64(time.Hour.Seconds()) * 24
	end := time.Now().Unix()
	start := end - day

	channels, err := nm.cfg.Client.Channels()
	if err != nil {
		err := errors.Errorf("unable fetch channels: %v", err)
		m.AddError(metrics.HighSeverity)
		return err
	}

	spendingStats, err := stats.GetChannelFeeSpendingReport(start, end,
		channels)
	if err != nil {
		err := errors.Errorf("unable calculate channels stats: %v", err)
		m.AddError(metrics.HighSeverity)
		return err
	}

	overallStats, err := stats.GetChannelOverallStats(channels)
	if err != nil {
		err := errors.Errorf("unable calculate channels overall stats: %v", err)
		m.AddError(metrics.HighSeverity)
		return err
	}

	bitcoinPriceUSD, err := nm.cfg.GetBitcoinPriceUSD()
	if err != nil {
		err := errors.Errorf("unable get bitcoin price: %v", err)
		m.AddError(metrics.HighSeverity)
		return err
	}

	closeChannelFeeUSD := spendingStats.CloseChannelFee.ToBTC() * bitcoinPriceUSD
	htlcSwipeFeeUSD := spendingStats.HtlcSwipeFee.ToBTC() * bitcoinPriceUSD
	openChannelFeeUSD := spendingStats.OpenChannelFee.ToBTC() * bitcoinPriceUSD
	commitFeeUSD := overallStats.CurrentCommitFee.ToBTC() * bitcoinPriceUSD
	limboBalanceUSD := overallStats.CurrentLimboBalance.ToBTC() * bitcoinPriceUSD
	stuckBalanceUSD := overallStats.CurrentStuckBalance.ToBTC() * bitcoinPriceUSD

	// We should be aware that we spend more that we expect on closing channels.
	if closeChannelFeeUSD+htlcSwipeFeeUSD > nm.cfg.MaxCloseSpendingPerDayUSD {
		log.Warnf("Too much funds were spent on channel close, "+
			"max($ %v), current($ %v)", nm.cfg.MaxCloseSpendingPerDayUSD,
			closeChannelFeeUSD+htlcSwipeFeeUSD)

		for _, c := range spendingStats.CloseChannels {
			name := string(c.NodeID)
			if nodeName, ok := nm.importantNodes[c.NodeID]; ok {
				name = nodeName
			}

			log.Warnf("  Closed / closing channel node(%v), channelID(%v)",
				name, c.ChannelID)
		}

		m.AddError(metrics.HighSeverity)
	} else {
		log.Tracef("Close fee today, max($ %v), current($ %v)",
			nm.cfg.MaxCloseSpendingPerDayUSD, closeChannelFeeUSD+htlcSwipeFeeUSD)
	}

	// We should be aware that we spend more that we expect on opening channels.
	if openChannelFeeUSD > nm.cfg.MaxOpenSpendingPerDayUSD {
		log.Warnf("Too much funds were spent on channel open, "+
			"max($ %v), current($ %v)", nm.cfg.MaxOpenSpendingPerDayUSD,
			openChannelFeeUSD)

		for _, c := range spendingStats.OpenChannels {
			name := string(c.NodeID)
			if nodeName, ok := nm.importantNodes[c.NodeID]; ok {
				name = nodeName
			}

			log.Warnf("  Opened / opening channel, node(%v), channelID(%v)",
				name, c.ChannelID)
		}

		m.AddError(metrics.HighSeverity)
	} else {
		log.Tracef("Open fee, max($ %v), current($ %v)",
			nm.cfg.MaxOpenSpendingPerDayUSD, openChannelFeeUSD)
	}

	// We should be aware that we will spend more that we expect on channels
	// close.
	if commitFeeUSD > nm.cfg.MaxCommitFeeUSD {
		log.Warnf("Too high commit fee, max($ %v), current($ %v)",
			nm.cfg.MaxCommitFeeUSD, commitFeeUSD)
		m.AddError(metrics.HighSeverity)
	} else {
		log.Tracef("Commit fee, max($ %v), current($ %v)",
			nm.cfg.MaxCommitFeeUSD, commitFeeUSD)
	}

	// We should be aware of number of funds which are in limbo,
	// which means that they are awaiting to be returned back to the wallet.
	if limboBalanceUSD > nm.cfg.MaxLimboUSD {
		log.Warnf("Too high limbo balance, max($ %v), current($ %v)",
			nm.cfg.MaxLimboUSD, limboBalanceUSD)
		m.AddError(metrics.HighSeverity)
	} else {
		log.Tracef("Limbo balance, max($ %v), current($ %v)",
			nm.cfg.MaxLimboUSD, limboBalanceUSD)
	}

	// We should be aware that we have a lot of pending htlc,
	// because it might be forward htlc, and we wouldn't be able to catch them
	// in router.
	if stuckBalanceUSD > nm.cfg.MaxStuckBalanceUSD {
		log.Warnf("Too high stuck balance in pending htlc, max($ %v), "+
			"current($ %v)", nm.cfg.MaxStuckBalanceUSD, stuckBalanceUSD)
		m.AddError(metrics.HighSeverity)
	} else {
		log.Tracef("Stuck balance in pending htlc, max($ %v), "+
			"current($ %v)", nm.cfg.MaxStuckBalanceUSD, stuckBalanceUSD)
	}

	return nil
}

// checkMaxFlowAmount ensures that we don't have situations where we have
// bunch of small channels, but all payment to important node are big,
// which would result in fails, as far as lightning don't AMP yet.
func (nm *NodeManager) checkMaxFlowAmount() {
	// Check max flow to the node, as far we don't have AMP.
	// TODO(andrew.svv) implement
}

// checkNodesAvailability ensures that we always connected with lightning
// channels with important nodes, and prevent situation where we don't have
// enough funds in channel with this nodes.
func (nm *NodeManager) checkNodesAvailability() error {
	m := crypto.NewMetric(nm.cfg.Asset, common.GetFunctionName(),
		nm.cfg.MetricsBackend)
	defer m.Finish()

	// Connect to all important nodes to ensure that channels are active.
	for importantNodeID, nodeName := range nm.importantNodes {
		if err := nm.cfg.Client.ConnectToNode(importantNodeID); err != nil {
			m.AddError(metrics.HighSeverity)
			log.Warnf("unable to connect to important node(%v), id(%v): %v",
				nodeName, importantNodeID, err)
		} else {
			log.Debugf("Node(%v), id(%v) is connected", nodeName,
				importantNodeID)
		}
	}

	// Aggregate statistic over the last week, and calculate average payment
	// flow for one day.
	nodeStats, err := nm.GetNodeStats("week")
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return errors.Errorf("unable to calculate nodes statistics: %v", err)
	}

	// Ensure that important node is present in node stats.
	// This is needed if previously important node hasn't been seen at all,
	// not in payment, nor in channels.
	idCache := make(map[lightning.NodeID]struct{})
	for id, _ := range nodeStats {
		idCache[id] = struct{}{}
	}

	for nodeID, name := range nm.importantNodes {
		if _, ok := idCache[nodeID]; !ok {
			log.Warnf("Important node(%v), (%v) doesn't have any trace "+
				"of previous interaction with our node", name, nodeID)
			nodeStats[nodeID] = stats.NodeStats{NodeID: nodeID}
		}
	}

	// Calculate number of additional funds which should have with node
	// locally, in order to avoid payments to fail with it.
	rankedNodes := stats.RankByNeededAdditionalCapacity(nodeStats)

	// For every important node lets create channel if need to.
	for _, stat := range rankedNodes {
		nodeName, ok := nm.importantNodes[stat.NodeID]
		if !ok {
			continue
		}

		// We wouldn't be able to create channel with ourselves.
		if stat.NodeID == nm.cfg.OurNodeID {
			continue
		}

		additionalCapacity := btcutil.Amount(stat.Rank)
		if additionalCapacity == 0 {
			log.Debugf("Important node(%v) not requires additional capacity, "+
				"stats(%v)", nodeName, spew.Sdump(stat))
			continue
		}

		bitcoinPriceUSD, err := nm.cfg.GetBitcoinPriceUSD()
		if err != nil {
			err := errors.Errorf("unable get bitcoin price: %v", err)
			m.AddError(metrics.HighSeverity)
			return err
		}

		minChannelSizeSat := btcutil.Amount(nm.cfg.MinChannelSizeUSD /
			bitcoinPriceUSD * btcutil.SatoshiPerBitcoin)

		maxChannelSizeSat := btcutil.Amount(nm.cfg.MaxChannelSizeUSD /
			bitcoinPriceUSD * btcutil.SatoshiPerBitcoin)

		averageSentInUSD := stat.AverageSentSat.ToBTC() *
			bitcoinPriceUSD

		averageSentForwardInUSD := stat.AverageSentForwardSat.ToBTC() *
			bitcoinPriceUSD

		averageReceivedForwardInUSD := stat.AverageReceivedForwardSat.ToBTC() *
			bitcoinPriceUSD

		// In average during the day we send more than we have in our
		// channels locked locally. We have to create another channel,
		// otherwise payments might start fail.

		channelSizeSat := additionalCapacity
		if channelSizeSat < minChannelSizeSat {
			// This is needed to avoid creation of dust channels,
			// during channel management control. Lots of small channel
			// could result in a lot of fees we need to pay later for close
			// of this channels.
			channelSizeSat = minChannelSizeSat

			log.Infof("Node(%v), \n"+
				" average sent(%v USD), \n"+
				" average forwarded(%v USD), \n"+
				" average received forwarded(%v), \n"+
				" create channel with minimal size(%v), \n"+
				"stats(%v)",
				nodeName,
				averageSentInUSD, averageSentForwardInUSD,
				averageReceivedForwardInUSD, channelSizeSat, spew.Sdump(stat))

		} else if channelSizeSat > maxChannelSizeSat {
			// This is needed to just avoid unexpected situations,
			// as far algo is still in beta, for that reason lets report on
			// that.
			m.AddError(metrics.HighSeverity)
			channelSizeSat = maxChannelSizeSat

			log.Infof("Node(%v), \n"+
				" average sent(%v USD), \n"+
				" average forwarded(%v USD), \n"+
				" average received forwarded(%v USD), \n"+
				" create channel with maximum size(%v), \n"+
				" stats(%v)",
				nodeName,
				averageSentInUSD, averageSentForwardInUSD,
				averageReceivedForwardInUSD, channelSizeSat, spew.Sdump(stat))
		} else {
			log.Infof("Node(%v), \n"+
				" average sent(%v USD), \n"+
				" average forwarded(%v USD), \n"+
				" average received forwarded(%v USD), \n"+
				" create channel with size(%v), \n"+
				" stats(%v)",
				nodeName,
				averageSentInUSD, averageSentForwardInUSD,
				averageReceivedForwardInUSD, channelSizeSat, spew.Sdump(stat))
		}

		if err := nm.cfg.Client.OpenChannel(stat.NodeID, channelSizeSat); err != nil {
			err := errors.Errorf("unable open channel with node(%v) id("+
				"%v), amount(%v): %v", nodeName, stat.NodeID, channelSizeSat, err)
			m.AddError(metrics.HighSeverity)

			if status.Code(err) != codes.DeadlineExceeded {
				if err := nm.suggestIdleNodes(channelSizeSat); err != nil {
					log.Warnf("unable to give suggestion which nodes are idle: %v"+
						"", err)
				}
			}

			return err
		}
	}

	return nil
}

// suggestIdleNodes is used as suggestion function for which funds and from which
// node we have to remove to release idle funds and most unused funds.
func (nm *NodeManager) suggestIdleNodes(amount btcutil.Amount) error {
	m := crypto.NewMetric(nm.cfg.Asset, common.GetFunctionName(),
		nm.cfg.MetricsBackend)
	defer m.Finish()

	nodeStats, err := nm.GetNodeStats("month")
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return errors.Errorf("unable to calculate nodes statistics: %v", err)
	}

	for _, node := range stats.RankByIdleFunds(nodeStats) {
		// Skip removing nodes which are important.
		if _, ok := nm.importantNodes[node.NodeID]; ok {
			continue
		}

		sentFlow := node.AverageSentForwardSat + node.AverageSentSat -
			node.AverageReceivedForwardSat

		// Calculate which number of funds are locked for nothing,
		// because expected flow in the direction of node is less that locked
		// funds.
		fundsToRelease := node.LockedLocallyOverall - sentFlow
		if fundsToRelease < 0 {
			fundsToRelease = 0
		}

		log.Debugf("Suggest to remove funds(%v) from node(%v)",
			fundsToRelease, node.NodeID)

		amount -= fundsToRelease
		if amount < 0 {
			break
		}
	}

	return nil
}

// AddImportantNode is used to notify node manager, which nodes are important
// and has to be monitored for channel existence, and availability. As well as
// report is something is wrong.
func (nm *NodeManager) AddImportantNode(nodeID lightning.NodeID, nodeName string) {
	nm.importantNodesMutex.Lock()
	defer nm.importantNodesMutex.Unlock()

	log.Infof("Add important node(%v), pub key(%v)", nodeName, nodeID)
	nm.importantNodes[nodeID] = nodeName
}

// getRandomPseudonym returns random pseudonym to obscure the real
// identification of receiver/sender.
func getRandomPseudonym() string {
	rand.Seed(time.Now().UTC().UnixNano())
	return names[rand.Intn(len(names))]
}

func (nm *NodeManager) GetDomain(nodeID lightning.NodeID) string {
	nm.importantNodesMutex.Lock()
	defer nm.importantNodesMutex.Unlock()

	if nm.cfg.OurNodeID == nodeID {
		return nm.cfg.OurName
	}

	name, ok := nm.importantNodes[nodeID]
	if !ok {
		return ""
	}

	return strings.ToLower(name)
}

// GetAlias return the alias by the given public key of the receiver/server,
// if node is not in the public list, than we obscure the name.
func (nm *NodeManager) GetAlias(nodeID lightning.NodeID) string {
	nm.importantNodesMutex.Lock()
	defer nm.importantNodesMutex.Unlock()

	if nm.cfg.OurNodeID == nodeID {
		return nm.cfg.OurName
	}

	name, ok := nm.importantNodes[nodeID]
	if !ok {
		name = getRandomPseudonym()
	}

	return strings.ToLower(name)
}

// GetNodeStats returns statistics which node managers is using to make
// decision about funds management.
func (nm *NodeManager) GetNodeStats(period string) (
	map[lightning.NodeID]stats.NodeStats, error) {

	m := crypto.NewMetric(nm.cfg.Asset, common.GetFunctionName(),
		nm.cfg.MetricsBackend)
	defer m.Finish()

	payments, err := nm.cfg.Client.ListPayments("", lightning.AllStatuses,
		lightning.AllDirections, lightning.AllSystems)
	if err != nil {
		err := errors.Errorf("unable list payments: %v", err)
		m.AddError(metrics.HighSeverity)
		return nil, err
	}

	forwardPayments, err := nm.cfg.Client.ListForwardPayments()
	if err != nil {
		err := errors.Errorf("unable list forward payments: %v", err)
		m.AddError(metrics.HighSeverity)
		return nil, err
	}

	channels, err := nm.cfg.Client.Channels()
	if err != nil {
		err := errors.Errorf("unable fetch channels: %v", err)
		m.AddError(metrics.HighSeverity)
		return nil, err
	}

	return stats.GetNodeStats(period, payments,
		forwardPayments, channels)
}
