package sqlite

import (
	"testing"
	"github.com/bitlum/hub/manager/router"
	"reflect"
)

func TestPeersStorage(t *testing.T) {
	db, clear, err := MakeTestDB()
	if err != nil {
		t.Fatalf("unable to create test database: %v", err)
	}
	defer clear()

	peersBefore := []*router.DbPeer{
		{
			PubKey:       "1",
			Alias:        "a",
			LockedByPeer: 1,
			LockedByHub:  1,
		},
		{
			PubKey:       "2",
			Alias:        "b",
			LockedByPeer: 2,
			LockedByHub:  2,
		},
	}

	if err := db.UpdatePeers(peersBefore); err != nil {
		t.Fatalf("unable to udpate peers: %v", err)
	}

	peersAfter, err := db.Peers()
	if err != nil {
		t.Fatalf("unable to get peers: %v", err)
	}

	if !reflect.DeepEqual(peersBefore, peersAfter) {
		t.Fatalf("wrong data")
	}
}

func TestPaymentsStorage(t *testing.T) {
	db, clear, err := MakeTestDB()
	if err != nil {
		t.Fatalf("unable to create test database: %v", err)
	}
	defer clear()

	paymentsBefore := []*router.DbPayment{
		{
			FromPeer: "a",
			ToPeer:   "b",
			Amount:   1,
			Status:   "s",
			Type:     "k",
			Time:     123,
		},
		{
			FromPeer: "a",
			ToPeer:   "b",
			Amount:   2,
			Status:   "s",
			Type:     "k",
			Time:     124,
		},
	}

	if err := db.StorePayment(paymentsBefore[0]); err != nil {
		t.Fatalf("unable to save payment: %v", err)
	}

	if err := db.StorePayment(paymentsBefore[1]); err != nil {
		t.Fatalf("unable to save payment: %v", err)
	}

	paymentsAfter, err := db.Payments()
	if err != nil {
		t.Fatalf("unable to get payments: %v", err)
	}

	if !reflect.DeepEqual(paymentsBefore, paymentsAfter) {
		t.Fatalf("wrong data")
	}
}

func TestNodeInfoStorage(t *testing.T) {
	db, clear, err := MakeTestDB()
	if err != nil {
		t.Fatalf("unable to create test database: %v", err)
	}
	defer clear()

	infoBefore := &router.DbInfo{
		Version:     "v",
		Network:     "n",
		BlockHeight: 1,
		BlockHash:   "h",
		NodeInfo: &router.DbNodeInfo{
			Alias:          "a",
			Host:           "h",
			Port:           "p",
			IdentityPubKey: "ik",
		},
		NeutrinoInfo: &router.DbNeutrinoInfo{
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
