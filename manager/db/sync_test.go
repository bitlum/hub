package db

import (
	"testing"
	"reflect"
)

func TestDB_PutLastForwardingTime(t *testing.T) {
	db, clear, err := MakeTestDB()
	if err != nil {
		t.Fatalf("unable to create test database: %v", err)
	}
	defer clear()

	if err := db.PutLastForwardingTime(156347689); err != nil {
		t.Fatalf("unable to put time: %v", err)
	}

	if time, err := db.LastForwardingTime(); err != nil {
		t.Fatalf("unable to get time: %v", err)
	} else if time != 156347689 {
		t.Fatalf("time is wrong")
	}
}

func TestDB_PutChannelsState(t *testing.T) {
	db, clear, err := MakeTestDB()
	if err != nil {
		t.Fatalf("unable to create test database: %v", err)
	}
	defer clear()

	state1 := map[string]string{
		"kek1": "kek2",
	}

	if err := db.PutChannelsState(state1); err != nil {
		t.Fatalf("unable to put state: %v", err)
	}

	state2, err := db.ChannelsState()
	if err != nil {
		t.Fatalf("unable to get state: %v", err)
	}

	if !reflect.DeepEqual(state1, state2) {
		t.Fatal("state are different")
	}
}
