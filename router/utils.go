package router

import (
	"github.com/bitlum/hub/lightning"
	"math/rand"
)

func generateRandomPaymentHash() lightning.PaymentHash {
	return lightning.PaymentHash(randStringBytes(32))
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
