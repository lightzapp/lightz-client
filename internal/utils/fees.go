package utils

import (
	"github.com/lightzapp/lightz-client/pkg/lightz"
	"github.com/lightzapp/lightz-client/pkg/lightzrpc"
)

func CalculateFeeEstimate(fees *lightzrpc.SwapFees, amount uint64) uint64 {
	serviceFee := lightz.Percentage(fees.Percentage).Calculate(amount)
	return serviceFee + fees.MinerFees
}
