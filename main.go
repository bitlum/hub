package main

import (
	"fmt"
	"github.com/bitlum/hub/common"
	"github.com/bitlum/hub/db/inmemory"
	"github.com/bitlum/hub/lightning/lnd/explorer/bitcoind"
	"github.com/bitlum/hub/manager"
	"github.com/bitlum/hub/metrics/rpc"
	"math"
	"os"
	"runtime"

	"context"
	"github.com/bitlum/hub/graphql"
	"github.com/bitlum/hub/hubrpc"
	"github.com/bitlum/hub/lightning"
	"github.com/bitlum/hub/lightning/lnd"
	"github.com/bitlum/hub/metrics"
	"github.com/bitlum/hub/metrics/crypto"
	"github.com/go-errors/errors"
	"github.com/jessevdk/go-flags"
	"google.golang.org/grpc"
	"net"
	"path/filepath"
)

var (
	// shutdownChannel is used to identify that process creator send us signal to
	// shutdown the backend service.
	shutdownChannel = make(chan struct{})
)

func backendMain() error {
	// Load the configuration, parse any command line options,
	// setup log rotation.
	config := getDefaultConfig()
	if err := config.loadConfig(); err != nil {
		return errors.Errorf("unable to load config: %v", err)
	}

	logFile := filepath.Join(config.LogDir, defaultLogFilename)

	closeRotator := initLogRotator(logFile)
	defer closeRotator()

	// Get log file path from config, which will be used for pushing lightning
	// node topology updates in it.
	if config.UpdateLogFile == "" {
		return errors.Errorf("update log file should be specified")
	}

	// Initialise prometheus endpoint handler listening for incoming prometheus
	// scraping.
	mainLog.Infof("Initialise prometheus endpoint on %v:%v",
		config.Prometheus.ListenHost, config.Prometheus.ListenPort)
	addr := net.JoinHostPort(config.Prometheus.ListenHost, config.Prometheus.ListenPort)
	server := metrics.StartServer(addr)

	// Create lightning client and connect to emulation or real network,
	// and subscribe on topology updates which will transformed and written
	// in the file, so that third-party optimisation program could read it
	// and make optimisation decisions.
	errChan := make(chan error)

	//// Create or open database file to host the last state of
	//// synchronization.
	//dbName := "lnd.sqlite"
	//mainLog.Infof("Opening sqlite database, path: '%v'",
	//	filepath.Join(config.LND.DataDir, dbName))
	//
	//database, err := sqlite.Open(config.LND.DataDir, dbName)
	//if err != nil {
	//	return errors.Errorf("unable to open database: %v", err)
	//}

	metricsBackend, err := crypto.InitMetricsBackend(config.LND.Network)
	if err != nil {
		return errors.Errorf("unable to init metrics backend for lnd: %v"+
			"", err)
	}

	rpcMetricsBackend, err := rpc.InitMetricsBackend(config.LND.Network)
	if err != nil {
		return errors.Errorf("unable to init metrics backend for lnd: %v"+
			"", err)
	}

	explorer, err := bitcoind.NewExplorer(&bitcoind.Config{
		RPCHost:  config.Bitcoind.Host,
		RPCPort:  config.Bitcoind.Port,
		User:     config.Bitcoind.User,
		Password: config.Bitcoind.Pass,
	})
	if err != nil {
		return errors.Errorf("unable to init bitcoin explorer: %v", err)
	}

	mainLog.Infof("Initialise lnd lightning client...")
	lndConfig := &lnd.Config{
		Asset:          "BTC",
		Host:           config.LND.GRPCHost,
		Port:           config.LND.GRPCPort,
		TlsCertPath:    config.LND.TlsCertPath,
		MacaroonPath:   config.LND.MacaroonPath,
		Storage:        inmemory.NewInfoStorage(),
		MetricsBackend: metricsBackend,
		Net:            config.LND.Network,
		NeutrinoHost:   config.LND.NeutrinoHost,
		NeutrinoPort:   config.LND.NeutrinoPort,
		PeerHost:       config.LND.PeerHost,
		PeerPort:       config.LND.PeerPort,
		Explorer:       explorer,
	}

	lndClient, err := lnd.NewClient(lndConfig)
	if err != nil {
		return errors.Errorf("unable to init lnd lightning client: %v", err)
	}

	// Start lightning client after the stats gatherer so that if lnd has
	// any new payment updates stats gatherer wouldn't lost them, because
	// of the late notification subscription.
	if err := lndClient.Start(); err != nil {
		return errors.Errorf("unable to start lnd lightning client: %v",
			err)
	}
	defer func() {
		if err := lndClient.Stop("shutdown"); err != nil {
			mainLog.Errorf("unable to stop lnd client: %v", err)
		}
	}()
	client := lndClient

	// Register our lightning network node us known, and important.
	info, err := client.Info()
	if err != nil {
		return errors.Errorf("unable get lightning node info: %v", err)
	}

	// Initialise and start node manager, which would ensure that we always
	// have channels and connection to the important nodes.

	managerConfig := &manager.Config{
		Client:             client,
		MetricsBackend:     metricsBackend,
		GetBitcoinPriceUSD: common.GetBitcoinUSDPRice,
		Asset:              "BTC",
		OurName:            "bitlum.io",
		OurNodeID:          lightning.NodeID(info.NodeInfo.IdentityPubKey),
	}

	switch config.LND.Network {
	case "testnet", "simnet":
		managerConfig.MaxChannelSizeUSD = math.MaxInt32
		managerConfig.MinChannelSizeUSD = math.MaxInt32
		managerConfig.MaxCloseSpendingPerDayUSD = math.MaxInt32
		managerConfig.MaxOpenSpendingPerDayUSD = math.MaxInt32
		managerConfig.MaxCommitFeeUSD = math.MaxInt32
		managerConfig.MaxLimboUSD = math.MaxInt32
		managerConfig.MaxStuckBalanceUSD = math.MaxInt32
	case "mainnet":
		managerConfig.MaxChannelSizeUSD = 400
		managerConfig.MinChannelSizeUSD = 50
		managerConfig.MaxCloseSpendingPerDayUSD = 1
		managerConfig.MaxOpenSpendingPerDayUSD = 1
		managerConfig.MaxCommitFeeUSD = 10
		managerConfig.MaxLimboUSD = 300
		managerConfig.MaxStuckBalanceUSD = 300
	}

	nodeManager, err := manager.NewNodeManager(managerConfig)
	if err != nil {
		return errors.Errorf("unable create node manager: %v", err)
	}

	nodeManager.Start()
	defer nodeManager.Stop("stop")

	mainLog.Infof("Start GraphQL server serving on: %v",
		net.JoinHostPort(config.GraphQL.ListenHost, config.GraphQL.ListenPort))
	graphQLServer, err := graphql.NewServer(graphql.Config{
		ListenIP:         config.GraphQL.ListenHost,
		ListenPort:       config.GraphQL.ListenPort,
		SecureListenPort: config.GraphQL.SecureListenPort,
		Client:           client,
		GetAlias:         nodeManager.GetAlias,
	})
	if err != nil {
		return errors.New("unable to create GraphQL server: " +
			err.Error())
	}

	if err := graphQLServer.Start(); err != nil {
		return errors.New("unable to start GraphQL server: " +
			err.Error())
	}

	defer func() {
		if err := graphQLServer.Stop(); err != nil {
			mainLog.Errorf("unable to stop GraphQL server: %v", err)
		}
	}()

	for nodeName, nodePubKey := range config.LND.KnownPeers {
		nodeID := lightning.NodeID(nodePubKey)
		nodeManager.AddImportantNode(nodeID, nodeName)
	}

	//paymentRouter, err := router.NewRouter(router.Config{Client: client})
	//if err != nil {
	//	return errors.Errorf("unable to create payment router: %v", err)
	//}

	// Setup gRPC endpoint to receive the management commands, and initialise
	// optimisation strategy which will dictate us how to convert from one
	// lightning network node state to another.
	grpcServer := grpc.NewServer([]grpc.ServerOption{}...)

	hub := hubrpc.NewHub(&hubrpc.Config{
		Client:         client,
		MetricsBackend: rpcMetricsBackend,
	})
	hubrpc.RegisterHubServer(grpcServer, hub)

	go func() {
		addr := net.JoinHostPort(config.Hub.Host, config.Hub.Port)
		mainLog.Infof("Start gRPC listening on: %v", addr)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			fail(errChan, "gRPC server unable to listen on %s", addr)
			return
		}
		defer func() {
			if err := lis.Close(); err != nil {
				mainLog.Errorf("unable to stop gRPC listener: %v", err)
			}
		}()

		mainLog.Infof("Start gRPC server serving on: %v", addr)
		if err := grpcServer.Serve(lis); err != nil {
			fail(errChan, "gRPC server unable to serve on %s", addr)
			return
		}

		mainLog.Infof("Stopped gRPC server serving on: %v", addr)
	}()

	addInterruptHandler(shutdownChannel, func() {
		close(errChan)
		grpcServer.Stop()
		if err := server.Shutdown(context.Background()); err != nil {
			mainLog.Errorf("unable to shutdown metric server: %v", addr)
		}
	})

	select {
	case err := <-errChan:
		if err != nil {
			mainLog.Error("exit program because of: %v", err)
			return err
		}
	case <-shutdownChannel:
		break
	}

	return nil
}

func main() {
	// Use all processor cores.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Call the "real" main in a nested manner so the defers will properly
	// be executed in the case of a graceful shutdown.
	if err := backendMain(); err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

func fail(errChan chan error, format string, params ...interface{}) {
	err := errors.Errorf(format, params...)
	select {
	case _, ok := <-errChan:
		if !ok {
			return
		}
	default:
	}

	select {
	case errChan <- err:
	default:
	}
}
