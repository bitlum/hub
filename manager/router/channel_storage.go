package router

type ChannelStorage interface {
	// AddChannelState adds state to the channel's state array. State array
	// should be initialised in the Channel object on the stage of getting
	// channels.
	AddChannelState(chanID ChannelID, state *ChannelState) error

	// AddChannel saves channel without saving its states.
	AddChannel(channels *Channel) error

	// Channels returns all channels and its states.
	Channels() ([]*Channel, error)

	// RemoveChannel removes the channel.
	RemoveChannel(channels *Channel) error
}
