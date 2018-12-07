package sqlite

import (
	"github.com/bitlum/hub/lightning"
)

// Runtime check to ensure that DB implements lightning.PaymentStorage interface.
var _ lightning.PaymentStorage = (*DB)(nil)

// StorePayment saves the payment.
func (d *DB) StorePayment(p *lightning.Payment) error {
	payment := &Payment{
		FromUser:  string(p.FromUser),
		ToUser:    string(p.ToUser),
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
func (d *DB) Payments() ([]*lightning.Payment, error) {
	var payments []Payment
	err := d.Model(&Payment{}).
		Order("time DESC").
		Find(&payments).Error
	if err != nil {
		return nil, err
	}

	pmts := make([]*lightning.Payment, len(payments))
	for i, p := range payments {
		pmts[i] = &lightning.Payment{
			FromUser:  lightning.UserID(p.FromUser),
			ToUser:    lightning.UserID(p.ToUser),
			FromAlias: p.FromAlias,
			ToAlias:   p.ToAlias,
			Amount:    lightning.BalanceUnit(p.Amount),
			Status:    lightning.PaymentStatus(p.Status),
			Type:      lightning.PaymentType(p.Type),
			Time:      p.Time,
		}
	}

	return pmts, nil
}
