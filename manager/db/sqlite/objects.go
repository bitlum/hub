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

	// CloseFee is the number of funds which were needed to open the channel
	// and lock funds.
	OpenFee int64

	// IsUserConnected is used to determine is user connected with tcp/ip
	// connection to the hub, which means that this channel could be used for
	// payments.
	IsUserConnected bool

	// States is the array of states which this channel went thorough.
	States []State `gorm:"foreignkey:ChannelID"`
}

type State struct {
	ID uint `gorm:"primary_key"`

	ChannelID string

	Time int64
	Name string
}

type User struct {
	ID           string `gorm:"primary_key"`
	IsConnected  bool
	Alias        string
	LockedByUser int64
	LockedByHub  int64
}

type Payment struct {
	gorm.Model

	FromPeer string
	ToPeer   string

	FromAlias string
	ToAlias   string

	Amount int64
	Status string
	Type   string
	Time   int64
}
