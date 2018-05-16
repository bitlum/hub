package emulation

// calculateForwardingFee calculates the number of funds taken by the router
// from payment amount in exchange for forwarding service.
func calculateForwardingFee(amount, feeBase, feeProportion int64) int64 {
	amtMs := toMilli(amount)
	feeMs := feeBase + (amtMs*feeProportion)/1000000
	return fromMilli(feeMs)
}
