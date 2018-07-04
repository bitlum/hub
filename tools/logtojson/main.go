package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/jessevdk/go-flags"
	"path/filepath"
	"github.com/go-errors/errors"
	"bytes"
	"github.com/golang/protobuf/jsonpb"
	"github.com/kr/pretty"
)

func printLogs(filePath string) {
	mainLog.Infof("Start printing logs, log file(%v)", filePath)

	watcher, err := newUpdateLogWatcher(filePath)
	if err != nil {
		err := errors.Errorf("unable create file watcher: %v", err)
		mainLog.Error(err.Error())
		return
	}
	defer watcher.stop()

	sub := watcher.subscribe()

	for {
		log := <-sub

		var b bytes.Buffer
		m := jsonpb.Marshaler{
			OrigName:     true,
			EmitDefaults: true,
			Indent:       "\t",
		}
		if err := m.Marshal(&b, log); err != nil {
			mainLog.Error(err.Error())
			return
		}

		mainLog.Infof("new log: %v", pretty.Sprintf("%v", b.String()))
	}
}

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

	if config.HubLogFile == "" {
		return errors.New("hub update log file path is not specified")
	}
	mainLog.Infof("Init hub log path(%v)", config.HubLogFile)

	go printLogs(config.HubLogFile)

	addInterruptHandler(shutdownChannel, func() {

	})

	<-shutdownChannel
	mainLog.Infof("Exit script")
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
