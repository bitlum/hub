package main

import (
	"fmt"
	"github.com/bitlum/hub/hubrpc"
	"github.com/go-errors/errors"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
	"strings"
)

func printRespJSON(resp proto.Message) {
	jsonMarshaler := &jsonpb.Marshaler{
		EmitDefaults: true,
		Indent:       "    ",
	}

	jsonStr, err := jsonMarshaler.MarshalToString(resp)
	if err != nil {
		fmt.Println("unable to decode response: ", err)
		return
	}

	fmt.Println(jsonStr)
}

var createInvoiceCommand = cli.Command{
	Name:     "createinvoice",
	Category: "Invoice",
	Usage:    "Generates new invoice.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "amount",
			Usage: "(optional) Amount is the amount which should be received on this " +

				"receipt.",
		},
		cli.StringFlag{
			Name: "description",
			Usage: "(optional) This description will be placed in the invoice " +
				"itself, which would allow user to see what he paid for later " +
				"in the wallet.",
		},
	},
	Action: createInvoice,
}

func createInvoice(ctx *cli.Context) error {
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	var (
		amount      string
		description string
	)

	if ctx.IsSet("amount") {
		amount = ctx.String("amount")
	}

	if ctx.IsSet("description") {
		description = ctx.String("description")
	}

	ctxb := context.Background()
	resp, err := client.CreateInvoice(ctxb, &hubrpc.CreateInvoiceRequest{
		Amount:      amount,
		Description: description,
	})
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}

var validateInvoiceCommand = cli.Command{
	Name:     "validateinvoice",
	Category: "Invoice",
	Usage:    "Validates given invoice.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "receipt",
			Usage: "Receipt is either blockchain address or lightning network.",
		},
		cli.StringFlag{
			Name: "amount",
			Usage: "(optional) Amount is the amount which should be received " +
				"on this receipt.",
		},
	},
	Action: validateInvoice,
}

func validateInvoice(ctx *cli.Context) error {
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	var (
		amount  string
		invoice string
	)

	if ctx.IsSet("amount") {
		amount = ctx.String("amount")
	}

	if ctx.IsSet("invoice") {
		invoice = ctx.String("invoice")
	} else {
		return errors.Errorf("invoice argument is missing")
	}

	ctxb := context.Background()
	resp, err := client.ValidateInvoice(ctxb, &hubrpc.ValidateInvoiceRequest{
		Amount:  amount,
		Invoice: invoice,
	})
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}

var balanceCommand = cli.Command{
	Name:     "balance",
	Category: "Balance",
	Usage:    "Return number of funds locked in the channels.",
	Action:   balance,
}

func balance(ctx *cli.Context) error {
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	ctxb := context.Background()
	resp, err := client.Balance(ctxb, &hubrpc.BalanceRequest{})
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}

var estimateFeeCommand = cli.Command{
	Name:     "estimatefee",
	Category: "Fee",
	Usage:    "Estimates fee of the payment.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "amount",
			Usage: "(optional) Amount is the amount which will be sent by" +
				" service.",
		},
		cli.StringFlag{
			Name: "invoice",
			Usage: "(optional) Invoice it is lightning network invoice, " +
				"which is the string which contains amount, description, " +
				"destination, and other info which is needed for sender " +
				"to successfully send payment.",
		},
	},
	Action: estimateFee,
}

func estimateFee(ctx *cli.Context) error {
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	var (
		amount  string
		invoice string
	)

	if ctx.IsSet("amount") {
		amount = ctx.String("amount")
	}

	if ctx.IsSet("invoice") {
		invoice = ctx.String("invoice")
	}

	ctxb := context.Background()
	resp, err := client.EstimateFee(ctxb, &hubrpc.EstimateFeeRequest{
		Amount:  amount,
		Invoice: invoice,
	})
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}

