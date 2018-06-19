package lnd

import "github.com/bitlum/hub/manager/router"

// SyncStorage is the storage which is needed for keeping the info about hub
// synchronisation state.
type SyncStorage interface {
	router.ChannelStorage

	// PutLastForwardingIndex is used to save last forward pagination index
	// which was used for getting forwarding events. With this we avoid
	// processing of the same forwarding events twice.
	PutLastForwardingIndex(uint32) error

	// LastForwardingIndex return last lnd forwarding pagination index of
	// which were preceded by the hub.
	LastForwardingIndex() (uint32, error)
}

type InMemorySyncStorage struct {
	lastIndex uint32
}

func (db *InMemorySyncStorage) PutLastForwardingIndex(lastIndex uint32) error {
	db.lastIndex = lastIndex
	return nil
}

func (db *InMemorySyncStorage) LastForwardingIndex() (uint32, error) {
	return db.lastIndex, nil
}
