package sqlite

import (
	"github.com/jinzhu/gorm"
)

type Counters struct {
	gorm.Model

	ForwardIndex uint32
}

type Channel struct {
	ID    string `gorm:"primary_key"`
	State string `gorm:"type:varchar(20)"`
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
