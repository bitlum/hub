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

	// Create router and connect to emulation or real network,
	// and subscribe on topology updates which will transformed and written
	// in the file, so that third-party optimisation program could read it
	// and make optimisation decisions.
	errChan := make(chan error)
	r := emulation.NewRouter(10, 200 * time.Millisecond)
	r.Start(config.Emulate.Host, config.Emulate.Port)
	defer r.Stop()

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
		grpcServer.Stop()
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
