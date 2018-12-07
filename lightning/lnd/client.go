package lnd

import (
	"github.com/bitlum/hub/common/broadcast"
	"github.com/bitlum/hub/lightning"
	"github.com/bitlum/hub/metrics"
	"github.com/bitlum/hub/metrics/crypto"
	"github.com/bitlum/hub/registry"
	"github.com/go-errors/errors"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/macaroons"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/macaroon.v2"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Config is a connector config.
type Config struct {
	// Asset name of the asset with which operates the lightning.
	Asset string

	// Port is gRPC port of lnd daemon.
	Port string

	// Host is gRPC host of lnd daemon.
	Host string

	// TlsCertPath is a path to certificate, which is needed to have a secure
	// gRPC connection with lnd daemon.
	TlsCertPath string

	// MacaroonPath is a path to macaroon token, which is needed to have
	// the RPC authorisation.
	MacaroonPath string

	// Storage is used to store all data which needs to be persistent,
	// exact implementation of database backend is unknown for the hub,
	// in the simplest case it might be in-memory storage.
	Storage ClientStorage

	// MetricsBackend is used to send metrics about internal state of the
	// lightning client, and act on errors accordingly.
	MetricsBackend crypto.MetricsBackend

	// Net is the blockchain network which hub should operate on,
	// if hub trying to connect to the lnd with different network,
	// than init should fail.
	Net string

	// TODO(andrew.shvv) Remove if lightning network info will be init in
	// another place.
	// NeutrinoHost is an information about neutrino backend host,
	// which is needed for external users to connect to our bitcoind node.
	NeutrinoHost string

	// TODO(andrew.shvv) Remove if lightning network info will be init in
	// another place.
	// NeutrinoPort is an information about neutrino backend port,
	// which is needed for external users to connect to our bitcoind node.
	NeutrinoPort string

	// TODO(andrew.shvv) Remove if lightning network info will be init in
	// another place.
	// PeerHost is a public host of hub lightning network node.
	PeerHost string

	// TODO(andrew.shvv) Remove if lightning network info will be init in
	// another place.
	// PeerHost is a public port of hub lightning network node,
	// to which users are connect to negotiate creation of channel.
	PeerPort string
}

func (c *Config) validate() error {
	if c.Asset == "" {
		return errors.Errorf("asset should be specified")
	}

	if c.Port == "" {
		return errors.Errorf("port should be specified")
	}

	if c.Host == "" {
		return errors.Errorf("host should be specified")
	}

	if c.TlsCertPath == "" {
		return errors.Errorf("tlc cert path should be specified")
	}

	if c.Storage == nil {
		return errors.Errorf("db should be specified")
	}

	if c.MetricsBackend == nil {
		return errors.Errorf("metrics backend should be specified")
	}

	if c.Net == "" {
		return errors.Errorf("net should be specified")
	}

	return nil
}

// Client is the lightning network daemon gRPC client wrapper which makes it
// compatible with our internal lightning client interface.
type Client struct {
	started  int32
	shutdown int32
	wg       sync.WaitGroup
	quit     chan struct{}

	client lnrpc.LightningClient
	conn   *grpc.ClientConn

	cfg                 *Config
	lightningNodeUserID lightning.UserID

	// broadcaster is used to broadcast lightning node updates in the
	// non-blocking manner. If one of receiver would not read te update write
	// wouldn't stuck.
	broadcaster *broadcast.Broadcaster
}

// Runtime check to ensure that Connector implements common.LightningConnector
// interface.
var _ lightning.Client = (*Client)(nil)

func NewClient(cfg *Config) (*Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, errors.Errorf("config is invalid: %v", err)
	}

	return &Client{
		cfg:         cfg,
		quit:        make(chan struct{}),
		broadcaster: broadcast.NewBroadcaster(),
	}, nil
}

