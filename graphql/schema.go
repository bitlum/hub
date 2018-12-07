package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/bitlum/hub/manager/router"
)

var typeNeutrinoInfo = graphql.NewObject(graphql.ObjectConfig{
	Name:        "NeutrinoInfo",
	Description: "",
	Fields: graphql.Fields{
		"host": &graphql.Field{
			Description: "network is a type of currently configured" +
				" blockchain network",
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return rp.Source.(*router.NeutrinoInfo).Host, nil
			},
		},
		"port": &graphql.Field{
			Description: "Host is a public host of service lightning network" +
				" daemon which is used by lightning network wallets to" +
				" connect to service directly",
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return rp.Source.(*router.NeutrinoInfo).Port, nil
			},
		},
	},
})

var typeNodeInfo = graphql.NewObject(graphql.ObjectConfig{
	Name:        "NodeInfo",
	Description: "",
	Fields: graphql.Fields{
		"host": &graphql.Field{
			Description: "network is a type of currently configured" +
				" blockchain network",
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return rp.Source.(*router.NodeInfo).Host, nil
			},
		},
		"port": &graphql.Field{
			Description: "Host is a public host of service lightning network" +
				" daemon which is used by lightning network wallets to" +
				" connect to service directly",
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return rp.Source.(*router.NodeInfo).Port, nil
			},
		},
		"identityPubkey": &graphql.Field{
			Description: "IdentityPubkey is a public key of the service" +
				"lightning network node, which is used on the stage of " +
				"connection to service from user's wallet and have a purpose" +
				" of ensuring that the user is connecting to the proper" +
				" node",
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return rp.Source.(*router.NodeInfo).IdentityPubKey, nil
			},
		},
		"alias": &graphql.Field{
			Description: "Alias is a meaningful name of lightning" +
				" network node, by which node could be found in the lightning " +
				"network explorers. Aliases are not validated, " +
				"and might be taken by any node on the network, " +
				"so it shouldn't be used as identification",
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return rp.Source.(*router.NodeInfo).Alias, nil
			},
		},
	},
})

var typeInfo = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Info",
	Description: "",
	Fields: graphql.Fields{
		"network": &graphql.Field{
			Description: "Network is a type of currently configured" +
				" blockchain network",
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return rp.Source.(*router.Info).Network, nil
			},
		},
		"version": &graphql.Field{
			Description: "Version is version of lightning network daemon",
			Type:        graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return rp.Source.(*router.Info).Version, nil
			},
		},
		"blockHeight": &graphql.Field{
			Description: "BlockHeight is service lightning network node's" +
				" current view of the height of chain",
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return rp.Source.(*router.Info).BlockHeight, nil
			},
		},
		"blockHash": &graphql.Field{
			Description: "BlockHash is service lightning network node's" +
				" current view of the hash of the best block",
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return rp.Source.(*router.Info).BlockHash, nil
			},
		},

		"neutrinoInfo": &graphql.Field{
			Description: "NeutrinoInfo is an information which is need to" +
				" connect to the neutrino node",
			Type: graphql.NewNonNull(typeNeutrinoInfo),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return rp.Source.(*router.Info).NeutrinoInfo, nil
			},
		},
		"nodeInfo": &graphql.Field{
			Description: "NodeInfo is an information which is needed to" +
				" connect or find lightning network node",
			Type: graphql.NewNonNull(typeNodeInfo),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return rp.Source.(*router.Info).NodeInfo, nil
			},
		},
	},
})

var typePeer = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Peer",
	Description: "",
	Fields: graphql.Fields{
		"alias": &graphql.Field{
			Description: "",
			Type:        graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return rp.Source.(*router.User).Alias, nil
			},
		},
		"lockedByPeer": &graphql.Field{
			Description: "",
			Type:        graphql.NewNonNull(graphql.Int),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return int64(rp.Source.(*router.User).LockedByUser), nil
			},
		},
		"lockedByHub": &graphql.Field{
			Description: "",
			Type:        graphql.NewNonNull(graphql.Int),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return int64(rp.Source.(*router.User).LockedByHub), nil
			},
		},
	},
})

var typePayment = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Payment",
	Description: "",
	Fields: graphql.Fields{
		"fromPeer": &graphql.Field{
			Description: "",
			Type:        graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return rp.Source.(*router.Payment).FromAlias, nil
			},
		},
		"toPeer": &graphql.Field{
			Description: "",
			Type:        graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return rp.Source.(*router.Payment).ToAlias, nil
			},
		},
		"amount": &graphql.Field{
			Description: "",
			Type:        graphql.NewNonNull(graphql.Int),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return int64(rp.Source.(*router.Payment).Amount), nil
			},
		},
		"status": &graphql.Field{
			Description: "",
			Type:        graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return string(rp.Source.(*router.Payment).Status), nil
			},
		},
		"time": &graphql.Field{
			Description: "",
			Type:        graphql.NewNonNull(graphql.Int),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return rp.Source.(*router.Payment).Time, nil
			},
		},
		"type": &graphql.Field{
			Description: "",
			Type:        graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return string(rp.Source.(*router.Payment).Type), nil
			},
		},
	},
})

func New(storage GraphQLStorage) (graphql.Schema, error) {
	return graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Description: "Queries are the type of requests which are not" +
				" mutate any data and used for fetching information about the" +
				" state of objects, setting, service configurations",
			Fields: graphql.Fields{
				"info": &graphql.Field{
					Description: "",
					Type:        graphql.NewNonNull(typeInfo),
					Resolve:     getInfoResolver(storage),
				},

				"peers": &graphql.Field{
					Description: "",
					Type:        graphql.NewNonNull(graphql.NewList(typePeer)),
					Resolve:     getPeersResolver(storage),
				},

				"payments": &graphql.Field{
					Description: "",
					Type:        graphql.NewNonNull(graphql.NewList(typePayment)),
					Resolve:     getPaymentsResolver(storage),
				},
			},
		}),
	})
}
