package main

import (
	"fmt"
	"github.com/bitlum/hub/hubrpc"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"os"
)

const (
	defaultRPCPort     = "8686"
	defaultRPCHostPort = "localhost:" + defaultRPCPort
)

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "[pscli] %v\n", err)
	os.Exit(1)
}

func getClient(ctx *cli.Context) (hubrpc.HubClient, func()) {
	conn := getClientConn(ctx, false)

	cleanUp := func() {
		conn.Close()
	}

	return hubrpc.NewHubClient(conn), cleanUp
}

func getClientConn(ctx *cli.Context, skipMacaroons bool) *grpc.ClientConn {
	// Create a dial options array.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial(ctx.GlobalString("rpcserver"), opts...)
	if err != nil {
		fatal(err)
	}

	return conn
}

func main() {
	app := cli.NewApp()
	app.Name = "hubcli"
	app.Version = fmt.Sprintf("0.1")
	app.Usage = "Control plane for your Hub Daemon (hubd)"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "rpcserver",
			Value: defaultRPCHostPort,
			Usage: "host:port of hub rpc",
		},
	}
	app.Commands = []cli.Command{
		createInvoiceCommand,
		validateInvoiceCommand,
		balanceCommand,
		estimateFeeCommand,
		sendPaymentCommand,
		paymentByIDCommand,
		paymentByInvoiceCommand,
		listPaymentsCommand,
		checkNodeStatsCommand,
	}

	if err := app.Run(os.Args); err != nil {
		fatal(err)
	}
}
