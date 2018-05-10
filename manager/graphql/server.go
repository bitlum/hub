package graphql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
	"github.com/graphql-go/graphql"
	"sync/atomic"
)

// Server is a zigzag graphQL server.
type Server struct {
	shutdown int32
	started  int32
	quit     chan struct{}

	// stopWG used for waiting until all servers's goroutines are
	// stopped before return from `Stop`.
	stopWG sync.WaitGroup

	cfg Config

	schema graphql.Schema

	httpServer *http.Server
}

func NewServer(c Config) (*Server, error) {
	log.Tracef("NewServer() ListenIP=%s, ListenPort=%s",
		c.ListenIP, c.ListenPort)
	if err := c.validate(); err != nil {
		return nil, fmt.Errorf("config validation error: %v",
			err.Error())
	}

	schema, err := New()
	if err != nil {
		return nil, fmt.Errorf("unable to create schema: %v",
			err)
	}

	const (
		queryPath = "/query"
	)

	addr := c.PublicHost
	if c.ListenPort != "80" {
		addr = net.JoinHostPort(addr, c.ListenPort)
	}

	faqRenderParams := faqRenderParams{
		QueryPath: queryPath,
		Addr:      addr,
	}

	switch c.Network {
	case "testnet":
	}

	faq, err := renderFAQ(faqRenderParams)
	if err != nil {
		return nil, errors.New("unable to generate graphiQL FAQ: " +
			err.Error())
	}

	graphiQLPage, err := renderGraphiQL(graphiQLParams{
		QueryPath:        queryPath,
		PublicHost:       c.PublicHost,
		ListenPort:       c.ListenPort,
		SecureListenPort: c.SecureListenPort,
		FAQ:              faq,
	})
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter,
		r *http.Request) {
		w.Write(graphiQLPage)
	}))
	mux.Handle(queryPath, &queryHandler{schema: schema})

	httpServer := &http.Server{Addr: c.listenAddr(), Handler: mux}

	return &Server{
		cfg:        c,
		schema:     schema,
		quit:       make(chan struct{}),
		httpServer: httpServer,
	}, nil
}

// Start starts http server and subscriptions processing.
func (s *Server) Start() error {
	log.Trace("Server.Start()")
	if !atomic.CompareAndSwapInt32(&s.started, 0, 1) {
		log.Warn("http server already started")
		return errors.New("http server already started")
	}

	s.stopWG.Add(1)
	go s.listenAndServeHTTP()

	// wait a little for goroutines to be completely run
	time.Sleep(100 * time.Millisecond)

	return nil
}

// Stop initiates server stop and waits infinitely until it completely
// stops.
func (s *Server) Stop() (err error) {
	log.Trace("Server.Stop()")
	if !atomic.CompareAndSwapInt32(&s.shutdown, 0, 1) {
		log.Warn("http server already shutdown")
		return errors.New("http server already shutdown")
	}

	close(s.quit)
	err = s.httpServer.Shutdown(context.Background())
	s.stopWG.Wait()
	return err
}

// listenAndServeHTTP starts HTTP server. Intended to be called as
// goroutine.
func (s *Server) listenAndServeHTTP() {
	log.Trace("Server.listenAndServeHTTP()")
	for {
		log.Infof("Starting HTTP server on `%s`", s.cfg.listenAddr())
		err := s.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Errorf("Unable to HTTP listen and serve: %v", err)
		}
		select {
		case <-s.quit:
			log.Infof("HTTP Server stopped")
			s.stopWG.Done()
			return
		default:
			time.Sleep(time.Second)
		}
	}
}

type queryHandler struct {
	schema graphql.Schema
}

func (h *queryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Tracef("queryHandler.ServeHTTP()")
	var (
		err error

		params struct {
			Query         string                 `json:"query"`
			OperationName string                 `json:"operationName"`
			Variables     map[string]interface{} `json:"variables"`
		}
	)

	if err = json.NewDecoder(r.Body).Decode(&params); err != nil {
		log.Errorf("Unable to decode request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res := graphql.Do(graphql.Params{
		Schema:         h.schema,
		RequestString:  params.Query,
		VariableValues: params.Variables,
		OperationName:  params.OperationName,
	})

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Errorf("Unable to encode response")
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
