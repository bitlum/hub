package sqlite

import (
	"github.com/bitlum/hub/lightning"
	"reflect"
	"testing"
)

func TestUserStorage(t *testing.T) {
	db, clear, err := MakeTestDB()
	if err != nil {
		t.Fatalf("unable to create test database: %v", err)
	}
	defer clear()

	userBefore := &lightning.User{
		UserID:       "1",
		Alias:        "a",
		LockedByUser: 1,
		LockedByHub:  1,
		IsConnected:  true,
	}

	if err := db.UpdateUser(userBefore); err != nil {
		t.Fatalf("unable to udpate peers: %v", err)
	}

	usersAfter, err := db.Users()
	if err != nil {
		t.Fatalf("unable to get peers: %v", err)
	}

	if !reflect.DeepEqual(userBefore, usersAfter[0]) {
		t.Fatalf("wrong data")
	}
}
