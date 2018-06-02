package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/jessevdk/go-flags"
	"path/filepath"
	"github.com/go-errors/errors"
	"github.com/gorilla/websocket"
	"net/http"
	"net"
	"github.com/kr/pretty"
	"sync"
	"github.com/golang/protobuf/jsonpb"
	"bytes"
	"github.com/golang/protobuf/proto"
	"encoding/json"
	"strings"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func createWebsocketHandler(hubLogFile, nonHubLogFile string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			mainLog.Infof("upgrade:", err)
			return
		}
		defer c.Close()

		tsc := &threadSafeConn{conn: c}

		var wg sync.WaitGroup

		wg.Add(1)
		go sendLogUpdates(tsc, hubLogFile, "hub", wg)

		wg.Add(1)
		go sendLogUpdates(tsc, nonHubLogFile, "nonhub", wg)

		wg.Wait()
	}
}

func sendLogUpdates(c *threadSafeConn, filePath, prefix string, wg sync.WaitGroup) {
	defer wg.Done()

	watcher, err := newUpdateLogWatcher(filePath)
	if err != nil {
		err := errors.Errorf("(%v) unable create file watcher: %v", prefix, err)
		mainLog.Error(err.Error())
		c.ubNormalClose(err)
		return
	}
	defer watcher.stop()

	sub := watcher.subscribe()
	wrappedSub := wrapUpdateLogSubscription(sub)

	for {
		log := <-wrappedSub
		mainLog.Infof("(%v) new log: %v", prefix, pretty.Sprintf("%v", log))

		err = c.send(prefix, log)
		if err != nil {
			mainLog.Errorf("(%v) unable to write websocket message:", prefix, err)
			break
		}
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

	if config.NonHubLogFile == "" {
		return errors.New("non-hub update log file path is not specified")
	}

	// Run websocket server and send log file notifications to every
	// subscribed client
	http.HandleFunc("/activity", createWebsocketHandler(config.HubLogFile,
		config.NonHubLogFile))
	addr := net.JoinHostPort(config.Websocket.ListenHost,
		config.Websocket.ListenPort)

	mainLog.Infof("Listening websocket subscription on %v", addr)
	http.ListenAndServe(addr, nil)

	addInterruptHandler(shutdownChannel, func() {
	})

	return nil
}

type threadSafeConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

type sendMessage struct {
	Prefix string
	Log    interface{}
}

func (c *threadSafeConn) send(prefix string, msg proto.Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var b bytes.Buffer
	m := jsonpb.Marshaler{}
	if err := m.Marshal(&b, msg); err != nil {
		return err
	}

	data, err := json.Marshal(sendMessage{Prefix: prefix, Log: b.String()})
	if err != nil {
		return err
	}

	s := string(data)
	res := strings.Replace(s, "\\", "", -1)
	return c.conn.WriteMessage(websocket.TextMessage, []byte(res))
}

func (c *threadSafeConn) ubNormalClose(err error) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, err.Error()))
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
