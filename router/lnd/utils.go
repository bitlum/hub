package lnd

import (
	"time"
	"context"
)

// panicRecovering is needed to ensure that our program not stops because if
// the panic, also this is needed to be bale properly send, alert to the metric
// server, because if metric server will be unable to scrape the metric than
// we wouldn't be able to see that on service radar.
func panicRecovering() {
	if r := recover(); r != nil {
		log.Error(r)
	}
}

func getContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*25)
	return ctx
}
