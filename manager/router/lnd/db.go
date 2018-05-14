package lnd

// SyncStorage is the storage which is needed for keeping the info about hub
// synchronisation state.
type SyncStorage interface {
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

type InMemorySyncStorage struct {
	lastIndex     uint32
	channelsState map[string]string
}

func (db *InMemorySyncStorage) StartUpdate() {}
func (db *InMemorySyncStorage) Flush()       {}

func (db *InMemorySyncStorage) PutLastForwardingIndex(lastIndex uint32) error {
	db.lastIndex = lastIndex
	return nil
}

func (db *InMemorySyncStorage) LastForwardingIndex() (uint32, error) {
	return db.lastIndex, nil
}

func (db *InMemorySyncStorage) ChannelsState() (map[string]string, error) {
	return db.channelsState, nil
}

func (db *InMemorySyncStorage) PutChannelsState(channelsState map[string]string) error {
	db.channelsState = channelsState
	return nil
}
