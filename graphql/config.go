package graphql

import (
	"errors"
	"github.com/bitlum/hub/lightning"
	"net"
	"strconv"
)

type Config struct {
	ListenIP         string
	ListenPort       string
	SecureListenPort string

	Client   lightning.Client
	GetAlias func(nodeID lightning.NodeID) string
}

func (c Config) validate() error {
	if c.ListenIP != "" && net.ParseIP(c.ListenIP) == nil {
		return errors.New("invalid listen ip address param")
	}

	if c.ListenPort != "" {
		p, err := strconv.Atoi(c.ListenPort)
		if err != nil {
			return errors.New("invalid listen port param: " +
				err.Error())
		}
		if p < 1 || p > 65535 {
			return errors.New("invalid listen port param: out" +
				" valid range")
		}
	}

	if c.SecureListenPort == "" {
		return errors.New("secure listen port should not be empty")
	}
	p, err := strconv.Atoi(c.SecureListenPort)
	if err != nil {
		return errors.New("invalid secure listen port param: " +
			err.Error())
	}
	if p < 1 || p > 65535 {
		return errors.New("invalid secure listen port param:" +
			" out valid range")
	}
	if c.SecureListenPort == c.ListenPort {
		return errors.New("secure listen port should not be" +
			" equal to listen port")
	}

	if c.Client == nil {
		return errors.New("client should be specified")
	}

	if c.GetAlias == nil {
		return errors.New("get alias func should be specified")
	}

	return nil
}

// listenAddr forms listen addr using defined `ListenIP` and
// `ListenPort`
func (c Config) listenAddr() string {
	if c.ListenIP == "" && c.ListenPort == "" {
		return ""
	}
	if c.ListenPort == "" {
		return c.ListenIP
	}
	if c.ListenIP == "" {
		return ":" + c.ListenPort
	}
	return net.JoinHostPort(c.ListenIP, c.ListenPort)
}