var sendPaymentCommand = cli.Command{
	Name:     "sendpayment",
	Category: "Payment",
	Usage:    "Sends payment",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "amount",
			Usage: "(optional) Amount is the amount which will be sent by" +
				" service.",
		},
		cli.StringFlag{
			Name: "invoice",
			Usage: "Invoice it is lightning network invoice, " +
				"which is the string which contains amount, description, " +
				"destination, and other info which is needed for sender " +
				"to successfully send payment.",
		},
	},
	Action: sendPayment,
}

func sendPayment(ctx *cli.Context) error {
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	var (
		amount  string
		invoice string
	)

	if ctx.IsSet("amount") {
		amount = ctx.String("amount")
	}

	if ctx.IsSet("invoice") {
		invoice = ctx.String("invoice")
	} else {
		return errors.Errorf("invoice argument is missing")
	}

	ctxb := context.Background()
	resp, err := client.SendPayment(ctxb, &hubrpc.SendPaymentRequest{
		Amount:  amount,
		Invoice: invoice,
	})
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}

var paymentByIDCommand = cli.Command{
	Name:     "paymentbyid",
	Category: "Payment",
	Usage:    "Return payment by the given id",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "id",
			Usage: "ID it is unique identificator of payment",
		},
	},
	Action: paymentByID,
}

func paymentByID(ctx *cli.Context) error {
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	var id string

	if ctx.IsSet("id") {
		id = ctx.String("id")
	} else {
		return errors.Errorf("id argument is missing")
	}

	ctxb := context.Background()
	resp, err := client.PaymentByID(ctxb, &hubrpc.PaymentByIDRequest{
		PaymentId: id,
	})
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}

var paymentByInvoiceCommand = cli.Command{
	Name:     "paymentbyinvoice",
	Category: "Payment",
	Usage:    "Return payment by the given invoice",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "invoice",
			Usage: "Invoice it is lightning network invoice, which is the " +
				"string which contains amount, description, destination, " +
				"and other info which is needed for sender to successfully" +
				" send payment.",
		},
	},
	Action: paymentByInvoice,
}

func paymentByInvoice(ctx *cli.Context) error {
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	var invoice string

	if ctx.IsSet("invoice") {
		invoice = ctx.String("invoice")
	} else {
		return errors.Errorf("invoice argument is missing")
	}

	ctxb := context.Background()
	resp, err := client.PaymentByInvoice(ctxb, &hubrpc.PaymentByInvoiceRequest{
		Invoice: invoice,
	})
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}

var listPaymentsCommand = cli.Command{
	Name:     "listpayments",
	Category: "Payment",
	Usage:    "Return list payments by the given filter parameters",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "direction",
			Usage: "Direction identifies the direction of the payment, " +
				"(incoming, outgoing).",
		},
		cli.StringFlag{
			Name: "status",
			Usage: "Status is the state of the payment, " +
				"(waiting, pending, completed, failed).",
		},
		cli.StringFlag{
			Name: "system",
			Usage: "System denotes is that payment belongs to business logic" +
				" of payment server or it was originated by " +
				"user / third-party service (internal, external).",
		},
	},
	Action: listPayments,
}

