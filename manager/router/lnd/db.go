package lnd

type DB interface {
	// PutLastForwardingIndex is used to save last forward pagination index
	// which was used for getting forwarding events. With this we avoid
	// processing of the same forwarding events twice.
	PutLastForwardingIndex(uint32) error

	// LastForwardingIndex return last lnd forwarding pagination index of
	// which were preceded by the hub.
	LastForwardingIndex() (uint32, error)

	// PutChannelsState is used to save the local topology of the router,
	// in order to later determine what has changed.
	PutChannelsState(map[string]string) error

	// ChannelsState is used to return previously saved local topology of the
	// router.
	ChannelsState() (map[string]string, error)
}

type InMemory struct {
	lastIndex     uint32
	channelsState map[string]string
}

func (db *InMemory) StartUpdate() {}
func (db *InMemory) Flush()       {}

func (db *InMemory) PutLastForwardingIndex(lastIndex uint32) error {
	db.lastIndex = lastIndex
	return nil
}

func (db *InMemory) LastForwardingIndex() (uint32, error) {
	return db.lastIndex, nil
}

func (db *InMemory) ChannelsState() (map[string]string, error) {
	return db.channelsState, nil
}

func (db *InMemory) PutChannelsState(channelsState map[string]string) error {
	db.channelsState = channelsState
	return nil
}
