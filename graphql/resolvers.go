package graphql

import (
	"github.com/bitlum/hub/lightning"
	"github.com/go-errors/errors"
	"github.com/graphql-go/graphql"
	"sort"
	"strings"
)

func getInfoResolver(client lightning.Client) graphql.FieldResolveFn {
	return func(rp graphql.ResolveParams) (
		interface{}, error) {
		return client.Info()
	}
}

func getPaymentsResolver(client lightning.PaymentClient,
	getAlias func(nodeID lightning.NodeID) string) graphql.FieldResolveFn {
	return func(rp graphql.ResolveParams) (
		interface{}, error) {

		var payments []*Payment

		directPayments, err := client.ListPayments("", lightning.AllStatuses,
			lightning.AllDirections, lightning.AllSystems)
		if err != nil {
			return nil, errors.Errorf("unable to list payments: %v", err)
		}

		forwardPayments, err := client.ListForwardPayments()
		if err != nil {
			return nil, errors.Errorf("unable to list forward payments: %v",
				err)
		}

		for _, payment := range directPayments {
			if payment.Status != lightning.Completed {
				continue
			}

			var fromNode, toNode string
			if payment.Direction == lightning.Outgoing {
				fromNode = "bitlum.io"
				toNode = getAlias(payment.Receiver)
			} else if payment.Direction == lightning.Incoming {
				toNode = "bitlum.io"
				fromNode = getAlias("random string")
			}

			payments = append(payments, &Payment{
				FromNode: fromNode,
				ToNode:   toNode,
				Amount:   int64(payment.Amount),
				// TODO(andrew.shvv) remove compatibility
				Status: "successful",
				Time:   payment.UpdatedAt,
				Type:   strings.ToLower(string(payment.Direction)),
			})
		}

		for _, payment := range forwardPayments {
			payments = append(payments, &Payment{
				FromNode: getAlias(payment.FromNode),
				ToNode:   getAlias(payment.ToNode),
				Amount:   int64(payment.OutgoingAmount),
				// TODO(andrew.shvv) remove compatibility
				Status: "successful",
				Time:   payment.Time,
				// TODO(andrew.shvv) remove compatibility
				Type: "forward",
			})
		}

		sort.Slice(payments, func(i, j int) bool {
			return payments[i].Time > payments[j].Time
		})

		return payments, nil
	}
}
