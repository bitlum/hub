package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/jessevdk/go-flags"
	"path/filepath"
	"github.com/bitlum/hub/manager/router/emulation"
	"github.com/go-errors/errors"
	"github.com/bitlum/hub/manager/router"
	"github.com/bitlum/hub/manager/hubrpc"
	"google.golang.org/grpc"
	"net"
	"time"
	"github.com/bitlum/hub/manager/router/lnd"
	"github.com/bitlum/hub/manager/metrics"
	"context"
	"github.com/bitlum/hub/manager/metrics/crypto"
	"github.com/bitlum/hub/manager/metrics/network"
	"github.com/bitlum/hub/manager/router/stats"
	"github.com/bitlum/hub/manager/db"
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

	// Get log file path from config, which will be used for pushing router
	// topology updates in it.
	if config.UpdateLogFile == "" {
		return errors.Errorf("update log file should be specified")
	}

	// Initialise prometheus endpoint handler listening for incoming prometheus
	// scraping.
	mainLog.Infof("Initialise prometheus endpoint on %v:%v",
		config.Prometheus.ListenHost, config.Prometheus.ListenPort)
	addr := net.JoinHostPort(config.Prometheus.ListenHost, config.Prometheus.ListenPort)
	server := metrics.StartServer(addr)

	// TODO(andrew.shvv) add simnet to config and check in lnd that we
	// connect to client with proper net
	metricsBackend, err := crypto.InitMetricsBackend("simnet")
	if err != nil {
		return errors.Errorf("unable to init metrics backend for lnd: %v"+
			"", err)
	}

	// Create router and connect to emulation or real network,
	// and subscribe on topology updates which will transformed and written
	// in the file, so that third-party optimisation program could read it
	// and make optimisation decisions.
	errChan := make(chan error)

	var r router.Router
	switch config.Backend {
	case "emulator":
		mainLog.Infof("Initialise emulator router...")
		emulationRouter := emulation.NewRouter(10, 200*time.Millisecond)
		emulationRouter.Start(config.Emulator.ListenHost, config.Emulator.ListenPort)
		defer emulationRouter.Stop()
		r = emulationRouter
	case "lnd":
		// Create or open database file to host the last state of
		// synchronization.
		mainLog.Info("Opening BoltDB database, path: '%v'", config.LND.DataDir)

		database, err := db.Open(config.LND.DataDir, "lnd")
		if err != nil {
			return errors.Errorf("unable to open database: %v", err)
		}

		mainLog.Infof("Initialise lnd router...")
		lndConfig := &lnd.Config{
			Asset:          "BTC",
			Host:           config.LND.Host,
			Port:           config.LND.Port,
			TlsCertPath:    config.LND.TlsCert,
			DB:             database,
			MetricsBackend: metricsBackend,
			Net:            config.LND.Net,
		}

		lndRouter, err := lnd.NewRouter(lndConfig)
		if err != nil {
			return errors.Errorf("unable to init lnd router: %v", err)
		}

		// TODO(andrew.shvv) add simnet to config and check in lnd that we
		// connect to client with proper net
		statsBackend, err := network.InitMetricsBackend("simnet")
		if err != nil {
			return errors.Errorf("unable to init metrics backend for lnd: %v"+
				"", err)
		}

		statsGatherer := stats.NewNetworkStatsGatherer(lndRouter, statsBackend)
		statsGatherer.Start()
		defer statsGatherer.Stop()

		// Start router after the stats gatherer so that if lnd has any new
		// payment updates stats gatherer wouldn't lost them,
		// because of the late notification subscription.
		if err := lndRouter.Start(); err != nil {
			return errors.Errorf("unable to start lnd router: %v", err)
		}
		defer lndRouter.Stop("shutdown")
		r = lndRouter

		// Start maintaining the balance of funds locked with users.
		enableChannelBalancing(lndRouter)

	default:
		return errors.Errorf("unhandled backend name: '%v'", config.Backend)
	}

	go updateLogFileGoroutine(r, config.UpdateLogFile, errChan)

	// Setup gRPC endpoint to receive the management commands, and initialise
	// optimisation strategy which will dictate us how to convert from one
	// router state to another.
	grpcServer := grpc.NewServer([]grpc.ServerOption{}...)

	s := router.NewChannelUpdateStrategy()
	hub := hubrpc.NewHub(r, s)
	hubrpc.RegisterManagerServer(grpcServer, hub)

	go func() {
		addr := net.JoinHostPort(config.Hub.Host, config.Hub.Port)
		mainLog.Infof("Start listening on: %v", addr)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			fail(errChan, "gRPC server unable to listen on %s", addr)
			return
		}
		defer lis.Close()

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
		server.Shutdown(context.Background())
	})

	select {
	case err := <-errChan:
		if err != nil {
			mainLog.Error("exit program because of: %v", err)
			return err
		}
	case err := <-r.Done():
		if err != nil {
			mainLog.Error("emulator router stopped working: %v", err)
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
