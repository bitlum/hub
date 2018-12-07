package graphql

import (
	"github.com/bitlum/hub/lightning"
	"github.com/graphql-go/graphql"
)

func getInfoResolver(storage lightning.InfoStorage) graphql.FieldResolveFn {
	return func(rp graphql.ResolveParams) (
		interface{}, error) {
		return storage.Info()
	}
}

func getPaymentsResolver(storage lightning.PaymentStorage) graphql.FieldResolveFn {
	return func(rp graphql.ResolveParams) (
		interface{}, error) {
		return storage.Payments()
	}
}

func getPeersResolver(storage lightning.UserStorage) graphql.FieldResolveFn {
	return func(rp graphql.ResolveParams) (
		interface{}, error) {
		var users []*lightning.User

		allUsers, err := storage.Users()
		if err != nil {
			return nil, err
		}

		for _, user := range allUsers {
			if user.IsConnected {
				users = append(users, user)
			}
		}

		return users, nil
	}
}
