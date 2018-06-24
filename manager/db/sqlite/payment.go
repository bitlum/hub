package sqlite

import (
	"github.com/bitlum/hub/manager/router"
)

// Runtime check to ensure that DB implements router.PaymentStorage interface.
var _ router.PaymentStorage = (*DB)(nil)

// StorePayment saves the payment.
func (d *DB) StorePayment(p *router.Payment) error {
	payment := &Payment{
		FromPeer:  string(p.FromUser),
		ToPeer:    string(p.ToUser),
		FromAlias: string(p.FromAlias),
		ToAlias:   string(p.ToAlias),
		Amount:    int64(p.Amount),
		Status:    string(p.Status),
		Type:      string(p.Type),
		Time:      p.Time,
	}

	return d.Save(payment).Error
}

// Payments returns the payments happening inside the hub local network,
func (d *DB) Payments() ([]*router.Payment, error) {
	var payments []Payment
	db := d.Find(&payments)
	if err := db.Error; err != nil {
		return nil, err
	}

	pmts := make([]*router.Payment, len(payments))
	for i, p := range payments {
		pmts[i] = &router.Payment{
			FromUser: router.UserID(p.FromPeer),
			ToUser:   router.UserID(p.ToPeer),
			Amount:   router.BalanceUnit(p.Amount),
			Status:   router.PaymentStatus(p.Status),
			Type:     router.PaymentType(p.Type),
			Time:     p.Time,
		}
	}

	return pmts, nil
}
