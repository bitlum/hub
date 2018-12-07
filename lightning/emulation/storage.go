package emulation

import "github.com/bitlum/hub/lightning"

// TODO(andrew.shvv) emulation network basically is the storage,
// we need use it instead
type StubChannelStorage struct {
}

func (s *StubChannelStorage) AddChannelState(chanID lightning.ChannelID,
	state *lightning.ChannelState) error {
	return nil
}

func (s *StubChannelStorage) UpdateChannel(channels *lightning.Channel) error {
	return nil
}

func (s *StubChannelStorage) Channels() ([]*lightning.Channel, error) {
	return nil, nil
}

func (s *StubChannelStorage) RemoveChannel(channels *lightning.Channel) error {
	return nil
}
