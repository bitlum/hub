package sqlite

import (
	"github.com/bitlum/hub/manager/router"
	"time"
	"reflect"
	"testing"
)

func TestDB_AddChannel(t *testing.T) {
	db, clear, err := MakeTestDB()
	if err != nil {
		t.Fatalf("unable to create test database: %v", err)
	}
	defer clear()

	channel1 := &router.Channel{
		ChannelID:     "1",
		UserID:        "1",
		OpenFee:       1,
		UserBalance:   1,
		RouterBalance: 1,
		Initiator:     "i",
		CloseFee:      1,
		States:        []*router.ChannelState{},
	}

	if err := db.AddChannel(channel1); err != nil {
		t.Fatalf("unable to put state: %v", err)
	}

	channels, err := db.Channels()
	if err != nil {
		t.Fatalf("unable to get state: %v", err)
	}

	if !reflect.DeepEqual(channel1, channels[0]) {
		t.Fatal("state are different")
	}
}

func TestDB_AddChannelState(t *testing.T) {
	db, clear, err := MakeTestDB()
	if err != nil {
		t.Fatalf("unable to create test database: %v", err)
	}
	defer clear()

	state := &router.ChannelState{
		Time:   time.Now(),
		Name:   "n",
		Status: "s",
	}

	if err := db.AddChannelState("1", state); err != nil {
		t.Fatalf("unable to put state: %v", err)
	}

	channels, err := db.Channels()
	if err != nil {
		t.Fatalf("unable to get state: %v", err)
	}

	// NOTE: Comparing fields one by one because of the time field inside the
	// structure which couldn't be compared with deep equal properly.
	if !reflect.DeepEqual(state.Status, channels[0].States[0].Status) {
		t.Fatal("state are different")
	}

	if !reflect.DeepEqual(state.Name, channels[0].States[0].Name) {
		t.Fatal("state are different")
	}
}