func listPayments(ctx *cli.Context) error {
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	var (
		status    hubrpc.PaymentStatus
		direction hubrpc.PaymentDirection
		system    hubrpc.PaymentSystem
	)

	if ctx.IsSet("status") {
		stringStatus := strings.ToLower(ctx.String("status"))
		switch stringStatus {
		case strings.ToLower(hubrpc.PaymentStatus_WAITING.String()):
			status = hubrpc.PaymentStatus_WAITING

		case strings.ToLower(hubrpc.PaymentStatus_PENDING.String()):
			status = hubrpc.PaymentStatus_PENDING

		case strings.ToLower(hubrpc.PaymentStatus_COMPLETED.String()):
			status = hubrpc.PaymentStatus_COMPLETED

		case strings.ToLower(hubrpc.PaymentStatus_FAILED.String()):
			status = hubrpc.PaymentStatus_FAILED
		default:
			return errors.Errorf("invalid status %v, supported statuses"+
				"are: 'waiting', 'pending', 'completed', 'failed'",
				stringStatus)
		}
	}

	if ctx.IsSet("direction") {
		stringDirection := strings.ToLower(ctx.String("direction"))
		switch stringDirection {

		case strings.ToLower(hubrpc.PaymentDirection_OUTGOING.String()):
			direction = hubrpc.PaymentDirection_OUTGOING

		case strings.ToLower(hubrpc.PaymentDirection_INCOMING.String()):
			direction = hubrpc.PaymentDirection_INCOMING

		default:
			return errors.Errorf("invalid direction %v, supported direction"+
				"are: 'incoming', 'outgoing'",
				stringDirection)
		}
	}

	if ctx.IsSet("system") {
		stringSystem := strings.ToLower(ctx.String("system"))
		switch stringSystem {
		case strings.ToLower(hubrpc.PaymentSystem_INTERNAL.String()):
			system = hubrpc.PaymentSystem_INTERNAL

		case strings.ToLower(hubrpc.PaymentSystem_EXTERNAL.String()):
			system = hubrpc.PaymentSystem_EXTERNAL

		default:
			return errors.Errorf("invalid system %v, supported system"+
				"are: 'internal', 'external'",
				stringSystem)
		}
	}

	ctxb := context.Background()
	resp, err := client.ListPayments(ctxb, &hubrpc.ListPaymentsRequest{
		Status:    status,
		Direction: direction,
		System:    system,
	})
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}

var checkNodeStatsCommand = cli.Command{
	Name:     "nodestats",
	Category: "Nodes",
	Usage:    "Return statistical information about nodes in ln",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "period",
			Usage: "(optional) Period of time for which we should calculate" +
				" statistic, (day, week, month, three month)",
		},
		cli.StringFlag{
			Name: "node",
			Usage: "(optional) Node is public key or node name of about which" +
				" we should show statistic, by default it will show all",
		},
		cli.StringFlag{
			Name:  "limit",
			Usage: "(optional) Limit output to the given number",
		},
		cli.StringFlag{
			Name: "sort_type",
			Usage: "(optional) Sorts nodes by the given sorting algo, " +
				"by default by volume, (by_sent_num, " +
				"by_idleness, by_volume)",
		},
	},
	Action: checkNodeStats,
}

func checkNodeStats(ctx *cli.Context) error {
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	var (
		period   hubrpc.Period
		node     string
		limit    int64
		sortType hubrpc.SortType
	)

	if ctx.IsSet("period") {
		p := ctx.String("period")
		switch p {
		case "day":
			period = hubrpc.Period_DAY
		case "week":
			period = hubrpc.Period_WEEK
		case "month":
			period = hubrpc.Period_MONTH
		case "three month":
			period = hubrpc.Period_THREE_MONTH
		default:
			return errors.Errorf("unknown period option(%v)", p)
		}
	}

	if ctx.IsSet("node") {
		node = ctx.String("node")
	}

	if ctx.IsSet("limit") {
		limit = ctx.Int64("limit")
	}

	if ctx.IsSet("sort_type") {
		st := ctx.String("sort_type")
		switch st {
		case "by_sent_num":
			sortType = hubrpc.SortType_BY_SENT_NUM
		case "by_idleness":
			sortType = hubrpc.SortType_BY_IDLENESS
		case "by_volume":
			sortType = hubrpc.SortType_BY_VOLUME
		default:
			return errors.Errorf("unknown sort type(%v)", st)
		}
	}

	ctxb := context.Background()
	resp, err := client.CheckNodeStats(ctxb, &hubrpc.CheckNodeStatsRequest{
		Period:   period,
		Node:     node,
		Limit:    int32(limit),
		SortType: sortType,
	})
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}
