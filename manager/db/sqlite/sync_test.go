package sqlite

import (
	"testing"
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
