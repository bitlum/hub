package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/bitlum/hub/manager/router"
)

func getInfoResolver(storage router.InfoStorage) graphql.FieldResolveFn {
	return func(rp graphql.ResolveParams) (
		interface{}, error) {
		return storage.Info()
	}
}

func getPaymentsResolver(storage router.PaymentStorage) graphql.FieldResolveFn {
	return func(rp graphql.ResolveParams) (
		interface{}, error) {
		return storage.Payments()
	}
}

func getPeersResolver(storage router.UserStorage) graphql.FieldResolveFn {
	return func(rp graphql.ResolveParams) (
		interface{}, error) {
		return storage.Users()
	}
}
