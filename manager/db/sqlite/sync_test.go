package sqlite

import (
	"testing"
	"reflect"
)

func TestDB_PutLastForwardingIndex(t *testing.T) {
	db, clear, err := MakeTestDB()
	if err != nil {
		t.Fatalf("unable to create test database: %v", err)
	}
	defer clear()

	if err := db.PutLastForwardingIndex(10); err != nil {
		t.Fatalf("unable to put index: %v", err)
	}

	if index, err := db.LastForwardingIndex(); err != nil {
		t.Fatalf("unable to get index: %v", err)
	} else if index != 10 {
		t.Fatalf("index is wrong")
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

	state3 := map[string]string{
		"kek1": "kek100",
	}

	if err := db.PutChannelsState(state3); err != nil {
		t.Fatalf("unable to put state: %v", err)
	}

	state4, err := db.ChannelsState()
	if err != nil {
		t.Fatalf("unable to get state: %v", err)
	}

	if !reflect.DeepEqual(state3, state4) {
		t.Fatal("state are different")
	}
}
