package graphql

import "github.com/bitlum/hub/lightning"

type GraphQLStorage interface {
	lightning.InfoStorage
	lightning.PaymentStorage
	lightning.UserStorage
}
