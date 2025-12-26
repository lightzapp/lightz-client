package autoswap

import (
	"math"

	"github.com/lightzapp/lightz-client/pkg/lightz"
	"github.com/lightzapp/lightz-client/pkg/lightzrpc"

	"github.com/lightzapp/lightz-client/internal/utils"
)

type checks struct {
	Amount           uint64
	FeeEstimate      uint64
	DismissedReasons []string
}

const (
	ReasonMaxFeePercent     = "fee exceeds maximum percentage"
	ReasonAmountBelowMin    = "amount below minimal"
	ReasonBudgetExceeded    = "budget exceeded"
	ReasonPendingSwap       = "pending swap"
	ReasonFailedSwap        = "failed swap"
	ReasonInsufficientFunds = "insufficient funds"
)

func (recommendation *checks) Dismiss(reason string) {
	recommendation.DismissedReasons = append(recommendation.DismissedReasons, reason)
}

func (recommendation *checks) Dismissed() bool {
	return len(recommendation.DismissedReasons) > 0
}

type checkParams struct {
	Amount           uint64
	MaxFeePercent    lightz.Percentage
	Budget           *uint64
	Pair             *lightzrpc.PairInfo
	DismissedReasons []string
}

func check(amount uint64, params checkParams) checks {
	adjustedAmount := uint64(math.Min(float64(amount), float64(params.Pair.Limits.Maximal)))
	checks := checks{
		Amount:           adjustedAmount,
		DismissedReasons: params.DismissedReasons,
		FeeEstimate:      utils.CalculateFeeEstimate(params.Pair.Fees, adjustedAmount),
	}

	if checks.Amount < params.Pair.Limits.Minimal {
		checks.Dismiss(ReasonAmountBelowMin)
	}

	maxFee := lightz.CalculatePercentage(params.MaxFeePercent, adjustedAmount)
	if checks.FeeEstimate > maxFee {
		checks.Dismiss(ReasonMaxFeePercent)
	}

	if params.Budget != nil {
		if checks.FeeEstimate > *params.Budget {
			checks.Dismiss(ReasonBudgetExceeded)
		} else {
			*params.Budget -= checks.FeeEstimate
		}
	}
	return checks
}
