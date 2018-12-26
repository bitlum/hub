package lnd

import (
	"github.com/bitlum/hub/common"
	"github.com/bitlum/hub/common/broadcast"
	"github.com/bitlum/hub/lightning"
	"github.com/bitlum/hub/lightning/lnd/explorer"
	"github.com/bitlum/hub/metrics"
	"github.com/bitlum/hub/metrics/crypto"
	"github.com/btcsuite/btcutil"
	"github.com/go-errors/errors"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/macaroons"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/macaroon.v2"
	"io/ioutil"
	"net"
	"sync"
	"sync/atomic"
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
	Storage InfoStorage

	// MetricsBackend is used to send metrics about internal state of the
	// lightning client, and act on errors accordingly.
	MetricsBackend crypto.MetricsBackend

	// Net is the blockchain network which hub should operate on,
	// if hub trying to connect to the lnd with different network,
	// than init should fail.
	Net string

	// Explorer is the service which allows to fetch data about bitcoin
	// blockchain.
	Explorer explorer.Explorer

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

	if c.Explorer == nil {
		return errors.Errorf("explorer should be specified")
	}

	return nil
}

// Client is the lightning network daemon gRPC client wrapper which makes it
// compatible with our internal lightning client interface.
type Client struct {
	started        int32
	shutdown       int32
	wg             sync.WaitGroup
	quit           chan struct{}
	startedTrigger chan struct{}

	rpc      lnrpc.LightningClient
	conn     *grpc.ClientConn

	cfg                 *Config
	lightningNodeUserID lightning.NodeID

	// averageFee...
	averageFee decimal.Decimal

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
		cfg:            cfg,
		quit:           make(chan struct{}),
		startedTrigger: make(chan struct{}),
		broadcaster:    broadcast.NewBroadcaster(),
	}, nil
}

// Start init rpc lightning client, ensure in validity of network, and launches
// goroutines for syncing information of lnd state.
func (c *Client) Start() error {
	if !atomic.CompareAndSwapInt32(&c.started, 0, 1) {
		log.Warn("lnd client already started")
		return nil
	}

	m := crypto.NewMetric(c.cfg.Asset, common.GetFunctionName(),
		c.cfg.MetricsBackend)
	defer m.Finish()

	creds, err := credentials.NewClientTLSFromFile(c.cfg.TlsCertPath, "")
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return errors.Errorf("unable to load credentials: %v", err)
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}

	if c.cfg.MacaroonPath != "" {
		macaroonBytes, err := ioutil.ReadFile(c.cfg.MacaroonPath)
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

	target := net.JoinHostPort(c.cfg.Host, c.cfg.Port)
	log.Infof("Connect to lnd grpc endpoint: %v", target)

	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return errors.Errorf("unable to to dial grpc: %v", err)
	}
	c.conn = conn
	c.rpc = lnrpc.NewLightningClient(c.conn)

	reqInfo := &lnrpc.GetInfoRequest{}
	respInfo, err := c.rpc.GetInfo(timeout(10), reqInfo)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return errors.Errorf("unable get lnd node info: %v", err)
	}

	lndNet := "simnet"
	if respInfo.Testnet {
		lndNet = "testnet"
	}

	if c.cfg.Net != "mainnet" {
		if lndNet != c.cfg.Net {
			return errors.Errorf("hub net is '%v', but config net is '%v'",
				c.cfg.Net, lndNet)
		}

		log.Infof("Init connector working with '%v' net", lndNet)
	} else {
		log.Info("Init connector working with 'mainnet' net")
	}

	log.Infof("Init lnd client working with network(%v) alias(%v) ", lndNet,
		respInfo.Alias)

	c.lightningNodeUserID = lightning.NodeID(respInfo.IdentityPubkey)
	log.Infof("Init lnd client with pub key: %v", respInfo.IdentityPubkey)

	if err := c.syncChannels(); err != nil {
		return err
	}

	c.wg.Add(1)
	go c.updateChannelStates()

	log.Info("lnd client started")
	close(c.startedTrigger)
	return nil
}

