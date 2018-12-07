package main

import (
	"github.com/bitlum/hub/lightning/emulation"
	"google.golang.org/grpc"
	"os"
	"context"
	"fmt"
	"time"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	client := emulation.NewClient(100, 200*time.Millisecond)
	client.Start("localhost", "3333")
	defer client.Stop()

	ops := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial("localhost:9393", ops...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	rpcClient := emulation.NewEmulatorClient(conn)
	if _, err := rpcClient.OpenChannel(context.Background(),
		&emulation.OpenChannelRequest{
			UserId:       "1",
			LockedByUser: 10,
		}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	n, err := client.Channels()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	spew.Dump(n)
}
