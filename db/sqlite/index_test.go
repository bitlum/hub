package sqlite

import "testing"

func TestChannelPointIndex(t *testing.T) {
	db, clear, err := MakeTestDB()
	if err != nil {
		t.Fatalf("unable to create test database: %v", err)
	}
	defer clear()

	_, err = db.GetChannelPointByShortChanID(1)
	if err == nil {
		t.Fatalf("index shoudn't be found")
	}

	err = db.AddChannelPointToShortChanIDIndex("1", 1)
	if err != nil {
		t.Fatalf("unbale to add index: %v", err)
	}

	err = db.AddChannelPointToShortChanIDIndex("2", 2)
	if err != nil {
		t.Fatalf("unbale to add index: %v", err)
	}

	chanID, err := db.GetChannelPointByShortChanID(1)
	if err != nil {
		t.Fatalf("unable fund index: %v", err)
	}

	if chanID != "1" {
		t.Fatalf("wrong index")
	}
}

func TestUserIndex(t *testing.T) {
	db, clear, err := MakeTestDB()
	if err != nil {
		t.Fatalf("unable to create test database: %v", err)
	}
	defer clear()

	_, err = db.GetUserIDByShortChanID(1)
	if err == nil {
		t.Fatalf("index shoudn't be found")
	}

	err = db.AddUserIDToShortChanIDIndex("1", 1)
	if err != nil {
		t.Fatalf("unbale to add index: %v", err)
	}

	err = db.AddUserIDToShortChanIDIndex("2", 2)
	if err != nil {
		t.Fatalf("unbale to add index: %v", err)
	}

	userID, err := db.GetUserIDByShortChanID(1)
	if err != nil {
		t.Fatalf("unable fund index: %v", err)
	}

	if userID != "1" {
		t.Fatalf("wrong index")
	}
}
