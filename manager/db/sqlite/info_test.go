package sqlite

import (
	"testing"
	"github.com/bitlum/hub/manager/router"
	"reflect"
)

func TestNodeInfoStorage(t *testing.T) {
	db, clear, err := MakeTestDB()
	if err != nil {
		t.Fatalf("unable to create test database: %v", err)
	}
	defer clear()

	infoBefore := &router.Info{
		Version:     "v",
		Network:     "n",
		BlockHeight: 1,
		BlockHash:   "h",
		NodeInfo: &router.NodeInfo{
			Alias:          "a",
			Host:           "h",
			Port:           "p",
			IdentityPubKey: "ik",
		},
		NeutrinoInfo: &router.NeutrinoInfo{
			Host: "h",
			Port: "p",
		},
	}

	if err := db.UpdateInfo(infoBefore); err != nil {
		t.Fatalf("unable to udpate peer info: %v", err)
	}

	infoAfter, err := db.Info()
	if err != nil {
		t.Fatalf("unable to get info: %v", err)
	}

	if !reflect.DeepEqual(infoBefore, infoAfter) {
		t.Fatalf("wrong data")
	}
}