// Stop gracefully stops the connection with lnd daemon.
func (c *Client) Stop(reason string) error {
	if !atomic.CompareAndSwapInt32(&c.shutdown, 0, 1) {
		log.Warn("lnd client already shutdown")
		return nil
	}

	close(c.quit)
	if err := c.conn.Close(); err != nil {
		return errors.Errorf("unable to close connection to lnd: %v", err)
	}

	c.wg.Wait()

	log.Infof("lnd client shutdown, reason(%v)", reason)
	return nil
}

// FreeBalance returns the amount of funds at lightning node disposal.
//
// NOTE: Part of the lightning.Client interface.
func (c *Client) AvailableBalance() (btcutil.Amount, error) {
	select {
	case <-c.startedTrigger:
	}

	m := crypto.NewMetric(c.cfg.Asset, common.GetFunctionName(), c.cfg.MetricsBackend)
	defer m.Finish()

	req := &lnrpc.ListChannelsRequest{}
	resp, err := c.rpc.ListChannels(timeout(20), req)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		log.Errorf("unable to list channels: %v", err)
		return 0, err
	}

	var balanceSat int64
	for _, channel := range resp.Channels {
		balanceSat += channel.LocalBalance
	}

	return btcutil.Amount(balanceSat), nil
}

// PendingBalance returns the amount of funds which in the process of
// being accepted by blockchain.
//
// NOTE: Part of the lightning.Client interface.
func (c *Client) PendingBalance() (btcutil.Amount, error) {
	select {
	case <-c.startedTrigger:
	}

	m := crypto.NewMetric(c.cfg.Asset, common.GetFunctionName(), c.cfg.MetricsBackend)
	defer m.Finish()

	req := &lnrpc.PendingChannelsRequest{}
	resp, err := c.rpc.PendingChannels(timeout(10), req)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		log.Errorf("unable to fetch pending channels: %v", err)
		return 0, err
	}

	var balanceSat int64
	for _, info := range resp.PendingClosingChannels {
		balanceSat += info.Channel.LocalBalance
	}

	for _, info := range resp.PendingOpenChannels {
		balanceSat += info.Channel.LocalBalance
	}

	for _, info := range resp.PendingForceClosingChannels {
		balanceSat += info.Channel.LocalBalance
	}

	for _, info := range resp.WaitingCloseChannels {
		balanceSat += info.Channel.LocalBalance
	}

	return btcutil.Amount(balanceSat), nil
}

// Asset returns asset with which corresponds to this lightning.
//
// NOTE: Part of the lightning.Client interface.
func (c *Client) Asset() string {
	return c.cfg.Asset
}

// Info returns the information about our lnd node.
//
// NOTE: Part of the lightning.Client interface.
func (c *Client) Info() (*lightning.Info, error) {
	select {
	case <-c.startedTrigger:
	}

	m := crypto.NewMetric(c.cfg.Asset, common.GetFunctionName(),
		c.cfg.MetricsBackend)
	defer m.Finish()

	req := &lnrpc.GetInfoRequest{}
	info, err := c.rpc.GetInfo(timeout(10), req)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return nil, err
	}

	return &lightning.Info{
		MinPaymentAmount: btcutil.Amount(1),
		MaxPaymentAmount: btcutil.Amount(4200000),
		Version:          info.Version,
		NodeInfo: &lightning.NodeInfo{
			Alias:          "",
			Host:           c.cfg.Host,
			Port:           c.cfg.Port,
			IdentityPubKey: info.IdentityPubkey,
		},
		BlockHeight: info.BlockHeight,
		BlockHash:   info.BlockHash,
		Network:     c.cfg.Net,
		NeutrinoInfo: &lightning.NeutrinoInfo{
			Host: c.cfg.NeutrinoHost,
			Port: c.cfg.PeerHost,
		},
	}, nil
}
