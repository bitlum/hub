package lnd

import "github.com/bitlum/hub/manager/router"

// RouterStorage is the storage which is needed for keeping the info about hub
// synchronisation state.
type RouterStorage interface {
	router.ChannelStorage
	router.UserStorage
	router.PaymentStorage
	router.InfoStorage
	SyncStorage
}

type SyncStorage interface {
	// PutLastForwardingIndex is used to save last forward pagination index
	// which was used for getting forwarding events. With this we avoid
	// processing of the same forwarding events twice.
	PutLastForwardingIndex(uint32) error

	// LastForwardingIndex return last lnd forwarding pagination index of
	// which were preceded by the hub.
	LastForwardingIndex() (uint32, error)
}
