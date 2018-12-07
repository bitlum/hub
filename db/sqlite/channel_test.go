package sqlite

import (
	"github.com/bitlum/hub/lightning"
	"reflect"
	"testing"
)

func TestDB_AddChannel(t *testing.T) {
	db, clear, err := MakeTestDB()
	if err != nil {
		t.Fatalf("unable to create test database: %v", err)
	}
	defer clear()

	channel1 := &lightning.Channel{
		ChannelID:     "1asdasd1",
		UserID:        "1",
		OpenFee:       1,
		RemoteBalance: 1,
		LocalBalance:  1,
		Initiator:     "i",
		CloseFee:      1,
		States:        []*lightning.ChannelState{},
	}

	if err := db.UpdateChannel(channel1); err != nil {
		t.Fatalf("unable to put state: %v", err)
	}

	channels, err := db.Channels()
	if err != nil {
		t.Fatalf("unable to get state: %v", err)
	}

	if !reflect.DeepEqual(channel1, channels[0]) {
		t.Fatal("state are different")
	}

	state := &lightning.ChannelState{
		Time: 123,
		Name: "opening",
	}

	if err := db.AddChannelState(channel1.ChannelID, state); err != nil {
		t.Fatalf("unable to put state: %v", err)
	}

	if err := db.RemoveChannel(channel1); err != nil {
		t.Fatalf("unable to remove state: %v", err)
	}

	channels, err = db.Channels()
	if err != nil {
		t.Fatalf("unable to get state: %v", err)
	}

	if len(channels) != 0 {
		t.Fatalf("channel hasn't been removed")
	}
}

func TestDB_AddChannelState(t *testing.T) {
	db, clear, err := MakeTestDB()
	if err != nil {
		t.Fatalf("unable to create test database: %v", err)
	}
	defer clear()

	// Ensure that states will be returned in reversed order,
	// last states should be first.
	states := []*lightning.ChannelState{
		{
			Time: 123,
			Name: "opening",
		},
		{
			Time: 125,
			Name: "opened",
		},
	}

	anotherState := &lightning.ChannelState{
		Time: 123,
		Name: "opening",
	}

	// Firstly add opening state.
	if err := db.AddChannelState("1", states[0]); err != nil {
		t.Fatalf("unable to put state: %v", err)
	}

	// Firstly add opened state.
	if err := db.AddChannelState("1", states[1]); err != nil {
		t.Fatalf("unable to put state: %v", err)
	}

	// Ensure that states are filtered by channel id.
	if err := db.AddChannelState("2", anotherState); err != nil {
		t.Fatalf("unable to put state: %v", err)
	}

	channels, err := db.Channels()
	if err != nil {
		t.Fatalf("unable to get state: %v", err)
	}

	if !reflect.DeepEqual(states, channels[0].States) {
		t.Fatal("state are different")
	}
}
