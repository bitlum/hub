package graphql

import "github.com/bitlum/hub/manager/router"

type GraphQLStorage interface {
	router.InfoStorage
	router.PaymentStorage
	router.UserStorage
}
