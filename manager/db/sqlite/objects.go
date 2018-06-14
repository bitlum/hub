package sqlite

import (
	"github.com/jinzhu/gorm"
)

type Counters struct {
	gorm.Model

	ForwardIndex uint32
}

type Channel struct {
	ID     string `gorm:"primary_key"`
	UserID string

	UserBalance   int64
	RouterBalance int64

	// Initiator side which initiated open of the channel.
	Initiator string

	// CloseFee is the number of funds which are needed to close the channel
	// and release locked funds, might change with time, because of the
	// commitment transaction size and fee rate in the network.
	CloseFee int64

	OpenFee int64

	States []State `gorm:"foreignkey:ChannelID"`
}

type State struct {
	ID uint `gorm:"primary_key"`

	ChannelID string

	Time   int64
	Name   string
	Status string
}

type Peer struct {
	ID           string `gorm:"primary_key"`
	Alias        string
	LockedByPeer int64
	LockedByHub  int64
}

type Payment struct {
	gorm.Model

	FromPeer string
	ToPeer   string
	Amount   int64
	Status   string
	Type     string
	Time     int64
}
