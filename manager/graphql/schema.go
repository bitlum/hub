package graphql

import (
	"github.com/graphql-go/graphql"
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
				return "neutrinohost", nil
			},
		},
		"port": &graphql.Field{
			Description: "Host is a public host of service lightning network" +
				" daemon which is used by lightning network wallets to" +
				" connect to service directly",
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return "neutrinoport", nil
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
				return "neutrinohost", nil
			},
		},
		"port": &graphql.Field{
			Description: "Host is a public host of service lightning network" +
				" daemon which is used by lightning network wallets to" +
				" connect to service directly",
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return "neutrinoport", nil
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
				return "kekkey", nil
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
				return "kekalias", nil
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
				return "keknetwork", nil
			},
		},
		"version": &graphql.Field{
			Description: "Version is version of lightning network daemon",
			Type:        graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return "kekversion", nil
			},
		},
		"blockHeight": &graphql.Field{
			Description: "BlockHeight is service lightning network node's" +
				" current view of the height of chain",
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return 1317, nil
			},
		},
		"blockHash": &graphql.Field{
			Description: "BlockHash is service lightning network node's" +
				" current view of the hash of the best block",
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return "kekblockchash", nil
			},
		},

		"neutrinoInfo": &graphql.Field{
			Description: "NeutrinoInfo is an information which is need to" +
				" connect to the neutrino node",
			Type: graphql.NewNonNull(typeNeutrinoInfo),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return struct{}{}, nil
			},
		},
		"nodeInfo": &graphql.Field{
			Description: "NodeInfo is an information which is needed to" +
				" connect or find lightning network node",
			Type: graphql.NewNonNull(typeNodeInfo),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return struct{}{}, nil
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
				return "kekalias", nil
			},
		},
		"lockedByPeer": &graphql.Field{
			Description: "",
			Type:        graphql.NewNonNull(graphql.Int),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return 100, nil
			},
		},
		"lockedByHub": &graphql.Field{
			Description: "",
			Type:        graphql.NewNonNull(graphql.Int),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return 100, nil
			},
		},
		"isActive": &graphql.Field{
			Description: "",
			Type:        graphql.NewNonNull(graphql.Boolean),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return "kekstatus", nil
			},
		},
		"lastUpdate": &graphql.Field{
			Description: "",
			Type:        graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return "keklastupdate", nil
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
				return "kekfrompeer", nil
			},
		},
		"toPeer": &graphql.Field{
			Description: "",
			Type:        graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return "kekversion", nil
			},
		},
		"amount": &graphql.Field{
			Description: "",
			Type:        graphql.NewNonNull(graphql.Int),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return 1000, nil
			},
		},
		"status": &graphql.Field{
			Description: "",
			Type:        graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return "kekstatus", nil
			},
		},
		"time": &graphql.Field{
			Description: "",
			Type:        graphql.NewNonNull(graphql.Int),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return 19237861, nil
			},
		},
		"paymentID": &graphql.Field{
			Description: "",
			Type:        graphql.NewNonNull(graphql.String),
			Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
				return "kekpaymentid", nil
			},
		},
	},
})

func New() (graphql.Schema, error) {
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
					Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
						return struct{}{}, nil
					},
				},

				"peers": &graphql.Field{
					Description: "",
					Type:        graphql.NewNonNull(graphql.NewList(typePeer)),
					Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
						return []struct{}{{}}, nil
					},
				},

				"payments": &graphql.Field{
					Description: "",
					Type:        graphql.NewNonNull(graphql.NewList(typePayment)),
					Resolve: func(rp graphql.ResolveParams) (interface{}, error) {
						return []struct{}{{}}, nil
					},
				},
			},
		}),
	})
}
