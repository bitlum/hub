package sqlite

import (
	"github.com/bitlum/hub/lightning"
	"reflect"
	"testing"
)

func TestPaymentsStorage(t *testing.T) {
	db, clear, err := MakeTestDB()
	if err != nil {
		t.Fatalf("unable to create test database: %v", err)
	}
	defer clear()

	paymentsBefore := []*lightning.Payment{
		{
			FromUser:  "a",
			ToUser:    "b",
			FromAlias: "a",
			ToAlias:   "b",
			Amount:    1,
			Status:    "s",
			Type:      "k",
			Time:      125,
		},
		{
			FromUser:  "a",
			ToUser:    "b",
			FromAlias: "a",
			ToAlias:   "b",
			Amount:    2,
			Status:    "s",
			Type:      "k",
			Time:      123,
		},
	}

	if err := db.StorePayment(paymentsBefore[1]); err != nil {
		t.Fatalf("unable to save payment: %v", err)
	}

	if err := db.StorePayment(paymentsBefore[0]); err != nil {
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
