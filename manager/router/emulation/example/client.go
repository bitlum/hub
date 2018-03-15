package main

import (
	"github.com/bitlum/hub/manager/router/emulation"
	"google.golang.org/grpc"
	"os"
	"context"
	"fmt"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	r := emulation.NewRouter(100)
	r.Start("localhost", "3333")
	defer r.Stop()

	ops := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial("localhost:3333", ops...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	client := emulation.NewEmulatorClient(conn)
	if _, err := client.OpenChannel(context.Background(),
		&emulation.OpenChannelRequest{
			UserId:       1,
			LockedByUser: 10,
			ChanId:       1,
		}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	n, err := r.Network()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	spew.Dump(n)
}