// Start...
func (client *Client) Start() error {
	if !atomic.CompareAndSwapInt32(&client.started, 0, 1) {
		log.Warn("Lnd client already started")
		return nil
	}

	m := crypto.NewMetric(client.cfg.Asset, "Start",
		client.cfg.MetricsBackend)
	defer m.Finish()

	creds, err := credentials.NewClientTLSFromFile(client.cfg.TlsCertPath, "")
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return errors.Errorf("unable to load credentials: %v", err)
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}

	if client.cfg.MacaroonPath != "" {
		macaroonBytes, err := ioutil.ReadFile(client.cfg.MacaroonPath)
		if err != nil {
			return errors.Errorf("unable to read macaroon file: %v", err)
		}

		mac := &macaroon.Macaroon{}
		if err = mac.UnmarshalBinary(macaroonBytes); err != nil {
			return errors.Errorf("unable to unmarshal macaroon: %v", err)
		}

		opts = append(opts,
			grpc.WithPerRPCCredentials(macaroons.NewMacaroonCredential(mac)))
	}

	target := net.JoinHostPort(client.cfg.Host, client.cfg.Port)
	log.Infof("Lightning client connects to lnd: %v", target)

	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return errors.Errorf("unable to to dial grpc: %v", err)
	}
	client.conn = conn
	client.client = lnrpc.NewLightningClient(client.conn)

	reqInfo := &lnrpc.GetInfoRequest{}
	respInfo, err := client.client.GetInfo(getContext(), reqInfo)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return errors.Errorf("unable get lnd node info: %v", err)
	}

	lndNet := "simnet"
	if respInfo.Testnet {
		lndNet = "testnet"
	}

	// TODO(andrew.shvv) not working for mainnet, as far response don't have
	// a mainnet param.
	if client.cfg.Net != "mainnet" {
		if lndNet != client.cfg.Net {
			return errors.Errorf("hub net is '%v', but config net is '%v'",
				client.cfg.Net, lndNet)
		}

		log.Infof("Init connector working with '%v' net", lndNet)
	} else {
		log.Info("Init connector working with 'mainnet' net")
	}

	log.Infof("Init lnd client working with network(%v) alias(%v) ", lndNet,
		respInfo.Alias)

	client.lightningNodeUserID = lightning.UserID(respInfo.IdentityPubkey)
	log.Infof("Init lnd client with pub key: %v", respInfo.IdentityPubkey)

	// Register ZigZag us known and public lighting network node.
	registry.AddKnownPeer(lightning.UserID(respInfo.IdentityPubkey), "ZigZag")

	client.wg.Add(1)
	go client.updateNodeInfo()

	client.wg.Add(1)
	go client.updateChannelStates()

	client.wg.Add(1)
	go client.listenForwardingPayments()

	client.wg.Add(1)
	go client.listenIncomingPayments()

	client.wg.Add(1)
	go client.listenOutgoingPayments()

	client.wg.Add(1)
	go client.updatePeers()

	log.Info("Lnd client started")
	return nil
}

// Stop gracefully stops the connection with lnd daemon.
func (client *Client) Stop(reason string) error {
	if !atomic.CompareAndSwapInt32(&client.shutdown, 0, 1) {
		log.Warn("lnd client already shutdown")
		return nil
	}

	close(client.quit)
	if err := client.conn.Close(); err != nil {
		return errors.Errorf("unable to close connection to lnd: %v", err)
	}

	client.wg.Wait()

	log.Infof("lnd client shutdown, reason(%v)", reason)
	return nil
}

// SendPayment makes the payment on behalf of lightning. In the context of
// lightning network hub manager this hook might be used for future
// off-chain channel re-balancing tactics.
//
// NOTE: Part of the lightning.Client interface.
func (client *Client) SendPayment(userID lightning.UserID, amount lightning.BalanceUnit) error {
	// TODO(andrew.shvv) Add implementation when rebalancing strategy will be
	// needed.
	return nil
}

// OpenChannel opens the channel with the given user.
//
// NOTE: Part of the lightning.Client interface.
func (client *Client) OpenChannel(id lightning.UserID, funds lightning.BalanceUnit) error {
	m := crypto.NewMetric(client.cfg.Asset, "OpenChannel",
		client.cfg.MetricsBackend)
	defer m.Finish()

	req := &lnrpc.OpenChannelRequest{
		NodePubkeyString:   string(id),
		LocalFundingAmount: int64(funds),

		// TODO(andrew.shvv) We have to make a subsystem for optimising this
		// ration.
		SatPerByte: 0,

		// TODO(andrew.shvv) Should we make is equal to minimum exchange amount?
		MinHtlcMsat: 0,

		// TODO(andrew.shvv) Does the Jack wallet supports private channels?
		Private: false,
	}

	_, err := client.client.OpenChannelSync(getContext(), req)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return err
	}

	return nil
}

// CloseChannel closes the specified channel.
//
// NOTE: Part of the lightning.Client interface.
func (client *Client) CloseChannel(id lightning.ChannelID) error {
	m := crypto.NewMetric(client.cfg.Asset, "CloseChannel",
		client.cfg.MetricsBackend)
	defer m.Finish()

	parts := strings.Split(string(id), ":")
	if len(parts) != 2 {
		m.AddError(metrics.HighSeverity)
		log.Error("unable to split chan point("+
			"%v) on funding tx id and output index", id)
		return errors.New("unable decode channel point")
	}

	index, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		log.Error("unable to parse index", parts[1])
		return errors.Errorf("unable decode tx index: %v", err)
	}

	fundingTx := &lnrpc.ChannelPoint_FundingTxidStr{
		FundingTxidStr: parts[0],
	}

	req := &lnrpc.CloseChannelRequest{
		ChannelPoint: &lnrpc.ChannelPoint{
			FundingTxid: fundingTx,
			OutputIndex: uint32(index),
		},
		Force: false,
	}

	if _, err = client.client.CloseChannel(getContext(), req); err != nil {
		m.AddError(metrics.HighSeverity)
		log.Errorf("unable close the channel: %v", err)
		return err
	}

	return nil
}

