package lnd

type DB interface {
	StartUpdate()
	Flush()

	PutLastForwardingIndex(uint32) error
	LastForwardingIndex() (uint32, error)

	ChannelsState() (map[string]string, error)
	PutChannelsState(map[string]string) error
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
