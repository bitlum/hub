package emulation

import "github.com/bitlum/hub/manager/router"

// TODO(andrew.shvv) emulation network basically is the storage,
// we need use it instead
type StubChannelStorage struct {
}

func (s *StubChannelStorage) AddChannelState(chanID router.ChannelID,
	state *router.ChannelState) error {
	return nil
}

func (s *StubChannelStorage) AddChannel(channels *router.Channel) error {
	return nil
}

func (s *StubChannelStorage) Channels() ([]*router.Channel, error) {
	return nil, nil
}

func (s *StubChannelStorage) RemoveChannel(channels *router.Channel) error {
	return nil
}