// UpdateChannel updates the number of locked funds in the specified
// channel.
//
// NOTE: Part of the lightning.Client interface.
func (client *Client) UpdateChannel(id lightning.ChannelID,
	funds lightning.BalanceUnit) error {

	// TODO(andrew.shvv) Implement
	// This hook would require splice-out splice-in mechanism to be
	// implemented in lnd.

	return nil
}

// RegisterOnUpdates returns updates about lightning node local network topology
// changes, about attempts of propagating the payment through the
// lightning node, about fee changes etc.
//
// NOTE: Part of the lightning.Client interface.
func (client *Client) RegisterOnUpdates() *broadcast.Receiver {
	return client.broadcaster.Subscribe()
}

// Network returns the information about the current local topology of
// our lightning node.
//
// NOTE: Part of the lightning.Client interface.
func (client *Client) Channels() ([]*lightning.Channel, error) {
	m := crypto.NewMetric(client.cfg.Asset, "Channels", client.cfg.MetricsBackend)
	defer m.Finish()

	channels, err := client.cfg.Storage.Channels()
	if err != nil {
		m.AddError(metrics.HighSeverity)
		log.Errorf("unable to fetch channels from sync storage: %v", err)
		return nil, err
	}

	// Initialise channels with lightning client broadcaster so that channel
	// notification will be sent to every lightning client subscribers.
	for _, channel := range channels {
		channel.SetConfig(&lightning.ChannelConfig{
			Broadcaster: client.broadcaster,
			Storage:     client.cfg.Storage,
		})
	}

	return channels, nil
}

func (client *Client) Users() ([]*lightning.User, error) {
	m := crypto.NewMetric(client.cfg.Asset, "Users", client.cfg.MetricsBackend)
	defer m.Finish()

	users, err := client.cfg.Storage.Users()
	if err != nil {
		m.AddError(metrics.HighSeverity)
		log.Errorf("unable to fetch users from sync storage: %v", err)
		return nil, err
	}

	for _, user := range users {
		user.SetConfig(&lightning.UserConfig{
			Storage: client.cfg.Storage,
		})
	}

	return users, nil
}

// FreeBalance returns the amount of funds at lightning node disposal.
//
// NOTE: Part of the lightning.Client interface.
func (client *Client) FreeBalance() (lightning.BalanceUnit, error) {
	m := crypto.NewMetric(client.cfg.Asset, "FreeBalance", client.cfg.MetricsBackend)
	defer m.Finish()

	req := &lnrpc.WalletBalanceRequest{}
	resp, err := client.client.WalletBalance(getContext(), req)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		log.Errorf("unable to get wallet balance channels: %v", err)
		return 0, err
	}

	return lightning.BalanceUnit(resp.ConfirmedBalance), nil
}

// PendingBalance returns the amount of funds which in the process of
// being accepted by blockchain.
//
// NOTE: Part of the lightning.Client interface.
func (client *Client) PendingBalance() (lightning.BalanceUnit, error) {
	m := crypto.NewMetric(client.cfg.Asset, "PendingBalance", client.cfg.MetricsBackend)
	defer m.Finish()

	req := &lnrpc.PendingChannelsRequest{}
	resp, err := client.client.PendingChannels(getContext(), req)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		log.Errorf("unable to fetch pending channels: %v", err)
		return 0, err
	}

	return lightning.BalanceUnit(resp.TotalLimboBalance), nil
}

// AverageChangeUpdateDuration average time which is needed the change of
// state to ba updated over blockchain.
//
// NOTE: Part of the lightning.Client interface.
func (client *Client) AverageChangeUpdateDuration() (time.Duration, error) {
	// TODO(andrew.shvv) Implement
	// This hook would require us to make some channel update registry,
	// probably this would require to write data in some persistent storage.

	return 0, nil
}

// Done returns error if lightning client stopped working for some reason,
// and nil if it was stopped.
//
// NOTE: Part of the lightning.Client interface.
func (client *Client) Done() chan error {
	// TODO(andrew.shvv) Implement
	return nil
}

// Asset returns asset with which corresponds to this lightning.
//
// NOTE: Part of the lightning.Client interface.
func (client *Client) Asset() string {
	return client.cfg.Asset
}

// SetFeeBase sets base number of milli units (i.e milli satoshis in
// Bitcoin) which will be taken for every forwarding payment.
func (client *Client) SetFeeBase(feeBase int64) error {
	// TODO(andrew.shvv) Implement
	return nil
}

// SetFeeProportional sets the number of milli units (i.e milli
// satoshis in Bitcoin) which will be taken for every killo-unit of
// forwarding payment amount as a forwarding fee.
func (client *Client) SetFeeProportional(feeProportional int64) error {
	// TODO(andrew.shvv) Implement
	return nil
}
