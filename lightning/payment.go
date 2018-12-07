package lightning

type Payment struct {
	FromUser UserID
	ToUser   UserID

	FromAlias string
	ToAlias   string

	Amount BalanceUnit

	Status PaymentStatus
	Type   PaymentType

	Time int64
}
