package router

// PaymentError...
type PaymentError string

const (
	// InsufficientFunds means that lightning node haven't posses/locked enough
	// funds with receiver peer to route through the payment.
	InsufficientFunds PaymentError = "insufficient_funds"

	// UserNotFound means that lightning node wasn't able to forward payment
	// because of the receiver peer not being connected.
	UserNotFound PaymentError = "user_not_found"

	// ExternalFail means that receiver failed to receive payment because of
	// the unknown to us reason.
	ExternalFail PaymentError = "external_fail"

	// UserLocalFail means that from user's side all channel are in
	// pending states or not exist at all, or number of funds from user side
	// is not enough.
	UserLocalFail PaymentError = "user_local_fail"
)
